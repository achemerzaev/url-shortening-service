package service

import (
	"fmt"

	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"

	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"
)

type URLRepository interface {
	RepositoryPost(ctx context.Context, data models.URLInfo) (models.URLInfo, error)
	RepositoryGet(ctx context.Context, requestedCode string) (models.URLInfo, error)
	RepositoryUpdate(ctx context.Context, requestedCode string, longURL string, updatedAt time.Time, ownerID int) (models.URLInfo, error)
	RepositoryDelete(ctx context.Context, requestedCode string, ownerID int) error
	RepositoryGetStats(ctx context.Context, requestedCode string) (models.URLInfo, error)
}

type RedisURLRepository interface {
	SaveUrl(ctx context.Context, data models.URLInfo) error
	GetUrl(ctx context.Context, shortCode string, ownerID int) (string, error)
	IncrementCounter(ctx context.Context, shortCode string) error
	GetUrlStats(ctx context.Context, shortCode string, ownerID int) (models.URLInfo, error)
	UpdateUrl(ctx context.Context, requestedCode, newlongURL string, updatedAt time.Time, ownerID int) (models.URLInfo, error)
	DeleteUrl(ctx context.Context, shortCode string, ownerID int) error
}

type URLService struct {
	repo      URLRepository
	redisrepo RedisURLRepository
}

func NewUrlService(r URLRepository, redisr RedisURLRepository) *URLService {
	return &URLService{repo: r, redisrepo: redisr}
}

func (s *URLService) ServicePost(ctx context.Context, data models.URLInfo) (models.URLInfo, error) {
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
	randomString := make([]byte, length)

	// Generating Random string
	for i := range randomString {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		randomString[i] = chars[n.Int64()]
	}
	return string(randomString), nil
}

func (s *URLService) ServiceGet(ctx context.Context, requestedCode string, ownerID int) (string, error) {
	var newData models.URLInfo
	longURL, err := s.redisrepo.GetUrl(ctx, requestedCode, ownerID)
	if err != nil {
		if !errors.Is(err, appErr.ErrNotFound) {
			if errors.Is(err, appErr.ErrForbidden) {
				return "", appErr.ErrForbidden
			} else {
				return "", err
			}
		}
	} else {
		err = s.redisrepo.IncrementCounter(ctx, requestedCode)
		return longURL, err
	}

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
	newData.AccessCount += 1
	err = s.redisrepo.SaveUrl(ctx, newData)
	if err != nil {
		return "", fmt.Errorf("service get save to redis error: %w", err)
	}
	longURL = newData.Url

	return longURL, nil
}

func (s *URLService) ServicePut(ctx context.Context, requestedCode string, longURL string, ownerID int) (models.URLInfo, error) {
	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "https://" + longURL
	}
	updatedAt := time.Now()
	cachedData, err := s.redisrepo.UpdateUrl(ctx, requestedCode, longURL, updatedAt, ownerID)
	if err != nil && !errors.Is(err, appErr.ErrNotFound) {
		if errors.Is(err, appErr.ErrForbidden) {
			return cachedData, appErr.ErrForbidden
		} else {
			return cachedData, err
		}
	}
	cacheHit := err == nil

	newData, err := s.repo.RepositoryUpdate(ctx, requestedCode, longURL, updatedAt, ownerID)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return newData, appErr.ErrNotFound
		} else if errors.Is(err, appErr.ErrForbidden) {
			return newData, appErr.ErrForbidden
		} else {
			return newData, fmt.Errorf("service put repository update error: %w", err)
		}
	}
	if cacheHit {
		newData.AccessCount = cachedData.AccessCount
	}
	err = s.redisrepo.SaveUrl(ctx, newData)
	if err != nil {
		return newData, fmt.Errorf("service put save to redis error: %w", err)
	}

	return newData, nil
}

func (s *URLService) ServiceDelete(ctx context.Context, requestedCode string, ownerID int) error {
	err := s.redisrepo.DeleteUrl(ctx, requestedCode, ownerID)
	if err != nil && !errors.Is(err, appErr.ErrNotFound) {
		if errors.Is(err, appErr.ErrForbidden) {
			return appErr.ErrForbidden
		}
		return fmt.Errorf("service delete redis error: %w", err)
	}
	err = s.repo.RepositoryDelete(ctx, requestedCode, ownerID)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return appErr.ErrNotFound
		} else {
			return fmt.Errorf("service delete repo error: %w", err)
		}
	}
	return nil
}

func (s *URLService) ServiceGetStats(ctx context.Context, requestedCode string, ownerID int) (models.URLInfo, error) {
	var newData models.URLInfo
	newData, err := s.redisrepo.GetUrlStats(ctx, requestedCode, ownerID)
	if err == nil {
		return newData, nil
	}
	if !errors.Is(err, appErr.ErrNotFound) {
		if errors.Is(err, appErr.ErrForbidden) {
			return newData, appErr.ErrForbidden
		}
		return newData, fmt.Errorf("service get stats redis error: %w", err)
	}

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

	return newData, nil
}
