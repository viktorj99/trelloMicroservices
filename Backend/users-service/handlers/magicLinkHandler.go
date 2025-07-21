package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"users-service/auth"
	"users-service/services"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func RequestMagicLinkHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("users-service").Start(r.Context(), "RequestMagicLinkHandler")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		span.SetAttributes(attribute.String("http.method", r.Method))
		span.RecordError(http.ErrBodyNotAllowed)
		return
	}

	var request struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		span.RecordError(err)
		return
	}

	email := request.Email
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		span.RecordError(fmt.Errorf("email is empty"))
		return
	}
	span.SetAttributes(attribute.String("user.email", email))

	user, err := services.GetUserByEmail(ctx, email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		span.RecordError(err)
		return
	}
	span.SetAttributes(attribute.String("user.id", user.ID), attribute.String("user.username", user.Username))

	token, err := services.GenerateJWTToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		span.RecordError(err)
		return
	}

	err = services.SendMagicLinkEmail(email, token)
	if err != nil {
		http.Error(w, "Failed to send magic link", http.StatusInternalServerError)
		span.RecordError(err)
		return
	}

	span.SetAttributes(attribute.String("magic_link.token", token))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Magic link sent to your email"})
}

func MagicLoginHandler(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer("users-service").Start(r.Context(), "MagicLoginHandler")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		span.SetAttributes(attribute.String("http.method", r.Method))
		span.RecordError(http.ErrBodyNotAllowed)
		return
	}

	claims, err := auth.ParseTokenBody(r)
	if err != nil {
		fmt.Println("Error parsing token:", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		span.RecordError(err)
		return
	}
	span.SetAttributes(attribute.String("user.id", claims.ID), attribute.String("user.username", claims.Username), attribute.String("user.role", claims.Role))

	sessionToken, err := services.GenerateJWTToken(claims.ID, claims.Username, claims.Role)
	if err != nil {
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		span.RecordError(err)
		return
	}

	span.SetAttributes(attribute.String("session.token", sessionToken))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"authToken": sessionToken})
}
