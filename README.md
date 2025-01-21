# Order Management System (OMS)

The **Order Management System (OMS)** is responsible for managing orders. It interacts with the **Product Management System (Service 2)** to validate customer IDs and product availability. The service runs on port `8080`.

---

## Functionalities

### 1. Get Order Details by Order ID
#### Description:
Fetches details of an order using its unique order ID.

#### Example Response:
![Get Order Details](https://github.com/user-attachments/assets/980e1563-6706-4758-98d4-30298821c772)

---

### 2. Create Order
#### Description:
Creates a new order and performs the following checks:
1. Validates the customer ID by calling **Service 2**.
2. Checks the availability of products (SKUs) in **Service 2**.

#### Outcomes:
##### a) Successfully Created Order
- The customer ID and SKUs are validated successfully.

**Request Example:**
![Successfully Created Order - Request](https://github.com/user-attachments/assets/2d40af5e-deb8-4253-8d5d-f963918af4b8)

**Response Example:**
![Successfully Created Order - Response](https://github.com/user-attachments/assets/849fa265-2ebc-4c21-8686-ffbe234e4902)

---

##### b) Order Not Created (Customer Not Found)
- The order creation fails because the customer ID is invalid.

**Request Example:**
![Order Not Created - Request](https://github.com/user-attachments/assets/ea380615-aa59-45c1-87ca-ab0b60b49d9e)

**Response Example:**
![Order Not Created - Response](https://github.com/user-attachments/assets/ab6e1785-3174-4d7e-9867-ab3bbb5c9bc1)

---


