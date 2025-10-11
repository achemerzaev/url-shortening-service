package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/boretsotets/url-shortening-service/internal/syncer"
	"github.com/boretsotets/url-shortening-service/internal/configLoader"

	"github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"

	"strconv"
)

func main() {
	ctx := context.Background()
	dbName := configLoader.MustReadSecret("/run/secrets/db_name")
	dbUser := configLoader.MustReadSecret("/run/secrets/db_user")
	dbPassword := configLoader.MustReadSecret("/run/secrets/db_password")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, host, port, dbName)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}


	redisPassword := configLoader.MustReadSecret("/run/secrets/redis_password")
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		Password: redisPassword,
		DB: redisDB,
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
			if err := syncer.SyncRedisToPostgres(ctx, rdb, pool); err != nil {
				fmt.Println("sync error:", err)
			}
		}

}