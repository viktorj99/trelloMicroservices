# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum from users-service
COPY users-service/go.mod users-service/go.sum ./

# Copy the pb directory
COPY pb ./pb

# Copy the rest of the application code
COPY users-service/. .

# Clean up Go module cache (optional)
RUN go clean -modcache

# Debugging: Verify directory structure
RUN ls -R /app

# Download dependencies
RUN go mod download

# Build the application
RUN go build -o users-service .

# Stage 2: Create a minimal image for running the application
FROM alpine:latest

COPY --from=builder /app/users-service /usr/local/bin/users-service

EXPOSE 8080
EXPOSE 50051

ENTRYPOINT ["users-service"]
