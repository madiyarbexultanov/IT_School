package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"it_school/config"
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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
    if !checkPasswordHash(req.Password, user.PasswordHash) {
        logger.Warn("Invalid password attempt", zap.String("email", req.Email))
        c.JSON(http.StatusUnauthorized, models.NewApiError("invalid credentials"))
        return
    }

    // Получаем роль пользователя
    role, err := h.rolesRepo.GetRoleByID(c.Request.Context(), user.RoleID)
    if err != nil {
        logger.Error("Failed to get user role", zap.Int("user_id", user.Id), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't find role"))
        return
    }

    // Генерация JWT токена
    token, err := h.generateJWTToken(c.Request.Context(), user.Id, user.RoleID)
    if err != nil {
        logger.Error("Failed to generate JWT token", zap.Int("user_id", user.Id), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate token"))
        return
    }

    // Генерация refresh токена
    refreshToken, err := generateRefreshToken(user.Id)
    if err != nil {
        logger.Error("Failed to generate refresh token", zap.Int("user_id", user.Id), zap.Error(err))
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
        logger.Error("Failed to create session", zap.Int("user_id", user.Id), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to create session"))
        return
    }

    // Устанавливаем cookie с refresh токеном
    c.SetCookie("session_token", refreshToken, int(session.ExpiresAt.Unix()), "/", "", false, true)

    logger.Info("Successful login", zap.Int("user_id", user.Id), zap.String("role", role.Name))

    // Ответ с JWT токеном и ролью пользователя
    c.JSON(http.StatusOK, gin.H{
        "token":   token,
        "role":    role.Name,
        "expires": time.Now().Add(time.Hour * 1).Unix(), // Время истечения JWT токена (1 час)
    })
}

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

func (h *AuthHandler) Refresh(c *gin.Context) {
    logger := logger.GetLogger()
    // Получаем session token из cookie
    sessionToken, err := c.Cookie("session_token")
    if err != nil {
        logger.Warn("Refresh attempt without session token")
        c.JSON(http.StatusUnauthorized, models.NewApiError("no session token"))
        return
    }

    // Получаем сессию по session token
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

    // Генерация нового JWT токена
    token, err := h.generateJWTToken(c.Request.Context(), session.UserID, roleID)
    if err != nil {
        logger.Error("Failed to generate JWT token", 
            zap.Int("user_id", session.UserID), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate token"))
        return
    }

    // Генерация нового refresh токена
    newRefreshToken, err := generateRefreshToken(session.UserID)
    if err != nil {
        logger.Error("Failed to generate refresh token", 
            zap.Int("user_id", session.UserID), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate refresh token"))
        return
    }

    // Обновляем сессию с новым refresh токеном и временем истечения
    session.RefreshToken = newRefreshToken
    session.ExpiresAt = time.Now().Add(time.Hour * 24 * 7)

    // Сохраняем обновленную сессию
    if err := h.sessionsRepo.UpdateSession(c.Request.Context(), session); err != nil {
        logger.Error("Failed to update session", 
            zap.Int("user_id", session.UserID), 
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("failed to update session"))
        return
    }

    // Устанавливаем новый session token в cookie
    c.SetCookie("session_token", newRefreshToken, int(session.ExpiresAt.Unix()), "/", "", false, true)

    logger.Info("Tokens refreshed successfully", zap.Int("user_id", session.UserID))

    // Ответ с новым токеном
    c.JSON(http.StatusOK, gin.H{
        "token":   token,
        "expires": time.Now().Add(time.Hour * 1).Unix(), // Время истечения нового токена
    })
}

func (h *AuthHandler) generateJWTToken(ctx context.Context, userID, roleID int) (string, error) {
    logger := logger.GetLogger()
    // Находим пользователя по его ID
    user, err := h.usersRepo.FindById(ctx, userID)
    if err != nil {
        logger.Error("Failed to find user by ID", 
            zap.Int("user_id", userID), 
            zap.Error(err))
        return "", err
    }

    // Получаем роль пользователя
    role, err := h.rolesRepo.GetRoleByID(ctx, user.RoleID)
    if err != nil {
        logger.Error("Failed to get user role", 
            zap.Int("user_id", userID), 
            zap.Error(err))
        return "", err
    }

    // Создаем JWT токен с ролью и ID пользователя
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":     strconv.Itoa(userID),
        "role":    role.Name,
        "role_id": roleID, // Добавлено role_id для более удобной проверки
        "exp":     time.Now().Add(time.Hour * 1).Unix(), // Время истечения токена — 1 час
    })

    logger.Debug("JWT token generated", zap.Int("user_id", userID), zap.String("role", role.Name))

    // Подписываем и возвращаем токен
    return token.SignedString([]byte(config.Config.JwtSecretKey))
}

func generateRefreshToken(userID int) (string, error) {
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

// checkPasswordHash — проверяет правильность пароля, сравнивая его с хешом
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}