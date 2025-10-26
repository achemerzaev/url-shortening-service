package service

import (
	"github.com/boretsotets/url-shortening-service/internal/models"
	"github.com/boretsotets/url-shortening-service/internal/redisrepo"
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/pkg/errors"
	"github.com/boretsotets/url-shortening-service/pkg/logger"

	"context"
	"crypto/rand"
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

func (s *UrlService) ServicePost(data models.UrlInfo) (models.UrlInfo, *errors.AppError) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = data.CreatedAt
	code, err := GenerateShortCode()
	data.ShortCode = code
	if err != nil {
		return data, errors.Wrap("INTERNAL_ERROR", "error posting URL", err)
	}
	if !strings.HasPrefix(data.Url, "http://") && !strings.HasPrefix(data.Url, "https://") {
		data.Url = "https://" + data.Url
	}
	newInsertion, err := s.repo.RepositoryPost(data)
	if err != nil {
		return newInsertion, errors.Wrap("INTERNAL_ERROR", "error posting URL", err)
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

func (s *UrlService) ServiceGet(requestedCode string, ownerID int) (string, *errors.AppError) {
	var newData models.UrlInfo
	longUrl, err := s.redisrepo.GetUrl(context.Background(), requestedCode, ownerID)
	if err != nil && strings.Contains(err.Error(), "does't own") {
		return "", errors.ErrForbidden
	} else if err != nil {
		newData, err = s.repo.RepositoryGet(requestedCode)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			return "", errors.ErrForbidden
		} else if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return "", errors.ErrNotFound
			} else {
				return "", errors.Wrap("INTERNAL_ERROR", "error getting URL", err)
			}
		}
		err = s.redisrepo.SaveUrl(context.Background(), newData)
		if err != nil {
			return "", errors.Wrap("INTERNAL_ERROR", "error getting URL", err)
		}
		longUrl = newData.Url
	}

	return longUrl, nil
}

func (s *UrlService) ServicePut(requestedCode string, longUrl string, ownerID int) (models.UrlInfo, *errors.AppError) {
	if !strings.HasPrefix(longUrl, "http://") && !strings.HasPrefix(longUrl, "https://") {
		longUrl = "https://" + longUrl
	}
	updatedAt := time.Now()
	newData, err := s.redisrepo.UpdateUrl(context.Background(), requestedCode, longUrl, updatedAt, ownerID)
	if err != nil && strings.Contains(err.Error(), "doesn't own") {
		return newData, errors.ErrForbidden
	} else if err != nil {
		newData, err := s.repo.RepositoryUpdate(requestedCode, longUrl, updatedAt, ownerID)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			return newData, errors.ErrForbidden
		} else if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return newData, errors.ErrNotFound
			} else {
				return newData, errors.Wrap("INTERNAL_ERROR", "error changing URL", err)
			}
		}
		err = s.redisrepo.SaveUrl(context.Background(), newData)
		if err != nil {
			return newData, errors.Wrap("INTERNAL_ERROR", "error changing URL", err)
		}
	}
	return newData, nil
}

func (s *UrlService) ServiceDelete(requestedCode string, ownerID int) *errors.AppError {
	err := s.redisrepo.DeleteUrl(context.Background(), requestedCode, ownerID)
	if err != nil && !strings.Contains(err.Error(), "nil") {
		if strings.Contains(err.Error(), "doesn't own") {
			return errors.ErrForbidden
		}
		return errors.Wrap("INTERNAL_ERROR", "error deleting url", err)
	}
	err = s.repo.RepositoryDelete(requestedCode, ownerID)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't own") {
			return errors.ErrForbidden
		} else if strings.Contains(err.Error(), "no rows") {
			return errors.ErrNotFound
		} else {
			return errors.Wrap("INTERNAL_ERROR", "error deleting url", err)
		}
	}
	return nil
}

func (s *UrlService) ServiceGetStats(requestedCode string, ownerID int) (models.UrlInfo, *errors.AppError) {
	var newData models.UrlInfo
	newData, err := s.redisrepo.GetUrlStats(context.Background(), requestedCode, ownerID)
	if newData.OwnerID != 0 && newData.OwnerID != ownerID {
		return newData, errors.ErrForbidden
	} else if err != nil {
		newData, err = s.repo.RepositoryGetStats(requestedCode)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			return newData, errors.ErrForbidden
		} else if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return newData, errors.ErrNotFound
			} else {
				return newData, errors.Wrap("INTERNAL_ERROR", "error getting stats", err)
			}
		}
		err = s.redisrepo.SaveUrl(context.Background(), newData)
		if err != nil {
			return newData, errors.Wrap("INTERNAL_ERROR", "error getting stats", err)
		}
	}

	return newData, nil
}
