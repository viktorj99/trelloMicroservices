package main

import (
	"context"
	"log"
	"os"
	"users-service/helpers"
	"users-service/repositories"
	"users-service/services"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	// err := utils.LoadCommonPasswordsToConsul("/app/config/common_passwords.txt")
	// if err != nil {
	// 	log.Fatalf("Failed to load common passwords to Consul: %v", err)
	// }

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	shutdown := initTracing()
	defer shutdown()

	dbURI := os.Getenv("MONGO_DB_URI")
	if dbURI == "" {
		logger.Fatal("MONGO_DB_URI is not set")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(dbURI))
	if err != nil {
		logger.Fatal("Failed to parse MongoDB URI: ", err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB: ", err)
	}
	defer client.Disconnect(ctx)

	repositories.InitRepository(client)

	go services.StartKeyExpirationListener(ctx)

	go helpers.StartHTTPServer()
	helpers.StartGRPCServer()
}

func initTracing() func() {
	ctx := context.Background()

	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient())
	if err != nil {
		log.Fatalf("Failed to create OTLP exporter: %v", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("users-service"),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create resource: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	log.Println("OpenTelemetry tracing initialized")

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down tracer provider: %v", err)
		}
		log.Println("OpenTelemetry tracing shut down")
	}
}
