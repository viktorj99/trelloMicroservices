package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Claims represents the custom claims for the JWT token
type Claims struct {
	ID                 string `json:"id"`
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
func ParseTokenBody(r *http.Request) (Claims, error) {
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

func ParseTokenHeader(r *http.Request) (Claims, error) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return Claims{}, errors.New("missing token")
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return Claims{}, errors.New("invalid token format")
	}

	claimsJSON, err := decodeBase64(parts[1])
	if err != nil {
		return Claims{}, err
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return Claims{}, err
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return Claims{}, errors.New("token has expired")
	}

	message := parts[0] + "." + parts[1]
	expectedSignature, err := decodeBase64(parts[2])
	if err != nil {
		return Claims{}, err
	}

	secretKey := []byte(os.Getenv("JWT_SECRET"))
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(message))
	actualSignature := h.Sum(nil)

	if !hmac.Equal(expectedSignature, actualSignature) {
		return Claims{}, errors.New("invalid token signature")
	}

	return claims, nil
}

func decodeBase64(input string) ([]byte, error) {
	return base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(input)
}
