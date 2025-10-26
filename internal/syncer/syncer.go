package syncer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/boretsotets/url-shortening-service/database"
	"github.com/boretsotets/url-shortening-service/internal/config"
	"github.com/boretsotets/url-shortening-service/pkg/logger"
)

func Run() {
	ctx := context.Background()

	// Загружаем конфиг
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	// Запускаем логгер
	logger, err := logger.New(cfg.Log.Level)
	if err != nil {
		log.Fatalf("error creating logger: %s", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("error syncing logger: %s", err)
		}
	}()

	// Подключаем Postgres
	pgPool, err := database.NewPostgres(cfg.DB)
	if err != nil {
		logger.Fatal("Postgres connection error: ", err)
	}
	defer pgPool.Close()

	// Подключаем Redis
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("Redis connection error: ", err)
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			logger.Error("Redis closing error: ", err)
		}
	}()

	// Настраиваем таймер
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := SyncRedisToPostgres(ctx, rdb, pgPool); err != nil {
			fmt.Println("sync error:", err)
		}
	}

}
