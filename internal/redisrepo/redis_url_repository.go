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
		if err != redis.Nil {
			// r.logger.Error("Error getting owner_id from redis", err)
		}
		return "", err
	}

	resint, err := strconv.Atoi(res["owner_id"])
	if err != nil || resint != ownerID {
		// r.logger.Error("User %s %d doesn't own this resource", res["owner_id"], ownerID, err)
		return "", errors.New("user doesn't own this resource")
	}
	err = r.client.HIncrBy(ctx, "short:"+shortCode, "access_count", 1).Err()
	if err != nil {
		// r.logger.Error("Error incrimenting access_count in redis", err)
		return "", err
	}

	return res["url"], nil
}

func (r *RedisURLRepository) GetUrlStats(ctx context.Context, shortCode string, ownerID int) (models.UrlInfo, error) {
	result, err := r.client.HGetAll(ctx, "short:"+shortCode).Result()
	var data models.UrlInfo
	if err != nil {
		// r.logger.Error("Error getting url stats from redis", err)
		return data, err
	}

	if len(result) == 0 {
		// r.logger.Error("No rows in result set", err)
		return data, errors.New("no rows in redis")
	}

	data.OwnerID, _ = strconv.Atoi(result["owner_id"])
	if data.OwnerID != ownerID {
		// r.logger.Error("User doesn't own this resource", err)
		return data, errors.New("user doesn't own this resource")
	}
	data.Id, _ = strconv.Atoi(result["id"])
	data.Url = result["url"]
	data.ShortCode = shortCode
	data.CreatedAt, _ = time.Parse(time.RFC3339Nano, result["created_at"])
	// if err != nil {
	// 	r.logger.Error("error parsing time", err)
	// }
	data.UpdatedAt, _ = time.Parse(time.RFC3339Nano, result["updated_at"])
	// if err != nil {
	// 	r.logger.Error("error parsing time", err)
	// }
	data.AccessCount, _ = strconv.Atoi(result["access_count"])
	return data, nil
}

func (r *RedisURLRepository) UpdateUrl(ctx context.Context, requestedCode, newLongUrl string, updatedAt time.Time, ownerID int) (models.UrlInfo, error) {
	var data models.UrlInfo
	res, err := r.client.HGet(ctx, "short:"+requestedCode, "owner_id").Result()
	if err != nil {
		// r.logger.Error("Error getting owner_id from redis", err)
		return data, appErr.ErrNotFound
	}
	resint, err := strconv.Atoi(res)
	if err != nil || resint != ownerID {
		// r.logger.Error("User doesn't own this resource", err)
		return data, appErr.ErrForbidden
	}
	err = r.client.HSet(ctx, "short:"+requestedCode, "url", newLongUrl).Err()
	if err != nil {
		// r.logger.Error("Error getting url from redis", err)
		return data, err
	}
	err = r.client.HSet(ctx, "short:"+requestedCode, "updated_at", updatedAt).Err()
	if err != nil {
		// r.logger.Error("Error changing update time in redis", err)
		return data, err
	}
	data, err = r.GetUrlStats(ctx, requestedCode, ownerID)
	if err != nil {
		// r.logger.Error("Error getting stats from redis", err)
		return data, err
	}
	return data, nil
}

func (r *RedisURLRepository) DeleteUrl(ctx context.Context, shortCode string, ownerID int) error {
	res, err := r.client.HGet(ctx, "short:"+shortCode, "owner_id").Result()
	if err != nil {
		// r.logger.Error("Error getting owner_id from redis in delete", err)
		return appErr.ErrNotFound
	}
	resint, err := strconv.Atoi(res)
	if err != nil || resint != ownerID {
		// r.logger.Error("User doesn't own this resource", err)
		return appErr.ErrForbidden
	}
	err = r.client.Del(ctx, "short:"+shortCode).Err()
	return err
}
