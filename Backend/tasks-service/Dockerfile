# Build stage
FROM golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY tasks-service/go.mod tasks-service/go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the code
COPY tasks-service/ ./

# Build the binary
RUN go build -o tasks-service .

# Production stage
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/tasks-service .

# Expose the port (adjust if necessary)
EXPOSE 8080
EXPOSE 8443

# Run the application
CMD ["./tasks-service"]
