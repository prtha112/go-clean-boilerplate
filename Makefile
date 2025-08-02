# Makefile for go-clean-boilerplate

.PHONY: test compose db-login get-jwt

test:
	go test ./internal/usecase/...

compose:
	docker-compose down
	rm -rf pgdata
	docker-compose up -d --build

db-login:
	docker exec -it postgresdb psql -U mock -d mockdb

get-jwt:
	go run ./scriptsget-jwt.go