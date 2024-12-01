package client

import (
	"context"
	"fmt"

	taskpb "pb/taskpb"

	"google.golang.org/grpc"
)

// TaskServiceClient is the client that connects to TaskService
type TaskClient struct {
	client taskpb.TaskServiceClient
	conn   *grpc.ClientConn
}

// NewTaskServiceClient creates a new TaskServiceClient
func NewTaskClient(address string) (*TaskClient, error) {
	// Create a connection to the TaskService gRPC server
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}

	client := taskpb.NewTaskServiceClient(conn)

	return &TaskClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the connection to the TaskService
func (c *TaskClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// GetUnassignedTasks fetches unassigned tasks for a project
func (c *TaskClient) GetUnassignedTasks(ctx context.Context, req *taskpb.ProjectRequest, opts ...grpc.CallOption) (*taskpb.TaskResponse, error) {
	return c.client.GetUnassignedTasks(ctx, req, opts...)
}

// CheckMemberTasksInProgress checks if a member has tasks in progress
func (c *TaskClient) CheckMemberTasksInProgress(ctx context.Context, req *taskpb.MemberRequest, opts ...grpc.CallOption) (*taskpb.BoolResponse, error) {
	return c.client.CheckMemberTasksInProgress(ctx, req, opts...)
}
