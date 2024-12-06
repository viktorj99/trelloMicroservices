package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"users-service/repositories"
	"users-service/utils"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var redisClient *redis.Client

func init() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	if redisHost == "" {
		redisHost = "localhost"
	}
	if redisPort == "" {
		redisPort = "6379"
	}

	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
}

func SaveVerificationCode(ctx context.Context, code, username string) error {
	tracer := otel.Tracer("users-service")
	ctx, span := tracer.Start(ctx, "SaveVerificationCode")
	defer span.End()

	span.SetAttributes(attribute.String("verification.code", code), attribute.String("user.username", username))

	err := redisClient.Set(ctx, code, username, time.Minute*15).Err()
	if err != nil {
		span.RecordError(err)
		log.Printf("Error saving verification code to Redis: %v", err)
		return err
	}

	err = redisClient.Set(ctx, "reverse:"+code, username, time.Minute*15).Err()
	if err != nil {
		span.RecordError(err)
		log.Printf("Error saving reverse mapping to Redis: %v", err)
		return err
	}

	return nil
}

func GetUserUsernameByVerificationCode(ctx context.Context, code string) (string, error) {
	tracer := otel.Tracer("users-service")
	ctx, span := tracer.Start(ctx, "GetUserUsernameByVerificationCode")
	defer span.End()

	span.SetAttributes(attribute.String("verification.code", code))

	username, err := redisClient.Get(ctx, code).Result()
	if err != nil {
		span.RecordError(err)
		log.Printf("Error retrieving Username from Redis: %v", err)
		return "", err
	}
	span.SetAttributes(attribute.String("user.username", username))
	return username, nil
}

func DeleteVerificationCode(ctx context.Context, code string) error {
	tracer := otel.Tracer("users-service")
	ctx, span := tracer.Start(ctx, "DeleteVerificationCode")
	defer span.End()

	span.SetAttributes(attribute.String("verification.code", code))

	err := redisClient.Del(ctx, code).Err()
	if err != nil {
		span.RecordError(err)
		log.Printf("Error deleting verification code from Redis: %v", err)
		return err
	}

	return nil
}

func ActivateUser(ctx context.Context, username, expiredKey string) error {
	tracer := otel.Tracer("users-service")
	ctx, span := tracer.Start(ctx, "ActivateUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.username", username))

	user, err := repositories.GetUserByUsername(ctx, username)
	if err != nil {
		span.RecordError(err)
		return err
	}

	user.IsActive = true
	if err := repositories.UpdateUser(ctx, user.ID, user); err != nil {
		span.RecordError(err)
		return err
	}

	_, delReverseErr := redisClient.Del(ctx, "reverse:"+expiredKey).Result()
	if delReverseErr != nil {
		span.RecordError(delReverseErr)
		log.Printf("Error deleting reverse mapping for key %s: %v", expiredKey, delReverseErr)
	}

	return nil
}

func ChangeForgotPassword(ctx context.Context, username, password, expiredKey string) error {
	tracer := otel.Tracer("users-service")
	ctx, span := tracer.Start(ctx, "ChangeForgotPassword")
	defer span.End()

	span.SetAttributes(attribute.String("user.username", username))

	user, err := repositories.GetUserByUsername(ctx, username)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if utils.CheckPasswordHash(password, user.Password) {
		err := errors.New("new password cannot be the same as the old password")
		span.RecordError(err)
		return err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		span.RecordError(err)
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	_, delReverseErr := redisClient.Del(ctx, "reverse:"+expiredKey).Result()
	if delReverseErr != nil {
		span.RecordError(delReverseErr)
		log.Printf("Error deleting reverse mapping for key %s: %v", expiredKey, delReverseErr)
	}

	return repositories.UpdateUser(ctx, user.ID, user)
}

func StartKeyExpirationListener(ctx context.Context) {
	pubsub := redisClient.Subscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	log.Println("Listening for Redis key expiration events...")
	for msg := range pubsub.Channel() {
		expiredKey := msg.Payload
		log.Printf("Key expired: %s", expiredKey)

		handleExpiredKey(ctx, expiredKey)
	}
}

func handleExpiredKey(ctx context.Context, expiredKey string) {
	username, err := redisClient.Get(ctx, "reverse:"+expiredKey).Result()
	if err == redis.Nil {
		log.Printf("Key %s expired but no reverse mapping found", expiredKey)
		return
	} else if err != nil {
		log.Printf("Error fetching reverse mapping for key %s: %v", expiredKey, err)
		return
	}

	_, delReverseErr := redisClient.Del(ctx, "reverse:"+expiredKey).Result()
	if delReverseErr != nil {
		log.Printf("Error deleting reverse mapping for key %s: %v", expiredKey, delReverseErr)
	}

	err = repositories.DeleteUserByUsername(ctx, username)
	if err != nil {
		log.Printf("Error deleting user %s from MongoDB: %v", username, err)
		return
	}

	log.Printf("Successfully deleted user %s and cleaned up reverse mapping", username)
}

func ChangePassword(ctx context.Context, username, password string) error {
	tracer := otel.Tracer("users-service")
	ctx, span := tracer.Start(ctx, "ChangePassword")
	defer span.End()

	span.SetAttributes(attribute.String("user.username", username))

	user, err := repositories.GetUserByUsername(ctx, username)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if utils.CheckPasswordHash(password, user.Password) {
		err := errors.New("new password cannot be the same as the old password")
		span.RecordError(err)
		return err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		span.RecordError(err)
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	return repositories.UpdateUser(ctx, user.ID, user)
}
