module projects-service

go 1.23

require (
	github.com/nats-io/nats.go v1.37.0
	go.mongodb.org/mongo-driver v1.17.1
)

require (
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/grpc v1.68.0
	pb/userpb v0.0.0-00010101000000-000000000000
)

replace pb/userpb => ./pb/userpb
