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
	gin.SetMode(gin.ReleaseMode)

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

	usersRepository := repositories.NewUsersRepository(conn)
	SessionsRepository := repositories.NewSessionsRepository(conn)
	RolesRepository := repositories.NewRoleRepository(conn)

	authHandler := handlers.NewAuthHandler(usersRepository, SessionsRepository, RolesRepository)
	resetPasswordHandler := handlers.NewResetPasswordHandler(usersRepository)

	// Маршруты для аутентификации
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/signup", authHandler.SignUp)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/refresh", authHandler.Refresh)

		authGroup.POST("/reset-password", resetPasswordHandler.ResetPassword)
		authGroup.POST("/new-password", resetPasswordHandler.SetNewPassword)
	}

	// Приватные маршруты (требуют аутентификацию)
	privateRoutes := r.Group("/")
	privateRoutes.Use(middlewares.AuthMiddleware(SessionsRepository, usersRepository, RolesRepository))

	// // Доступ к настройкам только у Директора
	// privateRoutes.GET("/settings", middlewares.PermissionMiddleware("access_settings"), _)

	// // Доступ к курсам только у Куратора
	// privateRoutes.GET("/courses", middlewares.PermissionMiddleware("access_courses"), _)

	// // Доступ к ученикам только у Куратора
	// privateRoutes.GET("/students", middlewares.PermissionMiddleware("access_students"), _)

	// // Доступ к урокам только у Менеджера
	// privateRoutes.GET("/lessons", middlewares.PermissionMiddleware("access_lessons"), _)


	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", swagger.WrapHandler(swaggerfiles.Handler))

	logger.Info("Application starting...")
	for _, route := range r.Routes() {
		logger.Info("Registered route", zap.String("method", route.Method), zap.String("path", route.Path))
	}

	if err := r.Run(config.Config.AppHost); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

func loadConfig() error {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	var mapConfig config.MapConfig
	err = viper.Unmarshal(&mapConfig)
	if err != nil {
		return err
	}

	config.Config = &mapConfig

	return nil
}

func connectToDb() (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), config.Config.DbConnectionString)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
