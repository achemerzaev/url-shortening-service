package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/achemerzaev/url-shortening-service/database"
	"github.com/achemerzaev/url-shortening-service/internal/config"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"log"
)

func Run() {
	// 1. Загружаем конфиг
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to read config: %v", err.Error())
	}

	// 2. Инициализируем логгер
	logger, err := logger.New(cfg.Log.Level)
	if err != nil {
		log.Fatalf("error creating logger: %s", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("error syncing logger: %s", err)
		}
	}()

	logger.Info("app started")

	// 4. Подключаем Postgres
	pgPool, err := database.NewPostgres(cfg.DB)
	if err != nil {
		logger.Fatal("Postgres connection error: ", err)
	}
	defer pgPool.Close()

	// 5. Подключаем Redis
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("Redis connection error: ", err)
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			logger.Error("Redis closing error: ", err)
		}
	}()

	// 6. Создаем и связываем слои
	handlers := BuildHandlers(pgPool, rdb, logger)

	// 7. Настраиваем роутер
	router := SetupRouter(handlers, logger)

	srv := &http.Server{
		Addr:         cfg.App.Addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Router starting error: ", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: ", err)
	}

	logger.Info("Server exited")
}
