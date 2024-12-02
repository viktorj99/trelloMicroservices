package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"users-service/services"
)

func DeleteUserById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("url", r.URL.Path)
	parts := strings.Split(r.URL.Path, "/")
	fmt.Println("url", parts)
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	userId := parts[3]

	if userId == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	result, err := services.DeleteUser(r.Context(), userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "User deleted successfully",
		"deletedCount": result.DeletedCount,
	})
}