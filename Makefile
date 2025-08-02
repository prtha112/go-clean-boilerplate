# Makefile for go-clean-boilerplate

.PHONY: test test-order test-user test-invoice

test:
	go test ./...

test-order:
	go test ./internal/usecase/order

test-user:
	go test ./internal/usecase/user

test-invoice:
	go test ./internal/usecase/invoice
