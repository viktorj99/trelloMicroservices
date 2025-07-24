# Use the official Go image for building
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the binary
RUN go build -o api-gateway .

# Create a smaller image for production
FROM alpine:latest

WORKDIR /app

# Copy the binary file from the builder stage
COPY --from=builder /app/api-gateway .

# Expose the port (adjust if needed)
EXPOSE 8080
EXPOSE 8443

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Run the application
CMD ["./api-gateway"]
