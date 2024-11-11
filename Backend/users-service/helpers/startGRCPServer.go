package helpers

import (
	"log"
	"net"
	userpb "pb/userpb"
	"users-service/server"

	"google.golang.org/grpc"
)

func StartGRPCServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	userServer := server.NewUserServer() 
	userpb.RegisterUserServiceServer(grpcServer, userServer) 

	log.Println("gRPC server is running on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}