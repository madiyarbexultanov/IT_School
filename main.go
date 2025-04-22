package main

import (
	"context"
	"it_school/config"
	"it_school/docs"
	"it_school/handlers"
	"it_school/logger"
	"it_school/middlewares"
	"it_school/repositories"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"

	"github.com/gin-contrib/cors"

	ginzap "github.com/gin-contrib/zap"
	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	logger := logger.GetLogger()

	defer func() {
		if r := recover(); r != nil {
			logger.Error("Application crashed!", zap.Any("error", r))
		}
	}()

	r.Use(
		ginzap.Ginzap(logger, time.RFC3339, true),
		ginzap.RecoveryWithZap(logger, true),
	)

	corsConfig := cors.Config{
		AllowAllOrigins: true,
		AllowHeaders:    []string{"*"},
		AllowMethods:    []string{"*"},
	}

	r.Use(cors.New(corsConfig))

	logger.Info("Loading configuration...")
	err := loadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	logger.Info("Connecting to database...")
	conn, err := connectToDb()
	if err != nil {
		logger.Fatal("Database connection failed", zap.Error(err))
	}

	r.Use(func(c *gin.Context) {
		c.Set("db", conn)
		c.Next()
	})

	// Health-check для Railway
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Server is running!",
		})
	})

	AuthRepository := repositories.NewAuthRepository(conn)
	UsersRepository := repositories.NewRUsersRepository(conn)
	SessionsRepository := repositories.NewSessionsRepository(conn)
	RolesRepository := repositories.NewRoleRepository(conn)

	StudentsRepository := repositories.NewStudentsRepository(conn)
	LessonsRepository := repositories.NewLessonsRepository(conn)
	CourseRepository := repositories.NewCourseRepository(conn)
	StudentsHandlers := handlers.NewStudentsHandlers(StudentsRepository)
	LessonsHandlers := handlers.NewLessonsHandlers(LessonsRepository)

	authHandler := handlers.NewAuthHandler(UsersRepository, SessionsRepository, RolesRepository)
	UserHandler := handlers.NewUserHandlers(UsersRepository)
	resetPasswordHandler := handlers.NewResetPasswordHandler(AuthRepository, UsersRepository)

	// Маршруты для аутентификации
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/refresh", authHandler.Refresh)

		authGroup.POST("/reset-password", resetPasswordHandler.ResetPassword)
		authGroup.POST("/new-password", resetPasswordHandler.SetNewPassword)
	}

	// Приватные маршруты (требуют аутентификацию)
	privateRoutes := r.Group("/")
	privateRoutes.Use(middlewares.AuthMiddleware(SessionsRepository, UsersRepository, RolesRepository))

	settingsRoutes := privateRoutes.Group("/settings")
	settingsRoutes.Use(middlewares.PermissionMiddleware("access_settings"))

	// Роуты для работы со студентами внутри настроек
	settingsRoutes.POST("/students", StudentsHandlers.Create)
	settingsRoutes.GET("/students/:studentId", StudentsHandlers.FindById)
	settingsRoutes.PUT("/students/:studentId", StudentsHandlers.Update)
	settingsRoutes.GET("/students", StudentsHandlers.FindAll)
	settingsRoutes.DELETE("/students/:studentId", StudentsHandlers.Delete)

	// Роуты для работы с уроками внутри настроек
	settingsRoutes.POST("/lessons", LessonsHandlers.Create)
	settingsRoutes.GET("/lessons/:lessonsId", LessonsHandlers.FindById)
	settingsRoutes.GET("/lessons", LessonsHandlers.FindAll)
	settingsRoutes.PUT("/lessons/:lessonsId", LessonsHandlers.Update)
	settingsRoutes.DELETE("/lessons/:lessonsId", LessonsHandlers.Delete)

	// Роуты для работы с пользователями внутри настроек
	settingsRoutes.POST("/users", UserHandler.Create)
	settingsRoutes.GET("/users/:userId", UserHandler.FindById)
	settingsRoutes.GET("/users", UserHandler.FindAll)
	settingsRoutes.PUT("/users/:userId", UserHandler.Update)
	settingsRoutes.DELETE("/users/:userId", UserHandler.Delete)

	settingsRoutes.GET("/managers", UserHandler.FindManagers)

	settingsRoutes.GET("/curators", UserHandler.FindCurators)

	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", swagger.WrapHandler(swaggerfiles.Handler))

	logger.Info("Application starting...")
	for _, route := range r.Routes() {
		logger.Info("Registered route", zap.String("method", route.Method), zap.String("path", route.Path))
	}

	port := viper.GetString("PORT")
	if port == "" {
		port = "8081"
	}

	logger.Info("Starting on port:", zap.String("port", port))

	if err := r.Run("0.0.0.0:" + port); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

func loadConfig() error {
	// Указываем путь к .env файлу
	viper.SetConfigFile(".env")

	// Загружаем переменные из .env, если он есть (необязательно)
	_ = viper.ReadInConfig() // не падаем, если файла нет

	// Читаем переменные окружения (например, из Railway)
	viper.AutomaticEnv()

	// Мапим переменные в структуру
	var mapConfig config.MapConfig
	err := viper.Unmarshal(&mapConfig)
	if err != nil {
		return err
	}

	config.Config = &mapConfig
	return nil
}

func connectToDb() (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), "postgres://postgres:123456@localhost:5432/it_school")
	if err != nil {
		return nil, err
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
