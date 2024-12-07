package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	userpb "pb/userpb"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request, userClient userpb.UserServiceClient) {
	ctx, span := otel.Tracer("projects-service").Start(r.Context(), "GetUsersHandler")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	response, err := userClient.GetAllUsers(ctx, &userpb.GetAllUsersRequest{})
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.message", err.Error()))
		span.SetStatus(codes.Error, "Failed to fetch users")

		http.Error(w, fmt.Sprintf("Failed to fetch users: %v", err), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("users.count", len(response.Users)))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response.Users); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to encode response")
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "Success")
}
