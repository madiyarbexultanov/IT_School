package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"it_school/config"
	"it_school/models"
	"it_school/repositories"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthHandler struct {
	usersRepo    *repositories.UsersRepository
	sessionsRepo *repositories.SessionsRepository
}

func NewAuthHandler(usersRepo *repositories.UsersRepository, sessionsRepo *repositories.SessionsRepository) *AuthHandler {
	return &AuthHandler{
		usersRepo:    usersRepo,
		sessionsRepo: sessionsRepo,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	// Получаем пользователя из репозитория
	user, err := h.usersRepo.FindByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewApiError("invalid credentials"))
		return
	}

	// Проверяем пароль
	if !checkPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, models.NewApiError("invalid credentials"))
		return
	}

	// Генерируем JWT токен
	token, err := h.generateJWTToken(user.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate token"))
		return
	}

	// Создаем сессию
	refreshToken, err := generateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate refresh token"))
		return
	}

	session := models.Session{
		UserID:       user.Id,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 7), // 7 дней
	}

	if err := h.sessionsRepo.CreateSession(c.Request.Context(), session); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("failed to create session"))
		return
	}

	// Устанавливаем cookie
	c.SetCookie("session_token", refreshToken, int(session.ExpiresAt.Unix()), "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"expires": time.Now().Add(time.Hour * 1).Unix(), // JWT на 1 час
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("no session token"))
		return
	}

	if err := h.sessionsRepo.DeleteSession(c.Request.Context(), sessionToken); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("failed to delete session"))
		return
	}

	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewApiError("no session token"))
		return
	}

	session, err := h.sessionsRepo.GetSession(c.Request.Context(), sessionToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewApiError("invalid session token"))
		return
	}

	token, err := h.generateJWTToken(session.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("failed to generate token"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"expires": time.Now().Add(time.Hour * 1).Unix(),
	})
}

func (h *AuthHandler) generateJWTToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": strconv.Itoa(userID),
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	})
	return token.SignedString([]byte(config.Config.JwtSecretKey))
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func checkPasswordHash(password, hash string) bool {
	// Реализация проверки пароля (например, bcrypt)
	// return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	return true // Заглушка - замените на реальную проверку
}