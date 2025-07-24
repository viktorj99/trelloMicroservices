# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY users-service/go.mod users-service/go.sum ./

COPY pb ./pb

COPY users-service/. .

COPY users-service/.env ./.env

RUN go clean -modcache

RUN ls -R /app

RUN go mod download

RUN go build -o users-service .

FROM alpine:latest

# Create a user and group to run the app (non-root)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Set ownership to the non-root user
RUN chown -R appuser:appgroup /app

COPY --from=builder /app/users-service /app/users-service
COPY --from=builder /app/.env /app/.env

EXPOSE 8080
EXPOSE 8443
EXPOSE 50051

# Switch to non-root user
USER appuser

# **Use relative path in ENTRYPOINT**
ENTRYPOINT ["./users-service"]
