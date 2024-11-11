package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"users-service/services"
)

type VerificationRequest struct {
	Code string `json:"code"`
}

func VerifyCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req VerificationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Println(req.Code, req)

	// Get user ID by verification code
	username, err := services.GetUserUsernameByVerificationCode(req.Code)
	if err != nil {
		http.Error(w, "Invalid verification code", http.StatusBadRequest)
		return
	}

	log.Println(username)

	// Update the user to set isActive to true
	err = services.ActivateUser(username)
	if err != nil {
		http.Error(w, "Failed to activate user", http.StatusInternalServerError)
		return
	}

	// Optionally delete the verification code after activation
	services.DeleteVerificationCode(req.Code)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User activated successfully"})
}
