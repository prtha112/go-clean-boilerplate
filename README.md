# go-clean-boilerplate

A simple Go project boilerplate using Clean Architecture principles.

## Features
- Layered structure: domain, usecase, infrastructure, interface
- Example user entity, repository, and use case
- Easy to extend for real-world applications

## Project Structure
```
go-clean-boilerplate/
├── cmd/                # Application entrypoint (main.go)
├── internal/
│   ├── domain/         # Business entities and repository interfaces
│   │   └── user/
│   ├── infrastructure/ # Implementations of repositories (e.g., DB)
│   │   └── user/
│   ├── interface/      # Delivery layer (HTTP handlers, routers)
│   │   └── http/
│   └── usecase/        # Application use cases
│       └── user/
├── go.mod
├── go.sum
└── README.md
```

## Getting Started
1. **Clone the repository**
   ```sh
   git clone <your-repo-url>
   cd go-clean-boilerplate
   ```

2. **Build Docker image**
   ```sh
   docker build -t go-clean-app .
   ```

3. **Run REST API (default port 8085)**
   ```sh
   docker run -e APP_MODE=restapi -p 8085:8085 go-clean-app
   ```

4. **Run Kafka Invoice Consumer**
   ```sh
   docker run -e APP_MODE=consume-invoice go-clean-app
   ```


## Example Invoice Kafka Message (JSON)

```json
{
  "id": "sojvp-001",
  "order_id": "order-123",
  "amount": 999.99,
  "created_at": 1722120000
}
```

## Running Tests

To run all unit tests:

```sh
# Run all tests in the project
 go test ./...
```

You can also run tests for a specific package, for example:

```sh
# Run only order usecase tests
 go test ./internal/usecase/order
```

## How to Extend
- Add new entities in `internal/domain/<entity>`
- Implement new repositories in `internal/infrastructure/<entity>`
- Add new use cases in `internal/usecase/<entity>`
- Add new handlers in `internal/interface/http`

## License
MIT