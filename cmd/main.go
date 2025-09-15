package main

import (
	"github.com/boretsotets/url-shortening-service/database"
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/internal/service"
	"github.com/boretsotets/url-shortening-service/internal/handler"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"context"
)

// ошибка на добавлении времени в таймстампы

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	logger.Info("app started", zap.String("env", "dev"))

	ctx := context.Background()
	dsn := "postgres://postgres:secret@localhost:5432/postgres?sslmode=disable"
	// localhost to be changed to contanier name
	pool, err := database.InitDb(ctx, dsn)
	if err != nil {
		logger.Fatal("cannot connect to database")
	}
	defer pool.Close()

	urlRepo := repository.NewUrlRepository(pool, logger)
	urlService := service.NewUrlService(urlRepo, logger)
	urlHandler := handler.NewUrlHandler(urlService, logger)


	router := gin.Default()

	router.POST("/shorten", urlHandler.HandlerPost)
	router.GET("/shorten/:shortcode", urlHandler.HandlerGet)
	router.PUT("/shorten/:shortcode", urlHandler.HandlerPut)
	router.DELETE("/shorten/:shortcode", urlHandler.HandlerDelete)
	router.GET("/shorten/:shortcode/stats", urlHandler.HandlerGetStats)

	router.Run("localhost:8080")
}