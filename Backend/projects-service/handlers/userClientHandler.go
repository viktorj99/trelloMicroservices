package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	userpb "pb/userpb"
	"time"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request, userClient userpb.UserServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Poziv GetAllUsers RPC metode
	response, err := userClient.GetAllUsers(ctx, &userpb.GetAllUsersRequest{})
	if err != nil {
		http.Error(w, "Greška pri dobijanju korisnika: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Vraća odgovor kao JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response.Users); err != nil {
		http.Error(w, "Greška pri enkodiranju odgovora: "+err.Error(), http.StatusInternalServerError)
		return
	}
}