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
)

type Claims struct {
	Username  string `json:"username"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
}

func ParseToken(r *http.Request) (Claims, error) {
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
