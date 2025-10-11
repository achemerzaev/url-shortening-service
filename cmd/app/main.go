package main

import (
	"github.com/boretsotets/url-shortening-service/database"
	"github.com/boretsotets/url-shortening-service/internal/handler"
	"github.com/boretsotets/url-shortening-service/internal/middleware"
	"github.com/boretsotets/url-shortening-service/internal/redisrepo"
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/internal/service"
	"github.com/boretsotets/url-shortening-service/internal/configLoader"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"context"
	"strconv"
	"log"
	"fmt"
	"os"
)


// проверка билда с секретами и внутренней сетю, разделение конфигов и докер компоуза на девелопмент и релиз
// сделат мейн и хендлеры тонкими
// деплой, сертификат
// ci/cd

func main() {
	// starting logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("error creating logger: %s", err)
	}

	// starting syncer
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("error syncing logger: %s", err)
		}
	}()

	logger.Info("app started", zap.String("env", "dev"))

	// connecting to database
	ctx := context.Background()
	dbName := configLoader.MustReadSecret("/run/secrets/db_name")
	dbUser := configLoader.MustReadSecret("/run/secrets/db_user")
	dbPassword := configLoader.MustReadSecret("/run/secrets/db_password")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, host, port, dbName)

	pool, err := database.InitDb(ctx, dsn)
	if err != nil {
		logger.Fatal("Postgres connection error: ", zap.Error(err))
	}
	defer pool.Close()

	// connecting to redis
	
	redisPassword := configLoader.MustReadSecret("/run/secrets/redis_password")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	redisClient, err := database.InitRedis(redisAddr, redisPassword, redisDB)
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

	// setting up a router
	router := gin.New()
	router.Use(middleware.RequestIdMiddleware())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.PrometheusMiddleware())
	router.Use(middleware.JSONValidationMiddleware())

	router.POST("/register", userHandler.HandlerRegister)
	router.POST("/login", userHandler.HandlerLogin)
	router.POST("/refresh", userHandler.HandlerRefresh)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))


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