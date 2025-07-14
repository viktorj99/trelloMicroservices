package client

import (
	"context"
	"fmt"
	"time"

	projectpb "pb/projectpb"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProjectClient struct {
	client         projectpb.ProjectServiceClient
	conn           *grpc.ClientConn
	circuitBreaker *gobreaker.CircuitBreaker
	retryAttempts  int
}

func NewProjectClient(address string) (*ProjectClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}

	client := projectpb.NewProjectServiceClient(conn)

	// Initialize circuit breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "ProjectServiceCB",
		MaxRequests: 5,
		Interval:    time.Minute,
		Timeout:     10 * time.Second,
		OnStateChange: func(name string, from, to gobreaker.State) {
			fmt.Printf("Circuit breaker state changed: %s -> %s\n", from.String(), to.String())
		},
	})

	return &ProjectClient{
		client:         client,
		conn:           conn,
		circuitBreaker: cb,
		retryAttempts:  3,
	}, nil
}

func (c *ProjectClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *ProjectClient) withRetry(ctx context.Context, call func() error) error {
	var err error
	for attempt := 0; attempt < c.retryAttempts; attempt++ {
		err = call()
		if err == nil || status.Code(err) == codes.Canceled || status.Code(err) == codes.DeadlineExceeded {
			break
		}
		fmt.Printf("Retrying after error: %v (attempt %d)\n", err, attempt+1)
		time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
	}
	return err
}

func (c *ProjectClient) CheckMemberInProject(ctx context.Context, projectID, memberID string, opts ...grpc.CallOption) (*projectpb.BoolResponse, error) {
	var resp *projectpb.BoolResponse

	_, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		// Retry wrapper
		err := c.withRetry(ctx, func() error {
			reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			req := &projectpb.MemberRequest{
				ProjectId: projectID,
				MemberId:  memberID,
			}

			var err error
			resp, err = c.client.CheckMemberInProject(reqCtx, req, opts...)
			if err != nil && status.Code(err) == codes.DeadlineExceeded {
				fmt.Println("Fallback: assuming member is not in project")
				resp = &projectpb.BoolResponse{Value: false}
				return nil
			}
			return err
		})
		if err != nil {
			return nil, err
		}
		return resp, nil
	})

	if err != nil {
		fmt.Printf("Error in CheckMemberInProject: %v\n", err)
		return nil, err
	}
	return resp, nil
}
