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
	err := redisClient.Set(ctx, code, username, time.Minute*15).Err()
	if err != nil {
		log.Printf("Error saving verification code to Redis: %v", err)
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

func ActivateUser(username string) error {
	user, err := repositories.GetUserByUsername(username)
	if err != nil {
		return err
	}

	user.IsActive = true
	return repositories.UpdateUser(user.ID, user)
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
