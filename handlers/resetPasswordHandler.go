package handlers

import (
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"it_school/utils"
	"net/http"
	"time"

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

// ResetPassword — обработчик для запроса сброса пароля
func (h *ResetPasswordHandler) ResetPassword(c *gin.Context) {
    logger := logger.GetLogger()
    var request ResetPasswordRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        logger.Warn("Invalid reset password request format", zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("invalid email format"))
        return
    }

    logger.Info("Password reset requested", zap.String("email", request.Email))

    // Пытаемся найти пользователя по email
    user, err := h.usersRepo.FindByEmail(c.Request.Context(), request.Email)
    if err != nil {
        // Если пользователь не найден, отвечаем с успехом, не раскрывая, существует ли такой email
        logger.Info("Password reset for non-existent email", zap.String("email", request.Email))
        c.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent."})
        return
    }

    // Генерация токена для сброса пароля
    resetToken, err := utils.GenerateResetToken()
    if err != nil {
        logger.Error("Failed to generate reset token", 
            zap.Int("user_id", user.Id), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate reset token"))
        return
    }

    // Устанавливаем время истечения действия токена (30 минут)
    expirationTime := time.Now().Add(30 * time.Minute)

    // Сохраняем токен в БД
    if err := h.usersRepo.SetResetToken(c.Request.Context(), request.Email, resetToken, expirationTime); err != nil {
        logger.Error("Failed to save reset token", 
            zap.Int("user_id", user.Id), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("internal server error"))
        return
    }

    logger.Info("Reset token saved", 
        zap.Int("user_id", user.Id), 
        zap.Time("expires_at", expirationTime))

    // Отправка email с токеном сброса пароля
    err = utils.SendEmail(request.Email, "Password Reset", "Your reset token: "+resetToken)
    if err != nil {
        logger.Error("Failed to send reset email", 
            zap.Int("user_id", user.Id), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to send reset email"))
        return
    }

    logger.Info("Reset email sent", zap.Int("user_id", user.Id))

    // Успешный ответ
    c.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent."})
}

// SetNewPassword — обработчик для установки нового пароля
func (h *ResetPasswordHandler) SetNewPassword(c *gin.Context) {
    logger := logger.GetLogger()
    var req SetNewPassword
    if err := c.ShouldBindJSON(&req); err != nil {
        logger.Warn("Invalid set new password request format", zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("invalid request"))
        return
    }

    logger.Info("Attempt to set new password", zap.String("reset_token", req.ResetToken))

    // Пытаемся найти пользователя по reset токену
    user, err := h.usersRepo.GetUserByResetToken(c.Request.Context(), req.ResetToken)
    if err != nil {
        logger.Warn("Invalid reset token attempt", zap.String("reset_token", req.ResetToken))
        c.JSON(http.StatusUnauthorized, models.NewApiError("invalid or expired reset token"))
        return
    }


    // Хешируем новый пароль
    hashedPassword, err := utils.HashPassword(req.NewPassword)
    if err != nil {
        logger.Error("Failed to hash new password", 
            zap.Int("user_id", user.Id), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to hash password"))
        return
    }

    // Обновляем пароль пользователя в базе данных
    if err := h.usersRepo.UpdatePassword(c.Request.Context(), user.Id, hashedPassword); err != nil {
        logger.Error("Failed to update password", 
            zap.Int("user_id", user.Id), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to update password"))
        return
    }

    // Удаляем reset токен после успешного изменения пароля
    if err := h.usersRepo.ClearResetToken(c.Request.Context(), user.Id); err != nil {
        logger.Error("Failed to clear reset token", 
            zap.Int("user_id", user.Id), 
            zap.Error(err))
        // Не прерываем выполнение, так как пароль уже изменен
    }

    logger.Info("Password successfully reset", zap.Int("user_id", user.Id))

    // Успешный ответ
    c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}