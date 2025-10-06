package main

import (
	"github.com/boretsotets/url-shortening-service/database"
	"github.com/boretsotets/url-shortening-service/internal/handler"
	"github.com/boretsotets/url-shortening-service/internal/middleware"
	"github.com/boretsotets/url-shortening-service/internal/redisrepo"
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"context"
	"log"
	"os"
)

// докеры
// ci/cd
// деплой
// прометеус и графана мб

// проверка ошибки синтакса на синке
// проверка инкремента аксесс каунта

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("error creating logger: %s", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("error syncing logger: %s", err)
		}
	}()

	logger.Info("app started", zap.String("env", "dev"))

	ctx := context.Background()
	pool, err := database.InitDb(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatal("Postgres connection error: ", zap.Error(err))
	}
	defer pool.Close()

	redisClient, err := database.InitRedis(os.Getenv("REDIS_ADDR"), "", 0)
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
	router.Use(middleware.JSONValidationMiddleware())

	router.POST("/register", userHandler.HandlerRegister)
	router.POST("/login", userHandler.HandlerLogin)
	router.POST("/refresh", userHandler.HandlerRefresh)


	private := router.Group("/")
	private.Use(middleware.AuthorizationMiddleware(logger))
	{
		private.POST("/shorten", urlHandler.HandlerPost)
		private.GET("/shorten/:shortcode", urlHandler.HandlerGet)
		private.PUT("/shorten/:shortcode", urlHandler.HandlerPut)
		private.DELETE("/shorten/:shortcode", urlHandler.HandlerDelete)
		private.GET("/shorten/:shortcode/stats", urlHandler.HandlerGetStats)
	}

	err = router.Run(":8080")
	if err != nil {
		logger.Fatal("Router starting error: ", zap.Error(err))
	}
}
