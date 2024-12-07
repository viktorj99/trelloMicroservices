package client

import (
	"context"
	"log"
	userpb "pb/userpb"

	"google.golang.org/grpc"
)

type UserClient struct {
	Client userpb.UserServiceClient
	Conn   *grpc.ClientConn
}

func NewUserClient(address string) (*UserClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to user-service: %v", err)
		return nil, err
	}
	client := userpb.NewUserServiceClient(conn)
	return &UserClient{Client: client, Conn: conn}, nil
}

func (uc *UserClient) Close() error {
	return uc.Conn.Close()
}

func (uc *UserClient) CheckIfUserExists(ctx context.Context, userID string) (bool, error) {
	req := &userpb.CheckUserExistsByIDRequest{UserId: userID}
	resp, err := uc.Client.CheckUserExistsByID(ctx, req)
	if err != nil {
		log.Printf("Error calling CheckIfUserExists: %v", err)
		return false, err
	}
	return resp.Exists, nil
}
