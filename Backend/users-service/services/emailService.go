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

	from := smtpUser
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	message := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" + body + "\r\n")

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
}

func SendVerificationEmail(email string, code string) error {
	subject := "Email Verification Code"
	body := fmt.Sprintf(`
	<html>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px; margin: 0;">
			<div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 20px; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);">
				<h2 style="color: #333333; text-align: center;">Email Verification</h2>
				<p style="font-size: 16px; color: #555555; text-align: center;">Your verification code is:</p>
				<div style="text-align: center; margin: 20px 0;">
					<span style="font-size: 24px; font-weight: bold; color: #4CAF50;">%s</span>
				</div>
				<p style="font-size: 16px; color: #555555; text-align: center;">Please use this code to complete your email verification process. If you didn't request this, you can safely ignore this email.</p>
				<hr style="border: none; border-top: 1px solid #eeeeee; margin: 20px 0;">
				<p style="font-size: 12px; color: #999999; text-align: center;">Thank you for signing up! We're excited to have you on board.</p>
			</div>
		</body>
	</html>`, code)

	return sendEmail(email, subject, body)
}

func SendPasswordResetEmail(email string, code string) error {
	subject := "Password Reset Verification Code"
	body := fmt.Sprintf(`
	<html>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px; margin: 0;">
			<div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 20px; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);">
				<h2 style="color: #333333; text-align: center;">Password Reset Request</h2>
				<p style="font-size: 16px; color: #555555; text-align: center;">You have requested to reset your password. Your password reset verification code is:</p>
				<div style="text-align: center; margin: 20px 0;">
					<span style="font-size: 24px; font-weight: bold; color: #4CAF50;">%s</span>
				</div>
				<p style="font-size: 16px; color: #555555; text-align: center;">Please use this code to reset your password. If you did not request this change, you can safely ignore this email.</p>
				<hr style="border: none; border-top: 1px solid #eeeeee; margin: 20px 0;">
				<p style="font-size: 12px; color: #999999; text-align: center;">If you need further assistance, please contact our support team.</p>
			</div>
		</body>
	</html>`, code)

	return sendEmail(email, subject, body)
}

func SendMagicLinkEmail(email, token string) error {
	link := fmt.Sprintf("http://localhost:5173/magic-login?token=%s", token)
	subject := "Magic Link Login"
	body := fmt.Sprintf(`
    <html>
        <body style="font-family: Arial, sans-serif;">
            <div style="max-width: 600px; margin: auto; padding: 20px; border: 1px solid #ddd; border-radius: 8px; background-color: #f9f9f9;">
                <h2 style="text-align: center; color: #333;">Login with Magic Link</h2>
                <p style="color: #555;">Click the button below to log in to your account:</p>
                <p style="text-align: center;">
                    <a href="%s" style="padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px;">Login</a>
                </p>
                <p style="font-size: 12px; color: #999;">If you did not request this email, please ignore it.</p>
            </div>
        </body>
    </html>`, link)
	return sendEmail(email, subject, body)
}
