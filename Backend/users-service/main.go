package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"users-service/database"
	"users-service/handlers"
	"users-service/repositories"

	userpb "pb/userpb"

	"google.golang.org/grpc"
)

// server struktura implementira UserService gRPC interfejs
type server struct {
	userpb.UnimplementedUserServiceServer
}

// Implementacija GetAllUsers RPC metode za gRPC
func (s *server) GetAllUsers(ctx context.Context, req *userpb.GetAllUsersRequest) (*userpb.GetAllUsersResponse, error) {
	// Koristimo funkciju repositories.GetAllUsers za dohvatanje korisnika iz baze
	usersFromDB, err := repositories.GetAllUsers()
	if err != nil {
		return nil, err
	}

	// Mapiranje korisnika iz baze na gRPC strukturu
	var users []*userpb.User
	for _, user := range usersFromDB {
		users = append(users, &userpb.User{
			Id:       user.ID,
			FirstName:     user.FirstName ,
			LastName:     user.LastName,
			Email:    user.Email,
			Username: user.Username,
			Role:     user.Role,
		})
	}

	return &userpb.GetAllUsersResponse{Users: users}, nil
}

func enableCORS(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handlerFunc(w, r)
	}
}

// Pokretanje HTTP servera za tradicionalne REST rute
func startHTTPServer() {
	http.HandleFunc("/register", enableCORS(handlers.RegisterUser))
	http.HandleFunc("/users", enableCORS(handlers.GetAllUsers))
	http.HandleFunc("/users/", enableCORS(handlers.GetUserByID))
	log.Println("HTTP server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Pokretanje gRPC servera
func startGRPCServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	userpb.RegisterUserServiceServer(s, &server{})
	log.Println("gRPC server is running on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	client, err := database.GetMongoClient()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	repositories.InitRepository(client)

	// Pokretanje HTTP i gRPC servera paralelno
	go startHTTPServer() // HTTP server na portu 8080
	startGRPCServer()    // gRPC server na portu 50051
}
