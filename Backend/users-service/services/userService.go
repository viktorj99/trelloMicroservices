package services

import (
	"errors"
	"users-service/model"
	"users-service/repositories"
)

func RegisterUser(user model.User) error {
	if user.FirstName == "" || user.LastName == "" || user.Email == "" || user.Username == "" || user.Password == "" {
		return errors.New("all fields are required")
	}

	return repositories.CreateUser(user)
}
