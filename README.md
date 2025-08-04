# Go Clean Architecture API

A REST API built with Go using Clean Architecture principles, featuring JWT authentication, PostgreSQL database, and Kafka integration.

## Features

- **Clean Architecture**: Separation of concerns with domain, use case, repository, and delivery layers
- **JWT Authentication**: Secure user authentication and authorization
- **PostgreSQL Database**: Relational database with proper migrations
- **Kafka Integration**: Event streaming for invoice operations
- **Docker Support**: Containerized application with docker-compose
- **RESTful API**: Complete CRUD operations for Products, Orders, Users, and Invoices

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- VS Code with REST Client extension (for API testing)

### Setup

1. **Clone and setup the project:**
   ```bash
   git clone <repository-url>
   cd go-clean-v2
   ```

2. **Start the infrastructure:**
   ```bash
   docker-compose up -d
   ```

3. **Run the application:**
   ```bash
   go run cmd/api/main.go
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

1. Make sure the server is running: `go run cmd/api/main.go`
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
├── cmd/api/                 # Application entry point
├── config/                  # Configuration management
├── internal/
│   ├── domain/             # Domain entities and interfaces
│   ├── usecase/            # Business logic layer
│   ├── repository/         # Data access layer
│   └── delivery/http/      # HTTP handlers and routing
├── pkg/
│   ├── database/           # Database connection utilities
│   ├── kafka/              # Kafka producer utilities
│   └── middleware/         # HTTP middleware
├── migrations/             # Database migrations
├── docker-compose.yml      # Docker services configuration
├── api.http               # API test collection
└── README.md              # This file
```

## Environment Variables

Create a `.env` file with the following variables:

```env
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_clean_db
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here

# Kafka Configuration
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=invoice-events
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

Events are published to the `invoice-events` topic with detailed invoice information.

## Testing Examples

### 1. Register and Login
```http
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

### 2. Create Product
```http
POST http://localhost:8080/api/v1/products
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
POST http://localhost:8080/api/v1/orders
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
4. **Port Already in Use**: Check if port 8080 is available

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