package syncer

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"context"
	"strings"
)

func SyncRedisToPostgres(ctx context.Context, rdb *redis.Client, db *pgxpool.Pool) error {
	const pattern = "short:*"

	iter := rdb.Scan(ctx, 0, pattern, 0).Iterator()

	updates := make([]struct {
		ShortCode   string
		AccessCount string
	}, 0, 100)

	for iter.Next(ctx) {
		key := iter.Val()
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			continue
		}
		urlCode := parts[1]

		count, err := rdb.HGet(ctx, key, "access_count").Result()
		if err != nil {
			continue
		}

		updates = append(updates, struct {
			ShortCode   string
			AccessCount string
		}{urlCode, count})
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(updates) == 0 {
		return nil
	}

	// build query string
	var sb strings.Builder

	sb.WriteString("UPDATE urls SET AccessCount = CASE ShortCode ")

	for _, value := range updates {
		sb.WriteString("WHEN '")
		sb.WriteString(value.ShortCode)
		sb.WriteString("' THEN ")
		sb.WriteString(value.AccessCount)
		sb.WriteByte(' ')
	}

	sb.WriteString("END WHERE ShortCode IN (")

	i := 0
	for _, value := range updates {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('\'')
		sb.WriteString(value.ShortCode)
		sb.WriteByte('\'')
		i++
	}

	sb.WriteByte(')')
	query := sb.String()

	if _, err := db.Exec(ctx, query); err != nil {
		return err
	}

	return nil

}
