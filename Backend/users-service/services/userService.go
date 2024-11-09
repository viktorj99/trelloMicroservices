package services

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"time"
	"users-service/model"
	"users-service/repositories"
	"users-service/utils"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

var validDomains = []string{"gmail.com", "yahoo.com", "outlook.com", "hotmail.com", "example.com"}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func isValidDomain(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := parts[1]
	for _, validDomain := range validDomains {
		if domain == validDomain {
			return true
		}
	}
	return false
}

func RegisterUser(user model.User) error {
	if user.FirstName == "" || user.LastName == "" || user.Email == "" || user.Username == "" || user.Password == "" || user.Role == "" {
		return errors.New("all fields are required")
	}

	if user.Role != model.RoleManager && user.Role != model.RoleMember {
		return errors.New("invalid role; must be 'Manager' or 'Member'")
	}

	if !emailRegex.MatchString(user.Email) || !isValidDomain(user.Email) {
		return errors.New("invalid email format or domain")
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	err = repositories.CreateUser(user)
	if err != nil {
		return err
	}

	return nil
}

func Login(username, password string) (string, error) {
	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		return "", errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid password")
	}

	token, err := generateJWTToken(user.Username, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}

func generateJWTToken(username, role string) (string, error) {
	if role != model.RoleManager && role != model.RoleMember {
		return "", errors.New("invalid role")
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
