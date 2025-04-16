package main

import (
	"context"
	"it_school/config"
	"it_school/docs"
	"it_school/handlers"
	"it_school/logger"
	"it_school/middlewares"
	"it_school/repositories"
	"os"
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

	usersRepository := repositories.NewUsersRepository(conn)
	SessionsRepository := repositories.NewSessionsRepository(conn)
	RolesRepository := repositories.NewRoleRepository(conn)

	StudentsRepository := repositories.NewStudentsRepository(conn)
	LessonsRepository := repositories.NewLessonsRepository(conn)
	StudentsHandlers := handlers.NewStudentsHandlers(StudentsRepository)
	LessonsHandlers := handlers.NewLessonsHandlers(LessonsRepository)

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

	//http://localhost:8081/students/
	r.POST("/students", StudentsHandlers.Create)
	r.GET("/students/:studentId", StudentsHandlers.FindById)
	r.PUT("/students/:studentId", StudentsHandlers.Update)
	r.GET("/students", StudentsHandlers.FindAll)
	r.DELETE("/students/:studentId", StudentsHandlers.Delete)

	//http://localhost:8081/lessons/
	r.POST("/lessons", LessonsHandlers.Create)
	r.GET("/lessons/:lessonsId", LessonsHandlers.FindById)
	r.GET("/lessons", LessonsHandlers.FindAll)
	r.PUT("/lessons/:lessonsId", LessonsHandlers.Update)
	r.DELETE("/lessons/:lessonsId", LessonsHandlers.Delete)

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // или значение из переменной Railway
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
