package main

import (
	"github.com/boretsotets/url-shortening-service/database"
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/internal/service"
	"github.com/boretsotets/url-shortening-service/internal/handler"

	"github.com/gin-gonic/gin"

	"context"
	"log"
	"os"
)

func main() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
    	log.Fatal("Failed to open log file:", err)
	}
		log.SetOutput(file)

	ctx := context.Background()
	dsn := "postgres://postgres:secret@localhost:5432/postgres?sslmode=disable"
	// localhost to be changed to contanier name
	pool, err := database.InitDb(ctx, dsn)
	if err != nil {
		log.Fatal("cannot connect to database")
	}
	defer pool.Close()

	urlRepo := repository.NewUrlRepository(pool)
	urlService := service.NewUrlService(urlRepo)
	urlHandler := handler.NewUrlHandler(urlService)


	router := gin.Default()

	router.POST("/shorten", urlHandler.HandlerPost)

	router.Run("localhost:8080")
}