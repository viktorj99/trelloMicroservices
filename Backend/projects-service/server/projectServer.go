package server

import (
	"context"
	"fmt"
	projectpb "pb/projectpb"
	"projects-service/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProjectServer struct implements the ProjectServiceServer interface from the proto
type ProjectServer struct {
	projectpb.UnimplementedProjectServiceServer
	projectRepo *repositories.ProjectRepo
}

// NewProjectServer creates a new instance of ProjectServer
func NewProjectServer(projectRepo *repositories.ProjectRepo) *ProjectServer {
	return &ProjectServer{
		projectRepo: projectRepo,
	}
}

// CheckMemberInProject implements the CheckMemberInProject gRPC method
func (s *ProjectServer) CheckMemberInProject(ctx context.Context, req *projectpb.MemberRequest) (*projectpb.BoolResponse, error) {
	projectID, err := primitive.ObjectIDFromHex(req.GetProjectId())
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %v", err)
	}

	memberID, err := primitive.ObjectIDFromHex(req.GetMemberId())
	if err != nil {
		return nil, fmt.Errorf("invalid member ID: %v", err)
	}

	isMember, err := s.projectRepo.IsMemberInProject(ctx, projectID, memberID)
	if err != nil {
		return nil, fmt.Errorf("error checking if member is in project: %v", err)
	}

	return &projectpb.BoolResponse{Value: isMember}, nil
}
