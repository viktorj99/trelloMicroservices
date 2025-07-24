# Use the official Go image for building
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY notifications-service/go.mod notifications-service/go.sum ./

# Clean up Go module cache (optional)
RUN go clean -modcache

# Download Go module dependencies
RUN go mod download

# Copy the entire notification-service directory
COPY notifications-service/. .

# Build the Go binary for the notification service
RUN go build -o notifications-service .

# Create a smaller image for production
FROM alpine:latest

# Set working directory in the final container
WORKDIR /app/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/notifications-service .

# Expose the application port (adjust if necessary)
EXPOSE 8080
EXPOSE 8443

# Create a non-root user
RUN adduser -D appuser
USER appuser

# Set the entry point for the container
CMD ["./notifications-service"]
