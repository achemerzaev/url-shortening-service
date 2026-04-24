package redisrepo

import (
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"github.com/redis/go-redis/v9"

	"context"
	"time"
)

type RedisUserRepository struct {
	client *redis.Client
	logger logger.Logger
}

func NewRedisUserRepository(client *redis.Client, logger logger.Logger) *RedisUserRepository {
	return &RedisUserRepository{client: client, logger: logger}
}

func (r *RedisUserRepository) SaveRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	return r.client.Set(ctx, "refresh:"+userID, token, ttl).Err()
}

func (r *RedisUserRepository) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	return r.client.Get(ctx, "refresh:"+userID).Result()
}
