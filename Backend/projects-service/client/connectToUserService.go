package client

import "log"

func ConnectToUserService(logger *log.Logger) *UserClient {
	userServiceAddress := "localhost:50051"
	userClient, err := NewUserClient(userServiceAddress)
	if err != nil {
		logger.Fatal("Failed to connect to user-service: ", err)
	}
	return userClient
}