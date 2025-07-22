# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o workflow-service ./main.go

# Final minimal image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/workflow-service .

EXPOSE 8085

CMD ["./workflow-service"]