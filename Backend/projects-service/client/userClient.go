package client

import (
	"context"
	"fmt"
	"log"
	"time"

	userpb "pb/userpb"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserClient struct {
	Client         userpb.UserServiceClient
	Conn           *grpc.ClientConn
	CircuitBreaker *gobreaker.CircuitBreaker
	RetryAttempts  int
}

func NewUserClient(address string) (*UserClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user-service: %v", err)
	}

	client := userpb.NewUserServiceClient(conn)

	// Initialize circuit breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "UserServiceCB",
		MaxRequests: 5,
		Interval:    time.Minute,
		Timeout:     10 * time.Second,
		OnStateChange: func(name string, from, to gobreaker.State) {
			fmt.Printf("Circuit breaker state changed: %s -> %s\n", from.String(), to.String())
		},
	})

	return &UserClient{
		Client:         client,
		Conn:           conn,
		CircuitBreaker: cb,
		RetryAttempts:  3,
	}, nil
}

func (uc *UserClient) Close() {
	if uc.Conn != nil {
		uc.Conn.Close()
	}
}

func (uc *UserClient) withRetry(ctx context.Context, call func() error) error {
	var err error
	for attempt := 0; attempt < uc.RetryAttempts; attempt++ {
		err = call()
		if err == nil || status.Code(err) == codes.Canceled || status.Code(err) == codes.DeadlineExceeded {
			break
		}
		fmt.Printf("Retrying after error: %v (attempt %d)\n", err, attempt+1)
		time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
	}
	return err
}

func (uc *UserClient) CheckIfUserExists(ctx context.Context, userID string) (bool, error) {
	var resp *userpb.CheckUserExistsByIDResponse

	_, err := uc.CircuitBreaker.Execute(func() (interface{}, error) {
		// Retry wrapper
		err := uc.withRetry(ctx, func() error {
			reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			req := &userpb.CheckUserExistsByIDRequest{UserId: userID}
			var err error
			resp, err = uc.Client.CheckUserExistsByID(reqCtx, req)
			if err != nil && status.Code(err) == codes.DeadlineExceeded {
				fmt.Println("Fallback: treating as user does not exist")
				resp = &userpb.CheckUserExistsByIDResponse{Exists: false}
				return nil // fallback triggers successful return
			}
			return err
		})
		if err != nil {
			return nil, err
		}
		return resp, nil
	})

	if err != nil {
		log.Printf("Error in CheckIfUserExists: %v", err)
		return false, err
	}
	return resp.Exists, nil
}
