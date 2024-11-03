package services

import (
	"errors"
	"regexp"
	"strings"
	"users-service/model"
	"users-service/repositories"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

var validDomains = []string{"gmail.com", "yahoo.com", "outlook.com", "hotmail.com", "example.com"}

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
		return errors.New("All fields are required")
	}

	if user.Role != model.RoleManager && user.Role != model.RoleMember {
		return errors.New("Invalid role; must be 'Manager' or 'Member'")
	}

	if !emailRegex.MatchString(user.Email) || !isValidDomain(user.Email) {
		return errors.New("Invalid email format or domain")
	}

	err := repositories.CreateUser(user)
	if err != nil {
		return err
	}

	return nil
}
