package services

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

func GenerateVerificationCode() (string, error) {
	const codeLength = 6
	const charset = "0123456789"

	code := make([]byte, codeLength)
	for i := range code {
		randIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[randIndex.Int64()]
	}

	return string(code), nil
}

func sendEmail(to, subject, body string) error {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	fmt.Println("SMTP Host:", smtpHost)
	fmt.Println("SMTP Port:", smtpPort)

	from := smtpUser
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	message := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
}

func SendVerificationEmail(email string, code string) error {
	subject := "Email Verification Code"
	body := fmt.Sprintf("Your verification code is: %s", code)

	return sendEmail(email, subject, body)
}
