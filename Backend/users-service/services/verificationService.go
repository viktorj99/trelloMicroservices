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
)

var ctx = context.Background()
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

func SaveVerificationCode(code string, username string) error {
	err := redisClient.Set(ctx, code, username, time.Minute*15).Err()
	if err != nil {
		log.Printf("Error saving verification code to Redis: %v", err)
		return err
	}

	err = redisClient.Set(ctx, "reverse:"+code, username, time.Minute*15).Err()
	if err != nil {
		log.Printf("Error saving reverse mapping to Redis: %v", err)
		return err
	}

	return nil
}

func GetUserUsernameByVerificationCode(code string) (string, error) {
	username, err := redisClient.Get(ctx, code).Result()
	if err != nil {
		log.Printf("Error retrieving Username from Redis: %v", err)
		return "", err
	}
	return username, nil
}

func DeleteVerificationCode(code string) error {
	err := redisClient.Del(ctx, code).Err()
	if err != nil {
		log.Printf("Error deleting verification code from Redis: %v", err)
		return err
	}
	return nil
}

func ActivateUser(username string, expiredKey string) error {
	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		return err
	}

	user.IsActive = true
	if err := repositories.UpdateUser(user.ID, user); err != nil {
		return err
	}

	_, delReverseErr := redisClient.Del(ctx, "reverse:"+expiredKey).Result()
	if delReverseErr != nil {
		log.Printf("Error deleting reverse mapping for key %s: %v", expiredKey, delReverseErr)
	}

	return nil
}

func ChangeForgotPassword(username string, password string, expiredKey string) error {
	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		return err
	}

	if utils.CheckPasswordHash(password, user.Password) {
		return errors.New("new password cannot be the same as the old password")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	_, delReverseErr := redisClient.Del(ctx, "reverse:"+expiredKey).Result()
	if delReverseErr != nil {
		log.Printf("Error deleting reverse mapping for key %s: %v", expiredKey, delReverseErr)
	}

	return repositories.UpdateUser(user.ID, user)
}

func StartKeyExpirationListener() {
	pubsub := redisClient.Subscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	log.Println("Listening for Redis key expiration events...")
	for msg := range pubsub.Channel() {
		expiredKey := msg.Payload
		log.Printf("Key expired: %s", expiredKey)

		handleExpiredKey(expiredKey)
	}
}

func handleExpiredKey(expiredKey string) {
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

	err = repositories.DeleteUserByUsername(username)
	if err != nil {
		log.Printf("Error deleting user %s from MongoDB: %v", username, err)
		return
	}

	log.Printf("Successfully deleted user %s and cleaned up reverse mapping", username)
}

func ChangePassword(username string, password string) error {
	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		return err
	}

	if utils.CheckPasswordHash(password, user.Password) {
		return errors.New("New password cannot be the same as the old password.")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	return repositories.UpdateUser(user.ID, user)
}
