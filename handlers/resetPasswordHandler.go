package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"it_school/models"
	"it_school/repositories"
	"it_school/utils"
    "it_school/logger"
	"net/http"

	"github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordHandler struct {
	usersRepo *repositories.UsersRepository
}

type SetNewPassword struct {
	ResetToken string `json:"reset_token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func NewResetPasswordHandler(usersRepo *repositories.UsersRepository) *ResetPasswordHandler {
	return &ResetPasswordHandler{usersRepo: usersRepo}
}

// generateResetToken — генерирует случайный токен для сброса пароля
func generateResetToken() (string, error) {
	b := make([]byte, 16) // Создаем 16 байт случайных данных
	_, err := rand.Read(b)
	if err != nil {
		return "", err // Ошибка при генерации токена
	}
	return hex.EncodeToString(b), nil // Возвращаем токен в виде строки
}

// ResetPassword — обработчик для запроса сброса пароля
func (h *ResetPasswordHandler) ResetPassword(c *gin.Context) {
	var request ResetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("invalid email format"))
		return
	}

	// Пытаемся найти пользователя по email
	user, err := h.usersRepo.FindByEmail(c.Request.Context(), request.Email)
	if err != nil {
		// Если пользователь не найден, отвечаем с успехом, не раскрывая, существует ли такой email
		c.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent."})
		return
	}

	// Генерация токена для сброса пароля
	resetToken, err := generateResetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate reset token"))
		return
	}

	// Устанавливаем время истечения действия токена (30 минут)
	expirationTime := time.Now().Add(30 * time.Minute)

	// Сохраняем токен в БД
	if err := h.usersRepo.SetResetToken(c.Request.Context(), request.Email, resetToken, expirationTime); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("internal server error"))
		return
	}

	// Отправка email с токеном сброса пароля
	err = utils.SendEmail(request.Email, "Password Reset", "Your reset token: "+resetToken)
	if err != nil {
		// Ошибка при отправке email
		c.JSON(http.StatusInternalServerError, models.NewApiError("failed to send reset email"))
		return
	}

	// Успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent."})
}

func (h *ResetPasswordHandler) SetNewPassword(c *gin.Context) {
	var req SetNewPassword
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("invalid request"))
		return
	}

	// Пытаемся найти пользователя по reset токену
	user, err := h.usersRepo.GetUserByResetToken(c.Request.Context(), req.ResetToken)
	if err != nil || time.Now().After(user.ResetTokenExpiresAt) {
		// Токен недействителен или истек
		c.JSON(http.StatusUnauthorized, models.NewApiError("invalid or expired reset token"))
		return
	}

	// Хешируем новый пароль
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("failed to hash password"))
		return
	}

	// Обновляем пароль пользователя в базе данных
	if err := h.usersRepo.UpdatePassword(c.Request.Context(), user.Id, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("failed to update password"))
		return
	}

	// Удаляем reset токен после успешного изменения пароля
	h.usersRepo.ClearResetToken(c.Request.Context(), user.Id)

	// Успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}