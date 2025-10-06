package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/boretsotets/url-shortening-service/internal/syncer"

	"github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
			if err := syncer.SyncRedisToPostgres(ctx, rdb, pool); err != nil {
				fmt.Println("sync error:", err)
			}
		}

}