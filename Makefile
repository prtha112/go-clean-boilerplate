# Makefile for go-clean-boilerplate

.PHONY: test compose db-login get-jwt

test:
	go test ./internal/usecase/...

compose-github-action:
	docker compose down
	rm -rf pgdata
	docker compose up -d --build
	sleep 10
	go run cmd/main.go restapi &

compose-local:
	docker compose down
	rm -rf pgdata
	docker compose up -d --build

db-login:
	docker exec -it postgresdb psql -U mock -d mockdb

get-jwt:
	go run ./scriptsget-jwt.go