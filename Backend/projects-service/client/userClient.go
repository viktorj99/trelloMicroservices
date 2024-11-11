package client

import (
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