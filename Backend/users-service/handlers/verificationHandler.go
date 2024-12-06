package handlers

import (
	"encoding/json"
	"net/http"
	"users-service/services"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type VerificationRequest struct {
	Code string `json:"code"`
}

func VerifyCode(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("users-service").Start(r.Context(), "VerifyCode")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		span.RecordError(http.ErrBodyNotAllowed)
		span.SetAttributes(attribute.String("http.method", r.Method))
		return
	}

	var req VerificationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		span.RecordError(err)
		return
	}
	span.SetAttributes(attribute.String("verification.code", req.Code))

	username, err := services.GetUserUsernameByVerificationCode(ctx, req.Code)
	if err != nil {
		http.Error(w, "Invalid verification code", http.StatusBadRequest)
		span.RecordError(err)
		return
	}
	span.SetAttributes(attribute.String("user.username", username))

	err = services.ActivateUser(ctx, username, req.Code)
	if err != nil {
		http.Error(w, "Failed to activate user", http.StatusInternalServerError)
		span.RecordError(err)
		return
	}

	err = services.DeleteVerificationCode(ctx, req.Code)
	if err != nil {
		http.Error(w, "Failed to delete verification code", http.StatusInternalServerError)
		span.RecordError(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User activated successfully"})
}
