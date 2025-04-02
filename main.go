package main

import (
	"context"
	"students/config"
	"students/docs"
	"students/handlers"
	"students/logger"
	"students/repositories"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

func main() {
	r := gin.Default()

	logger := logger.GetLogger()
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

	err := loadConfig()
	if err != nil {
		panic(err)
	}

	conn, err := connectToDb()
	if err != nil {
		panic(err)
	}

	StudentsRepository := repositories.NewStudentsRepository(conn)
	LessonsRepository := repositories.NewLessonsRepository(conn)
	StudentsHandlers := handlers.NewStudentsHandlers(StudentsRepository)
	LessonsHandlers := handlers.NewLessonsHandlers(LessonsRepository)

	// authorized := r.Group("")
	// authorized.Use(middlewares.AuthMiddleware)

	// admin := r.Group("/admin")
	// admin.Use(middlewares.AuthMiddleware)

	//endpoints

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
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", swagger.WrapHandler(swaggerfiles.Handler))

	logger.Info("Application starting...")
	r.Run(config.Config.AppHost)
}
func loadConfig() error {
	viper.SetConfigFile(".env")
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
