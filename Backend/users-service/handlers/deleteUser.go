package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"users-service/services"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func DeleteUserById(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("users-service").Start(r.Context(), "DeleteUserByIdHandler")
	defer span.End()

	fmt.Println("url", r.URL.Path)
	parts := strings.Split(r.URL.Path, "/")
	fmt.Println("url", parts)
	if len(parts) < 3 {
		span.SetAttributes(attribute.String("error", "Invalid URL"))
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	userId := parts[3]
	if userId == "" {
		span.SetAttributes(attribute.String("error", "Missing user ID"))
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("user.id", userId))

	result, err := services.DeleteUser(ctx, userId)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int64("deletedCount", result.DeletedCount))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "User deleted successfully",
		"deletedCount": result.DeletedCount,
	})
}
