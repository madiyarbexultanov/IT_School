package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"it_school/config"
	"it_school/logger"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error){
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(passwordHash), nil
}


func GenerateRefreshToken(userID int) (string, error) {
    logger := logger.GetLogger()
    // Генерация случайных данных для подписи
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        logger.Error("Failed to generate random bytes for refresh token", zap.Error(err))
        return "", err
    }

    // Кодируем user_id в base64
    userIDBase64 := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", userID)))

    // HMAC для подписи
    mac := hmac.New(sha256.New, []byte(config.Config.JwtSecretKey))
    mac.Write(b)
    signature := mac.Sum(nil)

    // Генерация refresh token в формате userID.signature
    refreshToken := fmt.Sprintf("%s.%s", userIDBase64, base64.URLEncoding.EncodeToString(signature))
    
    logger.Debug("Refresh token generated", zap.Int("user_id", userID))

    return refreshToken, nil
}

// generateResetToken — генерирует случайный токен для сброса пароля
func GenerateResetToken() (string, error) {
    logger := logger.GetLogger()
    b := make([]byte, 16) // Создаем 16 байт случайных данных
    _, err := rand.Read(b)
    if err != nil {
        logger.Error("Failed to generate random bytes for reset token", zap.Error(err))
        return "", err // Ошибка при генерации токена
    }
    token := hex.EncodeToString(b)
    logger.Debug("Reset token generated", zap.String("token", token))
    return token, nil // Возвращаем токен в виде строки
}

// checkPasswordHash — проверяет правильность пароля, сравнивая его с хешом
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}