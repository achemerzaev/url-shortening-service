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
	longUrl, err := s.redisrepo.GetUrl(context.Background(), requestedCode, ownerID)
	if newData.OwnerID != 0 && newData.OwnerID != ownerID {
		s.logger.Error("user has no access to this row")
		return "", errors.New("forbidden: user does not own this resource")
	} else if err != nil {
		newData, err = s.repo.RepositoryGet(requestedCode)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			s.logger.Error("user has no access to this row")
			return "", errors.New("forbidden: user does not own this resource")
		} else if err != nil {
			s.logger.Error("error getting redirect link from db")
			return "", err
		}
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
	if newData.OwnerID != 0 && newData.OwnerID != ownerID {
		s.logger.Error("user has no access to this row")
		return newData, errors.New("forbidden: user does not own this resource")
	} else if err != nil {
		s.logger.Info("error updating url in redis", zap.Error(err))
		newData, err := s.repo.RepositoryUpdate(requestedCode, longUrl, updatedAt)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			s.logger.Error("user has no access to this row")
			return newData, errors.New("forbidden: user does not own this resource")
		} else if err != nil {
			s.logger.Info("error updating url in db", zap.Error(err))
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
	// как тут с несовпадением ownerID
	err := s.redisrepo.DeleteUrl(context.Background(), requestedCode, ownerID)
	if err != nil {
		if !strings.Contains(err.Error(), "nil") {
			return err
		}
	}
	err = s.repo.RepositoryDelete(requestedCode, ownerID)
	if err != nil {
		return err
	}
	return nil
}

func (s *UrlService) ServiceGetStats(requestedCode string, ownerID int) (models.UrlInfo, error) {
	var newData models.UrlInfo
	newData, err := s.redisrepo.GetUrlStats(context.Background(), requestedCode, ownerID)
	if newData.OwnerID != 0 && newData.OwnerID != ownerID {
		s.logger.Error("user has no access to this row")
		return newData, errors.New("forbidden: user does not own this resource")
	} else if err != nil {
		s.logger.Info("stats not found in redis", zap.Error(err))
		newData, err := s.repo.RepositoryGetStats(requestedCode)
		if newData.OwnerID != 0 && newData.OwnerID != ownerID {
			s.logger.Error("user has no access to this row")
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
