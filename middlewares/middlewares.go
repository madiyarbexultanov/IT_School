package middlewares

import (
	"it_school/config"
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthMiddleware — middleware для аутентификации пользователя. Поддерживает как JWT, так и сессионную аутентификацию.
func AuthMiddleware(sessionsRepo *repositories.SessionsRepository, usersRepo *repositories.UsersRepository, rolesRepo *repositories.RoleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logger.GetLogger()

		// Извлекаем заголовок Authorization из запроса
		authHeader := c.GetHeader("Authorization")

		var userID uuid.UUID
		var isSessionAuth bool

		// Если Authorization header присутствует, пробуем аутентифицировать через JWT
		if authHeader != "" {
			// Извлекаем токен, удаляя префикс "Bearer "
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Разбираем JWT токен, используя секретный ключ
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(config.Config.JwtSecretKey), nil
			})

			// Если токен невалиден, возвращаем ошибку
			if err != nil || !token.Valid {
				logger.Warn("Invalid token", zap.Error(err)) // Логируем предупреждение
				c.JSON(http.StatusUnauthorized, models.NewApiError("invalid token"))
				c.Abort() // Прерываем выполнение дальнейших middleware
				return
			}

			// Извлекаем claims (данные из токена)
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.Warn("Invalid token claims")
				c.JSON(http.StatusUnauthorized, models.NewApiError("invalid token claims"))
				c.Abort()
				return
			}

			// Извлекаем идентификатор пользователя из токена (subject)
			subject, ok := claims["sub"].(string)
			if !ok {
				logger.Warn("Invalid subject in token")
				c.JSON(http.StatusUnauthorized, models.NewApiError("invalid token subject"))
				c.Abort()
				return
			}

			// Преобразуем строку в int для идентификатора пользователя
			userID, err = uuid.Parse(subject)
			if err != nil {
				logger.Warn("Invalid user ID format in token")
				c.JSON(http.StatusUnauthorized, models.NewApiError("invalid user ID format"))
				c.Abort()
				return
			}
		} else {
			// Если токена нет, пробуем аутентификацию через сессии
			isSessionAuth = true
			sessionToken, err := c.Cookie("session_token")
			if err != nil {
				logger.Warn("No session token found", zap.Error(err))
				c.JSON(http.StatusUnauthorized, models.NewApiError("no session token"))
				c.Abort()
				return
			}

			// Проверяем валидность сессионного токена
			session, _, err := sessionsRepo.GetSession(c.Request.Context(), sessionToken)
			if err != nil {
				logger.Warn("Invalid session token", zap.Error(err))
				c.JSON(http.StatusUnauthorized, models.NewApiError("invalid session token"))
				c.Abort()
				return
			}
			userID = session.UserID // Извлекаем ID пользователя из сессии
		}

		// Теперь ищем пользователя в базе данных по полученному userID
		user, err := usersRepo.FindById(c.Request.Context(), userID)
		if err != nil {
			logger.Warn("User not found", zap.Error(err))
			c.JSON(http.StatusUnauthorized, models.NewApiError("user not found"))
			c.Abort()
			return
		}

		// Получаем роль пользователя из базы данных
		role, err := rolesRepo.GetRoleByID(c.Request.Context(), user.RoleID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, models.NewApiError("couldn't find role"))
            c.Abort()
            return
        }

		// Сохраняем информацию о пользователе в контексте для использования в последующих middleware
		c.Set("userID", userID)
		c.Set("userRole", role)
		c.Set("isSessionAuth", isSessionAuth)


		logger.Info("User authenticated", 
			zap.Any("userID", userID),
			zap.String("role", role.Name),
			zap.Bool("isSessionAuth", isSessionAuth))
		
		c.Next()
	}
}

// PermissionMiddleware — middleware для проверки наличия разрешений у пользователя на выполнение действия.
func PermissionMiddleware(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        logger := logger.GetLogger()
        
        // Извлекаем роль из контекста
        roleObj, exists := c.Get("userRole")
        if !exists {
            logger.Warn("Role missing - access denied")
            c.JSON(http.StatusForbidden, models.NewApiError("access denied"))
            c.Abort()
            return
        }

        // Проверяем, что роль имеет правильный тип
        role, ok := roleObj.(*models.Role)
        if !ok {
            logger.Error("Invalid role type in context")
            c.JSON(http.StatusInternalServerError, models.NewApiError("role parsing error"))
            c.Abort()
            return
        }

        // Проверяем, есть ли у роли нужные разрешения
        if !role.Permissions[permission] {
            logger.Warn("Permission denied",
                zap.String("role", role.Name),
                zap.String("need", permission))
            
            c.JSON(http.StatusForbidden, models.NewApiError("forbidden"))
            c.Abort()
            return
        }


        logger.Debug("Access granted")
        c.Next() 
    }
}