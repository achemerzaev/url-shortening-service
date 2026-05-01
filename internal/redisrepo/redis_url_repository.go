package redisrepo

import (
	"context"
	"errors"
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

func (r *RedisURLRepository) SaveUrl(ctx context.Context, data models.UrlInfo) error {
	err := r.client.HSet(ctx, "short:"+data.ShortCode, map[string]interface{}{
		"id":           data.Id,
		"url":          data.Url,
		"created_at":   data.CreatedAt,
		"updated_at":   data.UpdatedAt,
		"access_count": data.AccessCount + 1,
		"owner_id":     data.OwnerID,
	}).Err()
	return err
}

func (r *RedisURLRepository) GetUrl(ctx context.Context, shortCode string, ownerID int) (string, error) {
	res, err := r.client.HGetAll(ctx, "short:"+shortCode).Result()

	if err != nil {
		return "", err
	}

	resint, err := strconv.Atoi(res["owner_id"])
	if err != nil || resint != ownerID {
		return "", errors.New("user doesn't own this resource")
	}
	err = r.client.HIncrBy(ctx, "short:"+shortCode, "access_count", 1).Err()
	if err != nil {
		return "", err
	}

	return res["url"], nil
}

func (r *RedisURLRepository) GetUrlStats(ctx context.Context, shortCode string, ownerID int) (models.UrlInfo, error) {
	result, err := r.client.HGetAll(ctx, "short:"+shortCode).Result()
	var data models.UrlInfo
	if err != nil {
		return data, err
	}

	if len(result) == 0 {
		return data, errors.New("no rows in redis")
	}

	data.OwnerID, _ = strconv.Atoi(result["owner_id"])
	if data.OwnerID != ownerID {
		return data, errors.New("user doesn't own this resource")
	}
	data.Id, _ = strconv.Atoi(result["id"])
	data.Url = result["url"]
	data.ShortCode = shortCode
	data.CreatedAt, _ = time.Parse(time.RFC3339Nano, result["created_at"])
	data.UpdatedAt, _ = time.Parse(time.RFC3339Nano, result["updated_at"])
	data.AccessCount, _ = strconv.Atoi(result["access_count"])
	return data, nil
}

func (r *RedisURLRepository) UpdateUrl(ctx context.Context, requestedCode, newLongUrl string, updatedAt time.Time, ownerID int) (models.UrlInfo, error) {
	var data models.UrlInfo
	res, err := r.client.HGet(ctx, "short:"+requestedCode, "owner_id").Result()
	if err != nil {
		return data, appErr.ErrNotFound
	}
	resint, err := strconv.Atoi(res)
	if err != nil || resint != ownerID {
		return data, appErr.ErrForbidden
	}
	err = r.client.HSet(ctx, "short:"+requestedCode, "url", newLongUrl).Err()
	if err != nil {
		return data, err
	}
	err = r.client.HSet(ctx, "short:"+requestedCode, "updated_at", updatedAt).Err()
	if err != nil {
		return data, err
	}
	data, err = r.GetUrlStats(ctx, requestedCode, ownerID)
	if err != nil {
		return data, err
	}
	return data, nil
}

func (r *RedisURLRepository) DeleteUrl(ctx context.Context, shortCode string, ownerID int) error {
	res, err := r.client.HGet(ctx, "short:"+shortCode, "owner_id").Result()
	if err != nil {
		return appErr.ErrNotFound
	}
	resint, err := strconv.Atoi(res)
	if err != nil || resint != ownerID {
		return appErr.ErrForbidden
	}
	err = r.client.Del(ctx, "short:"+shortCode).Err()
	return err
}
