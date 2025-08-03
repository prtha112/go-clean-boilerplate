# Makefile for go-clean-boilerplate

.PHONY: test compose-github-action compose-local db-login get-jwt

test:
	go test ./internal/usecase/...

compose-github-action:
	docker compose down
	rm -rf pgdata
	docker compose up -d --build
	kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 3 --topic invoice-topic
	sleep 10
	go run cmd/main.go restapi &

compose-local:
	docker compose down
	rm -rf pgdata
	docker compose up -d --build
	kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 3 --topic invoice-topic

db-login:
	docker exec -it postgresdb psql -U mock -d mockdb