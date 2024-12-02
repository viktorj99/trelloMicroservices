package helpers

import (
	"context"
	"log"
	"net"
	"os"
	projectpb "pb/projectpb"
	"projects-service/repositories"
	"projects-service/server"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func StartGRPCServer() {
	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)

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

	projectRepo, err := repositories.NewProjectRepo(ctx, logger, "project")
	if err != nil {
		logger.Fatal("Failed to create project repository: ", err)
	}

	// Set up gRPC listener
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Create and register the project server
	projectServer := server.NewProjectServer(projectRepo)
	projectpb.RegisterProjectServiceServer(grpcServer, projectServer)

	// Start the gRPC server in a goroutine
	go func() {
		logger.Println("gRPC server is running on port 50051...")
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("failed to serve: %v", err)
		}
	}()
}
