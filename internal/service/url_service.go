package service

import (
	"fmt"

	"github.com/achemerzaev/url-shortening-service/internal/models"
	"github.com/achemerzaev/url-shortening-service/internal/redisrepo"
	"github.com/achemerzaev/url-shortening-service/internal/repository"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"
)

type UrlService struct {
	repo      *repository.UrlRepository
	redisrepo *redisrepo.RedisRepository
	logger    logger.Logger
}

func NewUrlService(r *repository.UrlRepository, redisr *redisrepo.RedisRepository, logger logger.Logger) *UrlService {
	return &UrlService{repo: r, redisrepo: redisr, logger: logger}
}

func (s *UrlService) ServicePost(ctx context.Context, data models.UrlInfo) (models.UrlInfo, error) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = data.CreatedAt
	code, err := GenerateShortCode()
	data.ShortCode = code
	if err != nil {
		return data, fmt.Errorf("service post - generate short code: %w", err)
	}
	if !strings.HasPrefix(data.Url, "http://") && !strings.HasPrefix(data.Url, "https://") {
		data.Url = "https://" + data.Url
	}
	newInsertion, err := s.repo.RepositoryPost(ctx, data)
	if err != nil {
		return newInsertion, fmt.Errorf("service post - repository post: %w", err)
	}
	return newInsertion, nil
}

const chars = "1234567890abcdefghijklmnopqrstuvwxyz"

func GenerateShortCode() (string, error) {
	length := 6
	ran_str := make([]byte, length)

	// Generating Random string
	for i := range ran_str {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		ran_str[i] = chars[n.Int64()]
	}
	return string(ran_str), nil
}

func (s *UrlService) ServiceGet(ctx context.Context, requestedCode string, ownerID int) (string, error) {
	var newData models.UrlInfo
	longUrl, err := s.redisrepo.GetUrl(ctx, requestedCode, ownerID)
	if err != nil && errors.Is(err, appErr.ErrForbidden) {
		return "", appErr.ErrForbidden
	} else if err != nil {
		newData, err = s.repo.RepositoryGet(ctx, requestedCode)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			return "", appErr.ErrForbidden
		} else if err != nil {
			if errors.Is(err, appErr.ErrNotFound) {
				return "", err
			} else {
				return "", fmt.Errorf("service get repo internal error: %w", err)
			}
		}
		err = s.redisrepo.SaveUrl(ctx, newData)
		if err != nil {
			return "", fmt.Errorf("service get save to redis error: %w", err)
		}
		longUrl = newData.Url
	}

	return longUrl, nil
}

func (s *UrlService) ServicePut(ctx context.Context, requestedCode string, longUrl string, ownerID int) (models.UrlInfo, error) {
	if !strings.HasPrefix(longUrl, "http://") && !strings.HasPrefix(longUrl, "https://") {
		longUrl = "https://" + longUrl
	}
	updatedAt := time.Now()
	newData, err := s.redisrepo.UpdateUrl(ctx, requestedCode, longUrl, updatedAt, ownerID)
	if err != nil && errors.Is(err, appErr.ErrForbidden) {
		return newData, appErr.ErrForbidden
	} else if err != nil {
		newData, err := s.repo.RepositoryUpdate(ctx, requestedCode, longUrl, updatedAt, ownerID)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			return newData, appErr.ErrForbidden
		} else if err != nil {
			if errors.Is(err, appErr.ErrNotFound) {
				return newData, appErr.ErrNotFound
			} else {
				return newData, fmt.Errorf("service put repository update error: %w", err)
			}
		}
		err = s.redisrepo.SaveUrl(ctx, newData)
		if err != nil {
			return newData, fmt.Errorf("service put save to redis error: %w", err)
		}
	}
	return newData, nil
}

func (s *UrlService) ServiceDelete(ctx context.Context, requestedCode string, ownerID int) error {
	err := s.redisrepo.DeleteUrl(ctx, requestedCode, ownerID)
	if err != nil && !errors.Is(err, appErr.ErrNotFound) {
		if errors.Is(err, appErr.ErrForbidden) {
			return appErr.ErrForbidden
		}
		return fmt.Errorf("service delete redis error: %w", err)
	}
	err = s.repo.RepositoryDelete(ctx, requestedCode, ownerID)
	if err != nil {
		if errors.Is(err, appErr.ErrForbidden) {
			return appErr.ErrForbidden
		} else if errors.Is(err, appErr.ErrNotFound) {
			return appErr.ErrNotFound
		} else {
			return fmt.Errorf("service delete repo error: %w", err)
		}
	}
	return nil
}

func (s *UrlService) ServiceGetStats(ctx context.Context, requestedCode string, ownerID int) (models.UrlInfo, error) {
	var newData models.UrlInfo
	newData, err := s.redisrepo.GetUrlStats(ctx, requestedCode, ownerID)
	if newData.OwnerID != 0 && newData.OwnerID != ownerID {
		return newData, appErr.ErrForbidden
	} else if err != nil {
		newData, err = s.repo.RepositoryGetStats(ctx, requestedCode)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			return newData, appErr.ErrForbidden
		} else if err != nil {
			if errors.Is(err, appErr.ErrNotFound) {
				return newData, appErr.ErrNotFound
			} else {
				return newData, fmt.Errorf("service get stats repo error: %w", err)
			}
		}
		err = s.redisrepo.SaveUrl(ctx, newData)
		if err != nil {
			return newData, fmt.Errorf("service get stats redis save err: %w", err)
		}
	}

	return newData, nil
}
