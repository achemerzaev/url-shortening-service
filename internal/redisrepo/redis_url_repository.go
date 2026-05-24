package redisrepo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type RedisURLRepository struct {
	client *redis.Client
	logger logger.Logger
}

func NewRedisURLRepository(client *redis.Client, logger logger.Logger) *RedisURLRepository {
	return &RedisURLRepository{client: client, logger: logger}
}

const shortURLKeyPrefix = "short:"

func shortURLKey(code string) string {
	return shortURLKeyPrefix + code
}

func (r *RedisURLRepository) SaveUrl(ctx context.Context, data models.URLInfo) error {
	err := r.client.HSet(ctx, shortURLKey(data.ShortCode), map[string]interface{}{
		"id":           data.Id,
		"url":          data.Url,
		"created_at":   data.CreatedAt,
		"updated_at":   data.UpdatedAt,
		"access_count": data.AccessCount,
		"owner_id":     data.OwnerID,
	}).Err()
	return err
}

func (r *RedisURLRepository) GetUrl(ctx context.Context, shortCode string, ownerID int) (string, error) {
	res, err := r.client.HGetAll(ctx, shortURLKey(shortCode)).Result()

	if err != nil {
		return "", fmt.Errorf("unexpected error getting url from Redis: %w", err)
	}
	if len(res) == 0 {
		return "", appErr.ErrNotFound
	}

	resint, err := strconv.Atoi(res["owner_id"])
	if err != nil || resint != ownerID {
		return "", appErr.ErrForbidden
	}

	return res["url"], nil
}

func (r *RedisURLRepository) IncrementCounter(ctx context.Context, shortCode string) error {
	err := r.client.HIncrBy(ctx, shortURLKey(shortCode), "access_count", 1).Err()
	if err != nil {
		return fmt.Errorf("unexpected error incrementing access counter in Redis: %w", err)
	}
	return nil
}

func (r *RedisURLRepository) GetUrlStats(ctx context.Context, shortCode string, ownerID int) (models.URLInfo, error) {
	result, err := r.client.HGetAll(ctx, shortURLKey(shortCode)).Result()
	var data models.URLInfo
	if err != nil {
		return data, fmt.Errorf("unexpected error getting url from Redis: %w", err)
	}
	if len(result) == 0 {
		return data, appErr.ErrNotFound
	}

	data.OwnerID, err = strconv.Atoi(result["owner_id"])
	if err != nil {
		return data, err
	}
	if data.OwnerID != ownerID {
		return data, appErr.ErrForbidden
	}

	data.Id, err = strconv.Atoi(result["id"])
	if err != nil {
		return data, err
	}

	data.Url = result["url"]
	data.ShortCode = shortCode
	data.CreatedAt, err = time.Parse(time.RFC3339Nano, result["created_at"])
	if err != nil {
		return data, err
	}

	data.UpdatedAt, err = time.Parse(time.RFC3339Nano, result["updated_at"])
	if err != nil {
		return data, err
	}

	data.AccessCount, err = strconv.Atoi(result["access_count"])
	if err != nil {
		return data, err
	}
	return data, nil
}

func (r *RedisURLRepository) UpdateUrl(ctx context.Context, requestedCode, newlongURL string, updatedAt time.Time, ownerID int) (models.URLInfo, error) {
	var data models.URLInfo
	res, err := r.client.HGet(ctx, shortURLKey(requestedCode), "owner_id").Result()
	if err != nil {
		if err == redis.Nil {
			return data, appErr.ErrNotFound
		}
		return data, fmt.Errorf("unexpected error getting url from Redis: %w", err)
	}
	resint, err := strconv.Atoi(res)
	if err != nil || resint != ownerID {
		return data, appErr.ErrForbidden
	}
	err = r.client.HSet(ctx, shortURLKey(requestedCode), "url", newlongURL).Err()
	if err != nil {
		return data, fmt.Errorf("unexpected error updating url in Redis: %w", err)
	}
	err = r.client.HSet(ctx, shortURLKey(requestedCode), "updated_at", updatedAt).Err()
	if err != nil {
		return data, fmt.Errorf("unexpected error setting new 'updated_at' in Redis: %w", err)
	}
	data, err = r.GetUrlStats(ctx, requestedCode, ownerID)
	if err != nil {
		return data, fmt.Errorf("unexpected error getting url from Redis: %w", err)
	}
	return data, nil
}

func (r *RedisURLRepository) DeleteUrl(ctx context.Context, shortCode string, ownerID int) error {
	res, err := r.client.HGet(ctx, shortURLKey(shortCode), "owner_id").Result()
	if err != nil {
		if err == redis.Nil {
			return appErr.ErrNotFound
		}
		return fmt.Errorf("unexpected error getting url from Redis: %w", err)
	}
	resint, err := strconv.Atoi(res)
	if err != nil || resint != ownerID {
		return appErr.ErrForbidden
	}
	err = r.client.Del(ctx, shortURLKey(shortCode)).Err()
	if err != nil {
		return fmt.Errorf("unexpected error deletig url from Redis: %w", err)
	}
	return nil
}
