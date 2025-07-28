# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/app ./app
EXPOSE 8085

# Default to restapi if not specified
ARG APP_MODE=restapi
ENV APP_MODE=${APP_MODE}
ENTRYPOINT ["/bin/sh", "-c", "./app $APP_MODE"]
