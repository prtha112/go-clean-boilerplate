test:
	go test ./internal/delivery/... -cover
	go test ./internal/repository/... -cover
	go test ./internal/usecase/... -cover