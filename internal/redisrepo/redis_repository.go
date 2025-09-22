package redisrepo

import (
	"github.com/boretsotets/url-shortening-service/internal/models"

	"go.uber.org/zap"
	"github.com/redis/go-redis/v9"

	"time"
	"context"
	"strconv"
	"encoding/json"
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

func (r *RedisRepository) SaveUrl(ctx context.Context, data models.UrlInfo) (error) {
	err := r.client.HSet(ctx, "short:"+data.ShortCode, map[string]interface{}{
		"id": data.Id,
		"url": data.Url,
		"created_at": data.CreatedAt,
		"updated_at": data.UpdatedAt,
		"access_count": data.AccessCount+1,
		"owner_id": data.OwnerID,
	}).Err()
	return err
}

func (r *RedisRepository) GetUrl(ctx context.Context, shortCode string, ownerID int) (string, error) {
	res, err := r.client.HGet(ctx, "short:"+shortCode, "owner_id").Result()
	if err != nil {
		r.logger.Error("Error getting owner_id from redis", zap.Error(err))
		return "", err
	}
	resint, err := strconv.Atoi(res)
	if err != nil || resint != ownerID {
		r.logger.Error("Error checking owner_id equality", zap.Error(err))
		return "", err
	}
	err = r.client.HIncrBy(ctx, "short:"+shortCode, "access_count", 1).Err()
	if err != nil {
		r.logger.Error("Error incrimenting access_count in redis", zap.Error(err))
		return "", err
	}
	res, err = r.client.HGet(ctx, "short:"+shortCode, "long_url").Result()
	if err != nil || len(res) == 0 {
		r.logger.Error("Error getting long_url from redis", zap.Error(err))
		return "", err
	}

	return string(res), nil
}

func (r *RedisRepository) GetUrlStats(ctx context.Context, shortCode string, ownerID int) (models.UrlInfo, error) {
	result, err := r.client.Get(ctx, "short:"+shortCode).Result()
	var data models.UrlInfo
	if err != nil || len(result) == 0 {
		r.logger.Error("Error getting url stats from redis", zap.Error(err))
		return data, err
	}
	json.Unmarshal([]byte(result), &data)

	if data.Id != ownerID {
		r.logger.Error("Clinet don't have access to this task", zap.Error(err))
		return data, err
	}
	return data, nil
}

func (r *RedisRepository) UpdateUrl(ctx context.Context, requestedCode, newLongUrl string, updatedAt time.Time, ownerID int) (models.UrlInfo, error) {
	var data models.UrlInfo
	res, err := r.client.HGet(ctx, "short:"+requestedCode, "owner_id").Result()
	if err != nil {
		r.logger.Error("Error getting owner_id from redis", zap.Error(err))
		return data, err
	}
	resint, err := strconv.Atoi(res)
	if err != nil || resint != ownerID {
		r.logger.Error("Error checking owner_id equality", zap.Error(err))
		return data, err
	}
	err = r.client.HSet(ctx, "short:"+requestedCode, "url", newLongUrl).Err()
	if err != nil {
		r.logger.Error("Error getting url from redis", zap.Error(err))
		return data, err
	}
	err = r.client.HSet(ctx, "short:"+requestedCode, "updated_at", updatedAt).Err()
	if err != nil {
		r.logger.Error("Error changing update time in redis", zap.Error(err))
		return data, err
	}
	data, err = r.GetUrlStats(ctx, requestedCode, ownerID)
	if err != nil {
		r.logger.Error("Error getting stats from redis", zap.Error(err))
		return data, err
	}
	return data, nil
}

func (r *RedisRepository) DeleteUrl(ctx context.Context, shortCode string, ownerID int) (error) {
	res, err := r.client.HGet(ctx, "short:"+shortCode, "owner_id").Result()
	if err != nil {
		r.logger.Error("Error getting owner_id from redis", zap.Error(err))
		return err
	}
	resint, err := strconv.Atoi(res)
	if err != nil || resint != ownerID {
		r.logger.Error("Error checking owner_id equality", zap.Error(err))
		return err
	}
	err = r.client.HDel(ctx, "short:"+shortCode).Err()
	return err
}
