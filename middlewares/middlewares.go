package middlewares

import (
	"it_school/config"
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func AuthMiddleware(sessionsRepo *repositories.SessionsRepository, usersRepo *repositories.UsersRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logger.GetLogger()
		authHeader := c.GetHeader("Authorization")

		var userID int
		var isSessionAuth bool

		if authHeader != "" {
			// JWT аутентификация
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(config.Config.JwtSecretKey), nil
			})

			if err != nil || !token.Valid {
				logger.Warn("Invalid token", zap.Error(err))
				c.JSON(http.StatusUnauthorized, models.NewApiError("invalid token"))
				c.Abort()
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.Warn("Invalid token claims")
				c.JSON(http.StatusUnauthorized, models.NewApiError("invalid token claims"))
				c.Abort()
				return
			}

			subject, ok := claims["sub"].(string)
			if !ok {
				logger.Warn("Invalid subject in token")
				c.JSON(http.StatusUnauthorized, models.NewApiError("invalid token subject"))
				c.Abort()
				return
			}

			userID, err = strconv.Atoi(subject)
			if err != nil {
				logger.Warn("Invalid user ID format in token")
				c.JSON(http.StatusUnauthorized, models.NewApiError("invalid user ID format"))
				c.Abort()
				return
			}
		} else {
			// Сессионная аутентификация
			isSessionAuth = true
			sessionToken, err := c.Cookie("session_token")
			if err != nil {
				logger.Warn("No session token found", zap.Error(err))
				c.JSON(http.StatusUnauthorized, models.NewApiError("no session token"))
				c.Abort()
				return
			}

			session, err := sessionsRepo.GetSession(c.Request.Context(), sessionToken)
			if err != nil {
				logger.Warn("Invalid session token", zap.Error(err))
				c.JSON(http.StatusUnauthorized, models.NewApiError("invalid session token"))
				c.Abort()
				return
			}
			userID = session.UserID
		}

		// Проверяем существование пользователя
		_, err := usersRepo.FindById(c.Request.Context(), userID)
		if err != nil {
			logger.Warn("User not found", zap.Error(err))
			c.JSON(http.StatusUnauthorized, models.NewApiError("user not found"))
			c.Abort()
			return
		}

		// Сохраняем данные в контексте
		c.Set("userID", userID)
		c.Set("isSessionAuth", isSessionAuth)

		logger.Info("User authenticated", 
			zap.Int("userID", userID),
			zap.Bool("isSessionAuth", isSessionAuth))
		
		c.Next()
	}
}