package utils

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(plainTextPassword string) (string, error) {
	const cost = bcrypt.DefaultCost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), cost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return "", err
	}
	return string(hashedPassword), nil
}
