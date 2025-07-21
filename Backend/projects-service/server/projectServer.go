package server

import (
	"context"
	"fmt"
	projectpb "pb/projectpb"
	"projects-service/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// ProjectServer struct implements the ProjectServiceServer interface from the proto
type ProjectServer struct {
	projectpb.UnimplementedProjectServiceServer
	projectRepo *repositories.ProjectRepo
}

func NewProjectServer(projectRepo *repositories.ProjectRepo) *ProjectServer {
	return &ProjectServer{
		projectRepo: projectRepo,
	}
}

func (s *ProjectServer) CheckMemberInProject(ctx context.Context, req *projectpb.MemberRequest) (*projectpb.BoolResponse, error) {
	tracer := otel.Tracer("projects-service/server")
	ctx, span := tracer.Start(ctx, "CheckMemberInProject")
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.method", "CheckMemberInProject"),
		attribute.String("project.id", req.GetProjectId()),
		attribute.String("member.id", req.GetMemberId()),
	)

	projectID, err := primitive.ObjectIDFromHex(req.GetProjectId())
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid project ID: %v", err)
	}

	memberID, err := primitive.ObjectIDFromHex(req.GetMemberId())
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid member ID: %v", err)
	}

	isMember, err := s.projectRepo.IsMemberInProject(ctx, projectID, memberID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("error checking if member is in project: %v", err)
	}

	span.SetAttributes(attribute.Bool("is_member", isMember))

	return &projectpb.BoolResponse{Value: isMember}, nil
}
