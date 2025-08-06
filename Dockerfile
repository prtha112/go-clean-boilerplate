# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# Default to building the API service
ARG SERVICE=api
RUN if [ "$SERVICE" = "consumer" ]; then \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd && \
        echo "consumer" > /app/service_type; \
    else \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd && \
        echo "api" > /app/service_type; \
    fi

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migration files
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the binary with the appropriate service flag
CMD ["sh", "-c", "if [ -f /app/service_type ]; then SERVICE=$(cat /app/service_type); ./main -service=$SERVICE; else ./main -service=api; fi"]