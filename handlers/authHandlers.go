package handlers

import (
	"context"
	"it_school/config"
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"it_school/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthHandler struct {
	usersRepo    *repositories.UsersRepository
	sessionsRepo *repositories.SessionsRepository
	rolesRepo 	 *repositories.RoleRepository
}

func NewAuthHandler(usersRepo *repositories.UsersRepository, sessionsRepo *repositories.SessionsRepository, rolesRepo *repositories.RoleRepository) *AuthHandler {
	return &AuthHandler{
		usersRepo:    usersRepo,
		sessionsRepo: sessionsRepo,
		rolesRepo: 	  rolesRepo,
	}
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Вход в систему с email и паролем
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body AuthRequest true "Данные для входа"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ApiError
// @Failure 401 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
    logger := logger.GetLogger()
    var req AuthRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        logger.Warn("Invalid login request format", zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
        return
    }

    // Пытаемся найти пользователя по email
    user, err := h.usersRepo.FindByEmail(c.Request.Context(), req.Email)
    if err != nil {
        logger.Info("Login attempt with non-existent email", zap.String("email", req.Email))
        c.JSON(http.StatusUnauthorized, models.NewApiError("invalid credentials"))
        return
    }

    // Проверяем правильность пароля
    if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
        logger.Warn("Invalid password attempt", zap.String("email", req.Email))
        c.JSON(http.StatusUnauthorized, models.NewApiError("invalid credentials"))
        return
    }

    // Получаем роль пользователя
    role, err := h.rolesRepo.GetRoleByID(c.Request.Context(), user.RoleID)
    if err != nil {
        logger.Error("Failed to get user role", zap.String("user_id", user.Id.String()), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't find role"))
        return
    }

    // Генерация JWT токена
    token, err := h.generateJWTToken(c.Request.Context(), user.Id, user.RoleID)
    if err != nil {
        logger.Error("Failed to generate JWT token", zap.String("user_id", user.Id.String()), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate token"))
        return
    }

    // Генерация refresh токена
    refreshToken, err := utils.GenerateRefreshToken(user.Id)
    if err != nil {
        logger.Error("Failed to generate refresh token", zap.String("user_id", user.Id.String()), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate refresh token"))
        return
    }

    // Создаем сессию пользователя
    session := models.Session{
        UserID:       user.Id,
        RefreshToken: refreshToken,
        ExpiresAt:    time.Now().Add(time.Hour * 24 * 7), // Срок действия сессии — 7 дней
    }

    // Сохраняем сессию в репозитории
    if err := h.sessionsRepo.CreateSession(c.Request.Context(), session); err != nil {
        logger.Error("Failed to create session", zap.String("user_id", user.Id.String()), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to create session"))
        return
    }

    // Устанавливаем cookie с refresh токеном
    c.SetCookie("session_token", refreshToken, int(session.ExpiresAt.Unix()), "/", "", false, true)

    logger.Info("Successful login", zap.String("user_id", user.Id.String()), zap.String("role", role.Name))

    // Ответ с JWT токеном и ролью пользователя
    c.JSON(http.StatusOK, gin.H{
        "token":   token,
        "user":    user,
    })
}


// Logout godoc
// @Summary Выход из системы
// @Description Завершает текущую сессию пользователя
// @Tags auth
// @Produce json
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Security ApiKeyAuth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
    logger := logger.GetLogger()
    // Получаем session token из cookie
    sessionToken, err := c.Cookie("session_token")
    if err != nil {
        logger.Warn("Logout attempt without session token")
        c.JSON(http.StatusBadRequest, models.NewApiError("no session token"))
        return
    }

    // Удаляем сессию по session token
    if err := h.sessionsRepo.DeleteSession(c.Request.Context(), sessionToken); err != nil {
        logger.Error("Failed to delete session", zap.String("session_token", sessionToken), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to delete session"))
        return
    }

    // Удаляем cookie с session token
    c.SetCookie("session_token", "", -1, "/", "", false, true)

    logger.Info("Successful logout", zap.String("session_token", sessionToken))

    // Ответ о успешном выходе
    c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}


// Refresh godoc
// @Summary Обновление токена
// @Description Обновляет JWT токен с помощью refresh токена из cookie
// @Tags auth
// @Produce json
// @Success 200 {object} models.TokenResponse  // Убрано слово "object"
// @Failure 401 {object} models.ErrorResponse // Используем ErrorResponse вместо ApiError
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
    logger := logger.GetLogger()
    sessionToken, err := c.Cookie("session_token")
    if err != nil {
        logger.Warn("Refresh attempt without session token")
        c.JSON(http.StatusUnauthorized, models.NewApiError("no session token"))
        return
    }

    session, roleID, err := h.sessionsRepo.GetSession(c.Request.Context(), sessionToken)
    if err != nil {
        logger.Warn("Invalid session token", zap.String("session_token", sessionToken), zap.Error(err))
        c.JSON(http.StatusUnauthorized, models.NewApiError("invalid session token"))
        return
    }
    
    if time.Now().After(session.ExpiresAt) {
        logger.Warn("Expired session token", zap.String("session_token", sessionToken))
        c.JSON(http.StatusUnauthorized, models.NewApiError("expired session token"))
        return
    }

    token, err := h.generateJWTToken(c.Request.Context(), session.UserID, roleID)
    if err != nil {
        logger.Error("Failed to generate JWT token", 
            zap.String("user_id", session.UserID.String()),  
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate token"))
        return
    }

    newRefreshToken, err := utils.GenerateRefreshToken(session.UserID)
    if err != nil {
        logger.Error("Failed to generate refresh token", 
            zap.String("user_id", session.UserID.String()),  
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate refresh token"))
        return
    }

    session.RefreshToken = newRefreshToken
    session.ExpiresAt = time.Now().Add(time.Hour * 24 * 7)

    if err := h.sessionsRepo.UpdateSession(c.Request.Context(), session); err != nil {
        logger.Error("Failed to update session", 
            zap.String("user_id", session.UserID.String()),  
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to update session"))
        return
    }

    c.SetCookie("session_token", newRefreshToken, int(session.ExpiresAt.Unix()), "/", "", false, true)

    logger.Info("Tokens refreshed successfully", zap.String("user_id", session.UserID.String()))

    c.JSON(http.StatusOK, gin.H{
        "token":   token,
        "expires": time.Now().Add(time.Hour * 1).Unix(),
    })
}


func (h *AuthHandler) generateJWTToken(c context.Context, userID, roleID uuid.UUID) (string, error) {
    logger := logger.GetLogger()
    // Находим пользователя по его ID
    user, err := h.usersRepo.FindById(c, userID)
    if err != nil {
        logger.Error("Failed to find user by ID", 
            zap.String("user_id", userID.String()), 
            zap.Error(err))
        return "", err
    }

    // Получаем роль пользователя
    role, err := h.rolesRepo.GetRoleByID(c, user.RoleID)
    if err != nil {
        logger.Error("Failed to get user role", 
        zap.String("userID", userID.String()), 
            zap.Error(err))
        return "", err
    }

    // Создаем JWT токен с ролью и ID пользователя
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":     userID.String(),
        "role":    role.Name,
        "role_id": roleID, // Добавлено role_id для более удобной проверки
        "exp":     time.Now().Add(time.Hour * 1).Unix(), // Время истечения токена — 1 час
    })

    logger.Debug("JWT token generated", zap.String("userID", userID.String()), zap.String("role", role.Name))

    // Подписываем и возвращаем токен
    return token.SignedString([]byte(config.Config.JwtSecretKey))
}