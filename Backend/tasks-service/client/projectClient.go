package client

import (
	"context"
	"fmt"

	projectpb "pb/projectpb"

	"google.golang.org/grpc"
)

type ProjectClient struct {
	client projectpb.ProjectServiceClient
	conn   *grpc.ClientConn
}

func NewProjectClient(address string) (*ProjectClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}

	client := projectpb.NewProjectServiceClient(conn)

	return &ProjectClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *ProjectClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *ProjectClient) CheckMemberInProject(ctx context.Context, projectID, memberID string, opts ...grpc.CallOption) (*projectpb.BoolResponse, error) {
	req := &projectpb.MemberRequest{
		ProjectId: projectID,
		MemberId:  memberID,
	}
	return c.client.CheckMemberInProject(ctx, req, opts...)
}
