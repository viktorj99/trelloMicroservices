package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"users-service/auth"
	"users-service/services"
)

func RequestMagicLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	email := request.Email
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	user, err := services.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	token, err := services.GenerateJWTToken(user.Username, user.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	err = services.SendMagicLinkEmail(email, token)
	if err != nil {
		http.Error(w, "Failed to send magic link", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Magic link sent to your email"})
}

func MagicLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	claims, err := auth.ParseTokenBody(r)
	if err != nil {
		fmt.Println("Error parsing token:", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	sessionToken, err := services.GenerateJWTToken(claims.Username, claims.Role)
	if err != nil {
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"authToken": sessionToken})
}
