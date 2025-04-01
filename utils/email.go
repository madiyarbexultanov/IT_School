package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail отправляет письмо с токеном на почту
func SendEmail(to, subject, body string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")


	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Формируем заголовки и тело письма
	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

	// Аутентификация на SMTP-сервере
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Отправка письма
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	if err != nil {
		return err
	}

	return nil
}
