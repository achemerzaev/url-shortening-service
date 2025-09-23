package main

import (
	"github.com/boretsotets/url-shortening-service/database"
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/internal/redisrepo"
	"github.com/boretsotets/url-shortening-service/internal/service"
	"github.com/boretsotets/url-shortening-service/internal/handler"
	"github.com/boretsotets/url-shortening-service/internal/middleware"


	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"context"
)

// кеширование шорткодов
// мидлвер для логера, рейт лимитинга
// лучше коды возвратов на обработке ошибок
// докеры
// ci/cd
// деплой
// прометеус и графана мб

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	logger.Info("app started", zap.String("env", "dev"))

	ctx := context.Background()
	dsn := "postgres://postgres:secret@localhost:5432/postgres?sslmode=disable"
	// localhost to be changed to contanier name
	pool, err := database.InitDb(ctx, dsn)
	if err != nil {
		logger.Fatal("Postgres connection error: ", zap.Error(err))
	}
	defer pool.Close()

	redisClient, err := database.InitRedis("localhost:6379", "", 0)
	if err != nil {
		logger.Fatal("Redis connection error: ", zap.Error(err))
	}

	redisRepo := redisrepo.NewRedisRepository(redisClient, logger)

	urlRepo := repository.NewUrlRepository(pool, logger)
	urlService := service.NewUrlService(urlRepo, redisRepo, logger)
	urlHandler := handler.NewUrlHandler(urlService, logger)

	userRepo := repository.NewUserRepository(pool, logger)
	userService := service.NewUserService(userRepo, redisRepo, logger)
	userHandler := handler.NewUserHandler(userService, logger)


	router := gin.New()
	router.Use(middleware.RequestIdMiddleware())
	router.Use(middleware.LoggerMiddleware(logger))

	router.POST("/shorten", urlHandler.HandlerPost)
	router.GET("/shorten/:shortcode", urlHandler.HandlerGet)
	router.PUT("/shorten/:shortcode", urlHandler.HandlerPut)
	router.DELETE("/shorten/:shortcode", urlHandler.HandlerDelete)
	router.GET("/shorten/:shortcode/stats", urlHandler.HandlerGetStats)

	router.POST("/register", userHandler.HandlerRegister)
	router.POST("/login", userHandler.HandlerLogin)
	router.POST("/refresh", userHandler.HandlerRefresh)

	router.Run("localhost:8080")
}