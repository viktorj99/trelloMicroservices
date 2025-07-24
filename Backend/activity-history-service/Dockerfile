# Use the official Go image for building
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY activity-history-service/go.mod activity-history-service/go.sum ./

# Clean up Go module cache (optional)
RUN go clean -modcache

# Download Go module dependencies
RUN go mod download

# Copy the entire activity-history-service directory
COPY activity-history-service/. .

# Build the Go binary for the activity-history-service
RUN go build -o activity-history-service .

# Create a smaller image for production
FROM alpine:latest

# Set working directory in the final container
WORKDIR /app/

# Create the group and user (must be done before chown and USER)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the compiled binary from the builder stage
COPY --from=builder /app/activity-history-service .

# Set ownership of /app directory to appuser:appgroup
RUN chown -R appuser:appgroup /app

# Expose the application port (adjust if necessary)
EXPOSE 8080
EXPOSE 8443

# Switch to the non-root user
USER appuser

# Set the entry point for the container
CMD ["./activity-history-service"]