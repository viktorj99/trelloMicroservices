package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"users-service/auth"
	"users-service/services"
)

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
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

	user, err := services.GetUserByEmail(request.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	resetCode, err := services.GenerateVerificationCode()
	if err != nil {
		http.Error(w, "Failed to generate reset code", http.StatusInternalServerError)
		return
	}

	err = services.SaveVerificationCode(resetCode, user.Username)
	if err != nil {
		http.Error(w, "Failed to save reset code", http.StatusInternalServerError)
		return
	}

	err = services.SendPasswordResetEmail(user.Email, resetCode)
	if err != nil {
		log.Printf("Error while sending reset email: %v", err)
		http.Error(w, "Failed to send reset email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Reset code sent successfully"})
}

func ChangeForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Code        string `json:"code"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	username, err := services.GetUserUsernameByVerificationCode(request.Code)
	if err != nil {
		http.Error(w, "Invalid or expired verification code", http.StatusBadRequest)
		return
	}

	err = services.ChangeForgotPassword(username, request.NewPassword, request.Code)
	if err != nil {
		if err.Error() == "new password cannot be the same as the old password" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "Failed to change password", http.StatusInternalServerError)
		return
	}

	services.DeleteVerificationCode(request.Code)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password has been successfully changed"})
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	claims, err := auth.ParseTokenHeader(r)
	if err != nil {
		http.Error(w, "Unauthorized: Invalid or missing token", http.StatusUnauthorized)
		return
	}

	var request struct {
		NewPassword string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if request.NewPassword == "" {
		http.Error(w, "New password is required", http.StatusBadRequest)
		return
	}

	err = services.ChangePassword(claims.Username, request.NewPassword)
	if err != nil {
		if err.Error() == "New password cannot be the same as the old password." {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "Failed to change password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password has been successfully changed"})
}
