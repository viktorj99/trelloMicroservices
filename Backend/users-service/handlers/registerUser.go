package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"users-service/model"
	"users-service/services"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type LoginRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	CaptchaToken string `json:"captchaToken"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("users-service").Start(r.Context(), "RegisterUserHandler")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		span.SetAttributes(attribute.String("http.method", r.Method))
		span.RecordError(http.ErrBodyNotAllowed)
		return
	}

	var user model.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		span.RecordError(err)
		return
	}
	span.SetAttributes(attribute.String("user.email", user.Email), attribute.String("user.username", user.Username))

	captchaToken := r.Header.Get("captcha_token")

	if err := services.VerifyCaptcha(ctx, captchaToken); err != nil {
		http.Error(w, "Invalid captcha", http.StatusUnauthorized)
		span.RecordError(err)
		return
	}

	err = services.RegisterUser(ctx, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		span.RecordError(err)
		return
	}

	code, err := services.GenerateVerificationCode()
	if err != nil {
		http.Error(w, "Failed to generate verification code", http.StatusInternalServerError)
		span.RecordError(err)
		return
	}

	err = services.SendVerificationEmail(ctx, user.Email, code)
	if err != nil {
		log.Printf("Error while sending verification email: %v", err)
		http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
		span.RecordError(err)
		return
	}

	err = services.SaveVerificationCode(ctx, code, user.Username)
	if err != nil {
		http.Error(w, "Failed to save verification code", http.StatusInternalServerError)
		span.RecordError(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("users-service").Start(r.Context(), "LoginUserHandler")
	defer span.End()

	log.Printf("HTTP Method: %s, Path: %s", r.Method, r.URL.Path)
	span.SetAttributes(attribute.String("http.method", r.Method), attribute.String("http.path", r.URL.Path))

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		span.RecordError(http.ErrBodyNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		span.RecordError(err)
		return
	}
	span.SetAttributes(attribute.String("user.username", req.Username))

	if err := services.VerifyCaptcha(ctx, req.CaptchaToken); err != nil {
		http.Error(w, "Invalid captcha", http.StatusUnauthorized)
		span.RecordError(err)
		return
	}

	token, err := services.Login(ctx, req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		span.RecordError(err)
		return
	}

	span.SetAttributes(attribute.String("user.token", token))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("users-service").Start(r.Context(), "GetAllUsersHandler")
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		span.RecordError(http.ErrBodyNotAllowed)
		return
	}

	users, err := services.GetAllUsers(ctx)
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		span.RecordError(err)
		return
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetUserByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("users-service").Start(r.Context(), "GetUserByIDHandler")
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		span.RecordError(http.ErrBodyNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 || pathParts[2] == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		span.RecordError(errors.New("user ID missing"))
		return
	}
	userID := pathParts[2]
	span.SetAttributes(attribute.String("user.id", userID))

	user, err := services.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		span.RecordError(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
