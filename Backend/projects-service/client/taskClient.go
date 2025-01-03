package client

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	taskpb "pb/taskpb"
)

type TaskClient struct {
	client         taskpb.TaskServiceClient
	conn           *grpc.ClientConn
	circuitBreaker *gobreaker.CircuitBreaker
	retryAttempts  int
}

func NewTaskClient(address string) (*TaskClient, error) {
	// Create a connection to the TaskService gRPC server
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}

	client := taskpb.NewTaskServiceClient(conn)

	// Initialize circuit breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "TaskServiceCB",
		MaxRequests: 5,
		Interval:    time.Minute,
		Timeout:     10 * time.Second,
		OnStateChange: func(name string, from, to gobreaker.State) {
			fmt.Printf("Circuit breaker state changed: %s -> %s\n", from.String(), to.String())
		},
	})

	return &TaskClient{
		client:         client,
		conn:           conn,
		circuitBreaker: cb,
		retryAttempts:  3,
	}, nil
}

func (c *TaskClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *TaskClient) withRetry(ctx context.Context, call func() error) error {
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

func (c *TaskClient) GetUnassignedTasks(ctx context.Context, req *taskpb.ProjectRequest, opts ...grpc.CallOption) (*taskpb.TaskResponse, error) {
	var resp *taskpb.TaskResponse
	_, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // Explicit timeout
		defer cancel()

		var err error
		resp, err = c.client.GetUnassignedTasks(ctx, req, opts...)
		if err != nil {
			if status.Code(err) == codes.DeadlineExceeded {
				fmt.Println("Fallback: using default empty response")
				resp = &taskpb.TaskResponse{} // Default fallback response
				return resp, nil
			}
			return nil, err
		}
		return resp, nil
	})
	return resp, err
}

func (c *TaskClient) CheckMemberTasksInProgress(ctx context.Context, req *taskpb.MemberRequest, opts ...grpc.CallOption) (*taskpb.BoolResponse, error) {
	var resp *taskpb.BoolResponse
	_, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // Explicit timeout
		defer cancel()

		var err error
		resp, err = c.client.CheckMemberTasksInProgress(ctx, req, opts...)
		if err != nil {
			return nil, err
		}
		return resp, nil
	})
	return resp, err
}
