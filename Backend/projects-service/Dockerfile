# Use the official Go image for building
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum
COPY projects-service/go.mod projects-service/go.sum ./

# Copy the pb directory
COPY pb ./pb

# Copy the rest of the project-service code
COPY projects-service/. .

# Clean up Go module cache (optional)
RUN go clean -modcache

# Download dependencies
RUN go mod download

# Build the binary
RUN go build -o projects-service .

# Create a smaller image for production
FROM alpine:latest

WORKDIR /root/

# Copy the binary file from the previous phase
COPY --from=builder /app/projects-service .

# Expose the port (adjust if needed)
EXPOSE 8080

# Run the application
CMD ["./projects-service"]
