package services

import (
	"context"
	"errors"
	"log"
	"time"
	"users-service/repositories"
	"users-service/utils"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var redisClient = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func SaveVerificationCode(code string, username string) error {
	// Save the code as the key with the username as the value
	err := redisClient.Set(ctx, code, username, time.Minute*15).Err()
	if err != nil {
		log.Printf("Error saving verification code to Redis: %v", err)
		return err
	}

	// Save a reverse mapping from code to username
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
	// Fetch the user from MongoDB
	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		return err
	}

	// Activate the user
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

func ChangePassword(username string, password string) error {
	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		return err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword

	return repositories.UpdateUser(user.ID, user)
}

func StartKeyExpirationListener() {
	// Subscribe to key expiration events
	pubsub := redisClient.Subscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	log.Println("Listening for Redis key expiration events...")
	for msg := range pubsub.Channel() {
		// Parse the expired key
		expiredKey := msg.Payload
		log.Printf("Key expired: %s", expiredKey)

		// Handle the expired key
		handleExpiredKey(expiredKey)
	}
}

func handleExpiredKey(expiredKey string) {
	// Use the reverse mapping to find the associated username
	username, err := redisClient.Get(ctx, "reverse:"+expiredKey).Result()
	if err == redis.Nil {
		log.Printf("Key %s expired but no reverse mapping found", expiredKey)
		return
	} else if err != nil {
		log.Printf("Error fetching reverse mapping for key %s: %v", expiredKey, err)
		return
	}

	// Delete the reverse mapping
	_, delReverseErr := redisClient.Del(ctx, "reverse:"+expiredKey).Result()
	if delReverseErr != nil {
		log.Printf("Error deleting reverse mapping for key %s: %v", expiredKey, delReverseErr)
	}

	// Delete the user from MongoDB
	err = repositories.DeleteUserByUsername(username)
	if err != nil {
		log.Printf("Error deleting user %s from MongoDB: %v", username, err)
		return
	}

	log.Printf("Successfully deleted user %s and cleaned up reverse mapping", username)
}
