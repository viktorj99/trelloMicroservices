package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"
	"users-service/model"
	"users-service/repositories"
	"users-service/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/consul/api"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/crypto/bcrypt"
)

const recaptchaSecret = "6LeNfI8qAAAAAHUP6tTpTDb0uGtOwvKTDDIIPV6Y"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var passwordLengthRegex = regexp.MustCompile(`^[A-Za-z\d]{8,}$`)

var validDomains = []string{"gmail.com", "yahoo.com", "outlook.com", "hotmail.com", "example.com"}

type Claims struct {
	ID       string `json:"id"`
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

var consulClient *api.Client

func init() {
	var err error

	consulAddress := os.Getenv("CONSUL_DB")
	consulPort := os.Getenv("CONSUL_DB_PORT")

	if consulAddress == "" {
		consulAddress = "localhost"
	}
	if consulPort == "" {
		consulPort = "8500"
	}

	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%s", consulAddress, consulPort)

	consulClient, err = api.NewClient(config)
	if err != nil {
		fmt.Printf("Failed to create Consul client: %v\n", err)
	}
}

func isCommonPassword(ctx context.Context, password string) (bool, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "isCommonPassword")
	defer span.End()

	span.SetAttributes(attribute.String("password", password))

	key := fmt.Sprintf("common_passwords/%s", password)
	kvPair, _, err := consulClient.KV().Get(key, nil)
	if err != nil {
		span.RecordError(err)
		return false, err
	}

	return kvPair != nil, nil
}

func RegisterUser(ctx context.Context, user model.User) error {
	ctx, span := otel.Tracer("users-service").Start(ctx, "RegisterUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.firstName", user.FirstName),
		attribute.String("user.lastName", user.LastName),
		attribute.String("user.email", user.Email),
		attribute.String("user.username", user.Username),
	)

	isCommon, err := isCommonPassword(ctx, user.Password)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("error checking common password: %v", err)
	}
	if isCommon {
		err := errors.New("password is too common, please choose a more secure password")
		span.RecordError(err)
		return err
	}

	sanitizer := bluemonday.StrictPolicy()
	user.FirstName = sanitizer.Sanitize(user.FirstName)
	user.LastName = sanitizer.Sanitize(user.LastName)
	user.Email = sanitizer.Sanitize(user.Email)
	user.Username = sanitizer.Sanitize(user.Username)

	if user.FirstName == "" || user.LastName == "" || user.Email == "" || user.Username == "" || user.Password == "" || user.Role == "" {
		err := errors.New("all fields are required")
		span.RecordError(err)
		return err
	}

	if user.Role != model.RoleManager && user.Role != model.RoleMember {
		err := errors.New("invalid role; must be 'Manager' or 'Member'")
		span.RecordError(err)
		return err
	}

	if !emailRegex.MatchString(user.Email) || !isValidDomain(user.Email) {
		err := errors.New("invalid email format or domain")
		span.RecordError(err)
		return err
	}

	if !isValidPassword(user.Password) {
		err := errors.New("password must be at least 8 characters long, contain one uppercase letter, one lowercase letter, and one digit")
		span.RecordError(err)
		return err
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		span.RecordError(err)
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	_, err = repositories.CreateUser(ctx, user)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLower := strings.IndexFunc(password, unicode.IsLower) >= 0

	hasUpper := strings.IndexFunc(password, unicode.IsUpper) >= 0

	hasDigit := strings.IndexFunc(password, unicode.IsDigit) >= 0

	return hasLower && hasUpper && hasDigit
}

func Login(ctx context.Context, username, password string) (string, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "Login")
	defer span.End()

	span.SetAttributes(attribute.String("user.username", username))

	user, err := repositories.GetUserByUsername(ctx, username)
	if err != nil {
		span.RecordError(err)
		return "", errors.New("user not found")
	}

	if user.IsActive == false {
		err := errors.New("account is not active")
		span.RecordError(err)
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		err := errors.New("invalid password")
		span.RecordError(err)
		return "", err
	}

	token, err := GenerateJWTToken(user.ID, user.Username, user.Role)
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	span.SetAttributes(attribute.String("user.token", token))
	return token, nil
}

func GenerateJWTToken(id, username, role string) (string, error) {
	if role != model.RoleManager && role != model.RoleMember {
		return "", errors.New("invalid role")
	}

	expirationTime := time.Now().Add(30 * time.Minute)

	claims := &Claims{
		ID:       id,
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

func GetAllUsers(ctx context.Context) ([]model.User, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "GetAllUsers")
	defer span.End()

	users, err := repositories.GetAllUsers(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))
	return users, nil
}

func GetAllUserMembers(ctx context.Context) ([]model.User, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "GetAllUserMembers")
	defer span.End()

	users, err := repositories.GetAllUserMembers(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))
	return users, nil
}

func GetUserByID(ctx context.Context, userID string) (model.User, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "GetUserByID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID))

	user, err := repositories.GetUserByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return model.User{}, err
	}

	return user, nil
}

func GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "GetUserByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	user, err := repositories.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		return model.User{}, err
	}

	return user, nil
}

func DeleteUser(ctx context.Context, userId string) (*mongo.DeleteResult, error) {
	ctx, span := otel.Tracer("users-service").Start(ctx, "DeleteUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userId))

	result, err := repositories.DeleteUserById(ctx, userId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int64("deletedCount", result.DeletedCount))
	return result, nil
}

func VerifyCaptcha(ctx context.Context, captchaToken string) error {
	ctx, span := otel.Tracer("users-service").Start(ctx, "VerifyCaptcha")
	defer span.End()

	span.SetAttributes(attribute.String("captcha.token", captchaToken))

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
		"secret":   {recaptchaSecret},
		"response": {captchaToken},
	})
	if err != nil {
		span.RecordError(err)
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		return err
	}

	if !result.Success {
		err := errors.New("captcha verification failed")
		span.RecordError(err)
		return err
	}

	return nil
}
