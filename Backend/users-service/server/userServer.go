package server

import (
	"context"
	userpb "pb/userpb"
	"users-service/repositories"
	"users-service/services"
)

type UserServer struct {
	userpb.UnimplementedUserServiceServer
}

func NewUserServer() *UserServer {
	return &UserServer{}
}

func (s *UserServer) GetAllUsers(ctx context.Context, req *userpb.GetAllUsersRequest) (*userpb.GetAllUsersResponse, error) {
	usersFromDB, err := services.GetAllUsers()
	if err != nil {
		return nil, err
	}

	// Mapiranje korisnika iz baze na gRPC strukturu
	var users []*userpb.User
	for _, user := range usersFromDB {
		users = append(users, &userpb.User{
			Id:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Username:  user.Username,
			Role:      user.Role,
		})
	}

	return &userpb.GetAllUsersResponse{Users: users}, nil
}

func (s *UserServer) CheckUserExistsByID(ctx context.Context, req *userpb.CheckUserExistsByIDRequest) (*userpb.CheckUserExistsByIDResponse, error) {
	userID := req.GetUserId()

	exists, err := repositories.UserExistsByID(userID)
	if err != nil {
		return nil, err
	}

	return &userpb.CheckUserExistsByIDResponse{Exists: exists}, nil
}
