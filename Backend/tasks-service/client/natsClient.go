package client

import (
	"context"
	"log"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type NATSClient struct {
	Conn *nats.Conn
}

func NewNATSClient(natsURL string) *NATSClient {
	tracer := otel.Tracer("tasks-service/client")
	_, span := tracer.Start(context.Background(), "NewNATSClient")
	defer span.End()

	nc, err := nats.Connect(natsURL)
	if err != nil {
		span.RecordError(err)
		log.Fatalf("Error connecting to NATS: %v", err)
	}

	span.SetAttributes(attribute.String("nats.url", natsURL))

	return &NATSClient{
		Conn: nc,
	}
}
