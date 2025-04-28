package main

import (
	"context"
	"it_school/config"
	"it_school/docs"
	"it_school/handlers"
	"it_school/logger"
	"it_school/middlewares"
	"it_school/repositories"
	"it_school/utils"
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
	CuratorsRepository := repositories.NewCuratorsRepository(conn)
	CourseRepository := repositories.NewCourseRepository(conn)

	if err := utils.SeedAdminAndRoles(RolesRepository, UsersRepository); err != nil {
		logger.Fatal("Couldn't create admin", zap.Error(err))
	}

	StudentsRepository := repositories.NewStudentsRepository(conn)
	LessonsRepository := repositories.NewLessonsRepository(conn)
	StudentsHandlers := handlers.NewStudentsHandlers(StudentsRepository)
	LessonsHandlers := handlers.NewLessonsHandlers(LessonsRepository)
	CuratorsHandlers := handlers.NewCuratorsHandler(CuratorsRepository)
	CourseHandlers := handlers.NewCourseHandlers(CourseRepository)

	authHandler := handlers.NewAuthHandler(UsersRepository, SessionsRepository, RolesRepository)
	UserHandler := handlers.NewUserHandlers(UsersRepository, CuratorsRepository, RolesRepository)
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

	// Роуты настроек. Доступ имеет только Админ
	settingsRoutes := privateRoutes.Group("/settings")
	settingsRoutes.Use(middlewares.PermissionMiddleware("access_settings"))

	// Роуты для работы со студентами внутри настроек
	settingsRoutes.POST("/students", StudentsHandlers.Create)
	settingsRoutes.PUT("/students/:studentId", StudentsHandlers.Update)
	settingsRoutes.DELETE("/students/:studentId", StudentsHandlers.Delete)

	// Роуты для работы с курсами внутри настроек
	settingsRoutes.POST("/courses", CourseHandlers.Create)
	settingsRoutes.GET("/courses/:courseId", CourseHandlers.FindById)
	settingsRoutes.GET("/courses", CourseHandlers.FindAll)
	settingsRoutes.PUT("/courses/:courseId", CourseHandlers.Update)
	settingsRoutes.DELETE("/courses/:courseId", CourseHandlers.Delete)

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

	// Получение списков Менеджеров и Кураторов
	settingsRoutes.GET("/users/managers", UserHandler.FindManagers)
	settingsRoutes.GET("/users/curators", UserHandler.FindCurators)

	// Фунеции Куратора для работы со студентами и курсами
	curatorsRoutes := privateRoutes.Group("/curators")
	curatorsRoutes.Use(middlewares.PermissionMiddleware("access_curator"))
	{
		curatorsRoutes.POST("/add-student", CuratorsHandlers.AddStudent)
		curatorsRoutes.POST("/remove-student", CuratorsHandlers.RemoveStudent)
		curatorsRoutes.POST("/add-course", CuratorsHandlers.AddCourse)
		curatorsRoutes.POST("/remove-course", CuratorsHandlers.RemoveCourse)
	}

	// Функции Менеджера для просмотра студентов
	managerRoutes := privateRoutes.Group("/manager")
	managerRoutes.Use(middlewares.PermissionMiddleware("access_manager"))
	{
		managerRoutes.GET("/students", StudentsHandlers.FindAll)
		managerRoutes.GET("/students/:studentId", StudentsHandlers.FindById)
	}

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
    viper.SetConfigFile(".env")
    _ = viper.ReadInConfig()   
    viper.AutomaticEnv()      

    // 1) Сначала распакуем всё из конфиг-файла
    var cfg config.MapConfig
    if err := viper.Unmarshal(&cfg); err != nil {
        return err
    }

    // 2) А потом ДОСТАЁМ ЛЮБЫЕ ENV-ПЕРЕМЕННЫЕ прямо
    cfg.DbConnectionString = viper.GetString("DATABASE_URL")

    cfg.JwtSecretKey      	= viper.GetString("JWT_SECRET_KEY")
    cfg.JwtExpiresIn 		= viper.GetDuration("JWT_EXPIRE_DURATION")
    cfg.Initial_Password 	= viper.GetString("INITIAL_PASSWORD")
    cfg.Admin_Name 		= viper.GetString("ADMIN_NAME")
    cfg.Admin_Mail 		= viper.GetString("ADMIN_MAIL")
    cfg.Admin_Phone 		= viper.GetString("ADMIN_PHONE")
    cfg.SMTPPassword   	   	= viper.GetString("SMTP_PASSWORD")
    cfg.SMTPEmail	   	= viper.GetString("SMTP_EMAIL")
    cfg.SMTPHost           	= viper.GetString("SMTP_HOST")
    cfg.SMTPPort           	= viper.GetString("SMTP_PORT")
	
    config.Config = &cfg
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
