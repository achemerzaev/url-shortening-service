package redisrepo

import (
	"go.uber.org/zap"

	"github.com/redis/go-redis/v9"

	"time"
	"context"
)

type RedisRepository struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisRepository(client *redis.Client, logger *zap.Logger) *RedisRepository {
	return &RedisRepository{client: client, logger: logger}
}

func (r *RedisRepository) SaveRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	return r.client.Set(ctx, "refresh:"+userID, token, ttl).Err()
}

func (r *RedisRepository) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	return r.client.Get(ctx, "refresh:"+userID).Result()
}