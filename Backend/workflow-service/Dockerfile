FROM golang:1.20
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o workflow-service ./main.go
EXPOSE 8085
CMD ["./workflow-service"]
