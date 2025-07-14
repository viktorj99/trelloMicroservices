package helpers

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	projectpb "pb/projectpb"
	"projects-service/repositories"
	"projects-service/server"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

func StartGRPCServer() {
	tracer := otel.Tracer("projects-service")
	ctx, span := tracer.Start(context.Background(), "StartGRPCServer")
	defer span.End()

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

	dbURI := os.Getenv("MONGO_DB_URI")
	if dbURI == "" {
		err := fmt.Errorf("MONGO_DB_URI is not set")
		span.RecordError(err)
		logger.Fatal(err)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(dbURI))
	if err != nil {
		span.RecordError(err)
		logger.Fatal("Failed to parse MongoDB URI: ", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		span.RecordError(err)
		logger.Fatal("Failed to connect to MongoDB: ", err)
	}
	defer client.Disconnect(ctx)

	projectRepo, err := repositories.NewProjectRepo(ctx, logger, "project")
	if err != nil {
		span.RecordError(err)
		logger.Fatal("Failed to create project repository: ", err)
	}

	span.SetAttributes(attribute.String("db.uri", dbURI))

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		span.RecordError(err)
		logger.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	projectServer := server.NewProjectServer(projectRepo)
	projectpb.RegisterProjectServiceServer(grpcServer, projectServer)

	go func() {
		logger.Println("gRPC server is running on port 50051...")
		if err := grpcServer.Serve(lis); err != nil {
			span.RecordError(err)
			logger.Fatalf("failed to serve: %v", err)
		}
	}()
}
