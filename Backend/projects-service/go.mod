module projects-service

go 1.23.3

require (
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/nats-io/nats.go v1.37.0
	go.mongodb.org/mongo-driver v1.17.1
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.32.0
	pb/projectpb v0.0.0-00010101000000-000000000000
	pb/taskpb v0.0.0-00010101000000-000000000000
	pb/userpb v0.0.0-00010101000000-000000000000
)

require (
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.32.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/sony/gobreaker v1.0.0
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/grpc v1.68.0
)

replace pb/userpb => ./pb/userpb

replace pb/taskpb => ./pb/taskpb

replace pb/projectpb => ./pb/projectpb
