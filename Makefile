# Makefile for go-clean-boilerplate

.PHONY: test test-order test-user test-invoice

test:
	go test ./...

compose:
	docker-compose down
	rm -rf pgdata
	docker-compose up -d --build

db-login:
	docker exec -it postgresdb psql -U mock -d mockdb

get-jwt:
	go run ./scriptsget-jwt.go