package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Claims represents the custom claims for the JWT token
type Claims struct {
	Username           string `json:"username"`
	Role               string `json:"role"`
	ExpiresAt          int64  `json:"exp"`
	jwt.StandardClaims        // Embedding StandardClaims to get default fields like 'exp', 'iat', etc.
}

// Valid implements the jwt.Claims interface
func (c *Claims) Valid() error {
	// Check if the token is expired
	if c.ExpiresAt < time.Now().Unix() {
		return errors.New("token has expired")
	}
	return nil
}

// ParseToken parses and validates the JWT token
func ParseToken(r *http.Request) (Claims, error) {
	// Decode the incoming JSON body into the request structure
	var request struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return Claims{}, errors.New("invalid request payload")
	}

	tokenString := request.Token
	if tokenString == "" {
		return Claims{}, errors.New("missing token")
	}

	secretKey := os.Getenv("JWT_SECRET")

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return Claims{}, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return *claims, nil
}
