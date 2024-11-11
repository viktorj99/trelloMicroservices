package services

import (
	"context"
	"log"
	"time"
	"users-service/repositories"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var redisClient = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379", // Redis address
	Password: "",               // No password set
	DB:       0,                // Use default DB
})

func SaveVerificationCode(code string, username string) error {
	// Save the verification code in Redis with an expiration time of 15 minutes
	err := redisClient.Set(ctx, code, username, time.Minute*15).Err()
	if err != nil {
		log.Printf("Error saving verification code to Redis: %v", err)
		return err
	}
	return nil
}

func GetUserUsernameByVerificationCode(code string) (string, error) {
	// Retrieve the user ID associated with the verification code from Redis
	username, err := redisClient.Get(ctx, code).Result()
	if err != nil {
		log.Printf("Error retrieving Username from Redis: %v", err)
		return "", err
	}
	return username, nil
}

func DeleteVerificationCode(code string) error {
	// Delete the verification code from Redis once it is used
	err := redisClient.Del(ctx, code).Err()
	if err != nil {
		log.Printf("Error deleting verification code from Redis: %v", err)
		return err
	}
	return nil
}

func ActivateUser(username string) error {
	// Update the user's isActive field to true
	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		return err
	}

	user.IsActive = true
	return repositories.UpdateUser(user.ID, user)
}
