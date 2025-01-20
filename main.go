package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
)

var ordersJSON = `
{
  {
    "order_id": 101,
    "customer_id": 1,
    "skus": [1001, 1002, 1003],
    "created_at": "2025-01-20T10:30:00Z"
  },
  {
    "order_id": 102,
    "customer_id": 2,
    "skus": [2001, 2002],
    "created_at": "2025-01-20T11:15:00Z"
  }
}`

const (
	DBHost  = "127.0.0.1"
	DBPort  = 5432
	DBUser  = "root"
	DBPass  = "p@ssword"
	DBName  = "service1"
	AppPort = ":8080"
)

type Order struct {
	OrderId    int     `json:"order_id"`
	CustomerId int     `json:"customer_id"`
	SKUs       []int64 `json:"skus"`
	CreatedAt  string  `json:"created_at"`
}

var (
	db       *sql.DB
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func connectToDatabase() error {
	pgConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DBHost, DBPort, DBUser, DBPass, DBName)

	var err error

	db, err = sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v\n", err)
		return err
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to establish connection to PostgreSQL: %v\n", err)
		return err
	}

	fmt.Println("Successfully connected to the database.")
	return nil
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%T", r)
	vars := mux.Vars(r)
	order_ID := vars["order_id"]
	fmt.Println(order_ID)
	var o Order
	var skus pq.Int64Array
	err := db.QueryRow("SELECT * FROM orders WHERE order_id = $1", order_ID).Scan(&o.OrderId, &o.CustomerId, &skus, &o.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch order details", http.StatusInternalServerError)
		log.Printf("Error fetching orders: %v", err)
		return
	}
	o.SKUs = skus

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

// func getBulkOrder() {

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	conn.WriteMessage(websocket.TextMessage, []byte("Order processing started"))
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var o Order
	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		log.Println("Error decoding JSON:", err)
		return
	}
	if !validateCAndP(o.CustomerId, o.SKUs) {
		http.Error(w, "Invalid customer or product", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO orders (order_id, customer_id, skus, created_at) VALUES ($1, $2, $3, NOW())", o.OrderId, o.CustomerId, pq.Array(o.SKUs))
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		log.Println("Error inserting order:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":         "Order created successfully",
		"order_id":        o.OrderId,
		"database result": result,
	})
}

func validateCAndP(customerId int, skus []int64) bool {

	customerURL := fmt.Sprintf("http://localhost:8082/customer/%d", customerId)
	Val_cus, err := http.Get(customerURL)
	if err != nil {
		log.Printf("Failed to validate customer(%d): %v", customerId, err)
		return false
	}
	defer Val_cus.Body.Close()

	for _, sku := range skus {
		productURL := fmt.Sprintf("http://localhost:8082/product/%d", sku)
		Val_sku, err := http.Get(productURL)
		if err != nil {
			log.Printf("Failed to validate SKU(%d): %v", sku, err)
			return false
		}
		defer Val_sku.Body.Close()
	}

	return true
}

func main() {
	if err := connectToDatabase(); err != nil {
		log.Fatal("Database connection failed")
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/create_order", createOrder).Methods("POST")
	r.HandleFunc("/get_order/{order_id}", getOrder).Methods("GET")
	r.HandleFunc("/ws", handleWebSocket)
	log.Printf("OMS is running on port %s", AppPort)
	log.Fatal(http.ListenAndServe(AppPort, r))

	// bulk order
	// var orders []Order
	//  err := json.Unmarshal(ordersJSON, &orders)
	// if err != nil {
	// 	log.Fatalf("Failed to parse JSON: %v", err)
}
