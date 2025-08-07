# Go Clean Architecture API

A REST API built with Go using Clean Architecture principles, featuring JWT authentication, PostgreSQL database, and Kafka integration.

## Features

- **Clean Architecture**: Separation of concerns with domain, use case, repository, and delivery layers
- **JWT Authentication**: Secure user authentication and authorization
- **PostgreSQL Database**: Relational database with proper migrations
- **Kafka Integration**: Event streaming for invoice operations
- **Docker Support**: Containerized application with docker-compose
- **RESTful API**: Complete CRUD operations for Products, Orders, Users, and Invoices
- **Comprehensive Testing**: API tests with VS Code REST Client

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- VS Code with REST Client extension (for API testing)

### Setup

1. **Clone and setup the project:**
   ```bash
   git clone <repository-url>
   cd go-clean-boilerplate
   ```

2. **Start the infrastructure:**
   ```bash
   # Start all services including the application
   docker-compose up -d
   
   # Or start individual services
   docker-compose up -d postgres kafka1  # Start dependencies
   docker-compose up -d go-clean-api     # Start API service
   docker-compose up -d go-clean-consumer # Start consumer service
   ```

3. **Run the application:**
   ```bash
   # Run the API server
   go run cmd/main.go -service=api
   
   # Run the Kafka consumer
   go run cmd/main.go -service=consumer
   ```

4. **Test the API:**
   Open `api.http` in VS Code and run the tests sequentially.

## API Testing

The `api.http` file contains comprehensive API tests that you can run directly in VS Code with the REST Client extension.

### Test Sequence

1. **Health Check** - Verify the server is running
2. **Authentication** - Register and login to get JWT token
3. **Products** - Create, read, update, delete products
4. **Orders** - Create orders with multiple items
5. **Invoices** - Create manual invoices and invoices from orders
6. **Status Updates** - Update order and invoice statuses
7. **Error Cases** - Test authentication and validation errors

### Running Tests

1. Make sure the server is running: `go run cmd/main.go -service=api`
2. Open `api.http` in VS Code
3. Click "Send Request" on each test, starting from the top
4. The tests use variables that are automatically populated from responses

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `GET /api/v1/auth/profile` - Get user profile (protected)

### Products
- `POST /api/v1/products` - Create product (protected)
- `GET /api/v1/products` - Get all products (protected)
- `GET /api/v1/products/:id` - Get product by ID (protected)
- `PUT /api/v1/products/:id` - Update product (protected)
- `DELETE /api/v1/products/:id` - Delete product (protected)

### Orders
- `POST /api/v1/orders` - Create order (protected)
- `GET /api/v1/orders` - Get all orders (protected)
- `GET /api/v1/orders/:id` - Get order by ID (protected)
- `PATCH /api/v1/orders/:id/status` - Update order status (protected)
- `DELETE /api/v1/orders/:id` - Delete order (protected)

### Invoices
- `POST /api/v1/invoices` - Create manual invoice (protected)
- `POST /api/v1/invoices/from-order/:order_id` - Create invoice from order (protected)
- `GET /api/v1/invoices` - Get all invoices (protected)
- `GET /api/v1/invoices/:id` - Get invoice by ID (protected)
- `GET /api/v1/invoices/number/:invoice_number` - Get invoice by number (protected)
- `GET /api/v1/invoices/overdue` - Get overdue invoices (protected)
- `PATCH /api/v1/invoices/:id/status` - Update invoice status (protected)
- `DELETE /api/v1/invoices/:id` - Delete invoice (protected)

## Architecture

```
├── cmd/
│   └── main.go              # Application entry point for both API and consumer
├── config/                  # Configuration management
├── internal/
│   ├── domain/             # Domain entities and interfaces
│   ├── usecase/            # Business logic layer
│   ├── repository/         # Data access layer
│   ├── delivery/
│   │   ├── http/           # HTTP handlers and routing
│   │   └── consumer/       # Kafka consumer handlers
├── pkg/
│   ├── database/           # Database connection utilities
│   ├── kafka/              # Kafka producer/consumer utilities
│   └── middleware/         # HTTP middleware
├── migrations/             # Database migrations
├── docker-compose.yml      # Docker services configuration
├── api.http               # API test collection
└── README.md              # This file
```

## Environment Variables

Create a `.env` file with the following variables:

```env
# Database
DB_HOST=localhost
DB_PORT=7775
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=go_clean_db
DB_SSLMODE=disable

# App
SERVER_PORT=8085
SERVER_HOST=0.0.0.0
JWT_SECRET=mockjwtsecret

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=invoice-topic
KAFKA_GROUP_ID=invoice-group
KAFKA_READ_TIMEOUT=0
KAFKA_MIN_BYTES=1
KAFKA_MAX_BYTES=1048576
KAFKA_MAX_WAIT=1
KAFKA_COMMIT_INTERVAL=100
KAFKA_QUEUE_CAPACITY=1000
KAFKA_READ_LAG_INTERVAL=0
KAFKA_WATCH_PARTITION_CHANGES=false
KAFKA_PARTITION_WATCH_INTERVAL=0
# Uncomment if using SASL/PLAIN authentication (e.g. Confluent Cloud)
# KAFKA_USERNAME=your_kafka_username
# KAFKA_PASSWORD=your_kafka_password
```

## Database Migrations

The application uses SQL migrations located in the `migrations/` directory:

- `001_create_products_table` - Products table
- `002_create_orders_table` - Orders and order items tables
- `003_create_users_table` - Users table for authentication
- `004_create_invoices_table` - Invoices and invoice items tables

## Kafka Integration

The application publishes invoice events to Kafka when:
- Creating new invoices
- Updating invoice status
- Creating invoices from orders

Events are published to the `invoice-topic` topic with detailed invoice information.

### Kafka Consumer

The application also includes a Kafka consumer that processes invoice events:

```bash
go run cmd/main.go -service=consumer
```

The consumer listens for invoice events and logs them to the console.

### Kafka Invoice Event Example

When producing invoice item logs to Kafka, use the following JSON payload format:

```json
{
  "invoice_id": 12345,
  "product_id": 67890,
  "description": "สินค้า A",
  "quantity": 2,
  "unit_price": 150.0,
  "total_price": 300.0
}
{"invoice_id": 12345,"product_id": 67890,"description": "สินค้า A","quantity": 2,"unit_price": 150.0,"total_price": 300.0}
```

This payload matches the structure of `InvoiceKafka` and is used for logging invoice items via the consumer service.

---

## Usage Examples

### 1. Register and Login
```http
POST http://localhost:8085/api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

### 2. Create Product
```http
POST http://localhost:8085/api/v1/products
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN

{
  "name": "iPhone 15 Pro",
  "description": "Latest iPhone",
  "price": 999.99,
  "stock": 50
}
```

### 3. Create Order
```http
POST http://localhost:8085/api/v1/orders
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN

{
  "customer_name": "John Doe",
  "customer_email": "john@example.com",
  "customer_phone": "+1234567890",
  "shipping_address": "123 Main St",
  "items": [
  {
    "product_id": "PRODUCT_UUID",
    "quantity": 2,
    "unit_price": 999.99
  }
]
}
```

### 4. Create Invoice from Order
```http
POST http://localhost:8085/api/v1/invoices/from-order/ORDER_UUID
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN

{
  "due_date": "2024-09-01T00:00:00Z",
  "notes": "Invoice for order"
}
```

## Development

### Adding New Features

1. **Domain Layer**: Add entities and interfaces in `internal/domain/`
2. **Use Case Layer**: Implement business logic in `internal/usecase/`
3. **Repository Layer**: Add data access in `internal/repository/`
4. **Delivery Layer**: Add HTTP handlers in `internal/delivery/http/`
5. **Migrations**: Add database changes in `migrations/`

### Code Structure

The project follows Clean Architecture principles:

- **Domain**: Core business entities and rules
- **Use Cases**: Application-specific business rules
- **Interface Adapters**: Controllers, presenters, and gateways
- **Frameworks & Drivers**: Web frameworks, databases, external services

## Troubleshooting

### Common Issues

1. **Database Connection Error**: Ensure PostgreSQL is running via docker-compose
2. **Kafka Connection Error**: Ensure Kafka and Zookeeper are running
3. **JWT Token Invalid**: Make sure to use the token from login response
4. **Port Already in Use**: Check if port 8085 is available

### Logs

The application provides detailed logging for:
- HTTP requests and responses
- Database operations
- Kafka message publishing
- Authentication events

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests in `api.http`
5. Submit a pull request

## License

This project is licensed under the MIT License.