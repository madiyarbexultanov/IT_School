package utils

import (
	"fmt"
	"it_school/config"
	"net/smtp"
)

// SendEmail отправляет письмо с токеном на почту
func SendEmail(to, subject, body string) error {
	from := config.Config.SMTPEmail
	password := config.Config.SMTPPassword


	smtpHost := config.Config.SMTPHost
	smtpPort := config.Config.SMTPPort

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
