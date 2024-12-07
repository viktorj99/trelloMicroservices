package handlers

import (
	"encoding/json"
	"net/http"
	"users-service/services"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func GetAllUserMembers(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("users-service").Start(r.Context(), "GetAllUserMembersHandler")
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		span.SetAttributes(attribute.String("http.method", r.Method))
		span.RecordError(http.ErrBodyNotAllowed)
		return
	}

	users, err := services.GetAllUserMembers(ctx)
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		span.RecordError(err)
		return
	}

	span.SetAttributes(attribute.Int("user.count", len(users)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
