package syncer

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"context"
	"strings"
)

func SyncRedisToPostgres(ctx context.Context, rdb *redis.Client, db *pgxpool.Pool) error {
	const pattern = "short:*"

	iter := rdb.Scan(ctx, 0, pattern, 0).Iterator()

	updatedCodes := make([]string, 0)
	updatedCounts := make([]int, 0)

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
		countInt, _ := strconv.Atoi(count)

		updatedCodes = append(updatedCodes, urlCode)
		updatedCounts = append(updatedCounts, countInt)
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(updatedCodes) == 0 {
		return nil
	}

	if _, err := db.Exec(ctx,
		`UPDATE urls AS u
		SET accesscount = v.access_count
		FROM unnest($1::text[], $2::int[]) AS v(shortcode, access_count)
		WHERE u.shortcode = v.shortcode;`, updatedCodes, updatedCounts); err != nil {
		return err
	}

	return nil

}
