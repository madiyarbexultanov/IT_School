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
	authRepo *repositories.AuthRepository
    usersRepo *repositories.UsersRepository
}

type SetNewPassword struct {
	ResetToken string `json:"reset_token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func NewResetPasswordHandler(authRepo *repositories.AuthRepository, usersRepo *repositories.UsersRepository) *ResetPasswordHandler {
	return &ResetPasswordHandler{authRepo: authRepo, usersRepo: usersRepo}
}

// ResetPassword godoc
// @Summary Запрос сброса пароля
// @Description Инициирует процесс сброса пароля по email. Отправляет токен сброса на указанный email (если он существует в системе).
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Email для сброса пароля" example={"email": "user@example.com"}
// @Success 200 {object} models.MessageResponse "Всегда возвращает успех, даже если email не существует (security through obscurity)"
// @Failure 400 {object} models.ApiError "Неверный формат email"
// @Failure 500 {object} models.ApiError "Ошибка сервера при обработке запроса"
// @Router /auth/reset-password [post]
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
            zap.String("user_id", user.Id.String()), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate reset token"))
        return
    }

    // Устанавливаем время истечения действия токена (30 минут)
    expirationTime := time.Now().Add(30 * time.Minute)

    // Сохраняем токен в БД
    if err := h.authRepo.SetResetToken(c.Request.Context(), request.Email, resetToken, expirationTime); err != nil {
        logger.Error("Failed to save reset token", 
            zap.String("user_id", user.Id.String()), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("internal server error"))
        return
    }

    logger.Info("Reset token saved", 
        zap.String("user_id", user.Id.String()), 
        zap.Time("expires_at", expirationTime))

    // Отправка email с токеном сброса пароля
    err = utils.SendEmail(request.Email, "Password Reset", "Your reset token: "+resetToken)
    if err != nil {
        logger.Error("Failed to send reset email", 
            zap.String("user_id", user.Id.String()), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to send reset email"))
        return
    }

    logger.Info("Reset email sent", zap.String("user_id", user.Id.String()))

    // Успешный ответ
    c.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent."})
}

// SetNewPassword godoc
// @Summary Установка нового пароля
// @Description Устанавливает новый пароль после сброса. Требует валидный токен сброса.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SetNewPassword true "Данные для сброса пароля" example={"reset_token": "valid-reset-token-123", "new_password": "newSecurePassword123"}
// @Success 200 {object} models.MessageResponse "Пароль успешно обновлен"
// @Failure 400 {object} models.ApiError "Неверный формат запроса"
// @Failure 401 {object} models.ApiError "Недействительный или просроченный токен"
// @Failure 500 {object} models.ApiError "Ошибка сервера при обновлении пароля"
// @Router /auth/new-password [post]
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
    user, err := h.authRepo.GetUserByResetToken(c.Request.Context(), req.ResetToken)
    if err != nil {
        logger.Warn("Invalid reset token attempt", zap.String("reset_token", req.ResetToken))
        c.JSON(http.StatusUnauthorized, models.NewApiError("invalid or expired reset token"))
        return
    }


    // Хешируем новый пароль
    hashedPassword, err := utils.HashPassword(req.NewPassword)
    if err != nil {
        logger.Error("Failed to hash new password", 
            zap.String("user_id", user.Id.String()), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to hash password"))
        return
    }

    // Обновляем пароль пользователя в базе данных
    if err := h.authRepo.UpdatePassword(c.Request.Context(), user.Id, hashedPassword); err != nil {
        logger.Error("Failed to update password", 
            zap.String("user_id", user.Id.String()), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to update password"))
        return
    }

    // Удаляем reset токен после успешного изменения пароля
    if err := h.authRepo.ClearResetToken(c.Request.Context(), user.Id); err != nil {
        logger.Error("Failed to clear reset token", 
            zap.String("user_id", user.Id.String()), 
            zap.Error(err))
        // Не прерываем выполнение, так как пароль уже изменен
    }

    logger.Info("Password successfully reset", zap.String("user_id", user.Id.String()))

    // Успешный ответ
    c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}