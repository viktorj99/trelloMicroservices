package helpers

// import (
// 	"log"
// 	"net"
// 	taskpb "pb/taskpb"
// 	"tasks-service/server"

// 	"google.golang.org/grpc"
// )

// func StartGRPCServer() {
// 	lis, err := net.Listen("tcp", ":50051")
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}

// 	grpcServer := grpc.NewServer()
// 	taskServer := server.NewTaskServer(taskRepo)
// 	taskpb.RegisterTaskServiceServer(grpcServer, taskServer)

// 	log.Println("gRPC server is running on port 50051...")
// 	if err := grpcServer.Serve(lis); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// }
