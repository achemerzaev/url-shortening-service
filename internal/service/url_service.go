package service

import (
	"github.com/boretsotets/url-shortening-service/internal/models"
	"github.com/boretsotets/url-shortening-service/internal/redisrepo"
	"github.com/boretsotets/url-shortening-service/internal/repository"

	"go.uber.org/zap"

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
	logger    *zap.Logger
}

func NewUrlService(r *repository.UrlRepository, redisr *redisrepo.RedisRepository, logger *zap.Logger) *UrlService {
	return &UrlService{repo: r, redisrepo: redisr, logger: logger}
}

func (s *UrlService) ServicePost(data models.UrlInfo) (models.UrlInfo, error) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = data.CreatedAt
	code, err := GenerateShortCode()
	data.ShortCode = code
	if err != nil {
		return data, err
	}
	if !strings.HasPrefix(data.Url, "http://") && !strings.HasPrefix(data.Url, "https://") {
		data.Url = "https://" + data.Url
	}
	newInsertion, err := s.repo.RepositoryPost(data)
	return newInsertion, err
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

func (s *UrlService) ServiceGet(requestedCode string, ownerID int) (string, error) {
	var newData models.UrlInfo
	s.logger.Info("service get: ", zap.String("code: ", requestedCode))
	longUrl, err := s.redisrepo.GetUrl(context.Background(), requestedCode, ownerID)
	s.logger.Info("service get redis, error here: ", zap.Error(err))
	if err != nil && strings.Contains(err.Error(), "does't own") {
		return "", errors.New("forbidden: user does not own this resource")
	} else if err != nil {
		newData, err = s.repo.RepositoryGet(requestedCode)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			return "", errors.New("forbidden: user does not own this resource")
		} else if err != nil {
			return "", err
		}
		s.logger.Info("service get save url: ", zap.String("code: ", newData.ShortCode))
		err = s.redisrepo.SaveUrl(context.Background(), newData)
		if err != nil {
			s.logger.Error("error saving link to redis")
			return "", err
		}
		longUrl = newData.Url
	}

	return longUrl, err
}

func (s *UrlService) ServicePut(requestedCode string, longUrl string, ownerID int) (models.UrlInfo, error) {
	if !strings.HasPrefix(longUrl, "http://") && !strings.HasPrefix(longUrl, "https://") {
		longUrl = "https://" + longUrl
	}
	updatedAt := time.Now()
	newData, err := s.redisrepo.UpdateUrl(context.Background(), requestedCode, longUrl, updatedAt, ownerID)
	if err != nil && strings.Contains(err.Error(), "doesn't own") {
		s.logger.Info("error here:", zap.Error(err))
		return newData, errors.New("forbidden: user does not own this resource")
	} else if err != nil {
		s.logger.Info("error updating url in redis", zap.Error(err))
		newData, err := s.repo.RepositoryUpdate(requestedCode, longUrl, updatedAt, ownerID)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			return newData, errors.New("forbidden: user does not own this resource")
		} else if err != nil {
			return newData, err
		}
		err = s.redisrepo.SaveUrl(context.Background(), newData)
		if err != nil {
			s.logger.Info("error saving url in redis", zap.Error(err))
			return newData, err
		}
	}
	return newData, nil
}

func (s *UrlService) ServiceDelete(requestedCode string, ownerID int) error {
	err := s.redisrepo.DeleteUrl(context.Background(), requestedCode, ownerID)
	s.logger.Info("Error here:", zap.Error(err))
	if err != nil  && !strings.Contains(err.Error(), "nil")  {
		if strings.Contains(err.Error(), "doesn't own") {
			return errors.New("forbidden: user does not own this resource")
		}
		return err
	}
	s.logger.Info("i get here")
	err = s.repo.RepositoryDelete(requestedCode, ownerID)
	s.logger.Info("Error here:", zap.Error(err))
	if err != nil {
		return err
	}
	return nil
}

func (s *UrlService) ServiceGetStats(requestedCode string, ownerID int) (models.UrlInfo, error) {
	var newData models.UrlInfo
	newData, err := s.redisrepo.GetUrlStats(context.Background(), requestedCode, ownerID)
	if newData.OwnerID != 0 && newData.OwnerID != ownerID {
		return newData, errors.New("forbidden: user does not own this resource")
	} else if err != nil {
		s.logger.Info("error getting stats from redis", zap.Error(err))
		newData, err = s.repo.RepositoryGetStats(requestedCode)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			return newData, errors.New("forbidden: user does not own this resource")
		} else if err != nil {
			s.logger.Error("error getting stats from db", zap.Error(err))
			return newData, err
		}
		err = s.redisrepo.SaveUrl(context.Background(), newData)
		if err != nil {
			s.logger.Error("error saving url in redis", zap.Error(err))
			return newData, err
		}
	}

	return newData, err
}
