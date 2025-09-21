package service

import (
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/internal/redisrepo"
	"github.com/boretsotets/url-shortening-service/internal/models"

	"go.uber.org/zap"

	"time"
	"crypto/rand"
	"math/big"
	"strings"
)

type UrlService struct {
	repo *repository.UrlRepository
	redisrepo *redisrepo.RedisRepository
	logger *zap.Logger
}

func NewUrlService(r *repository.UrlRepository, redisr *redisrepo.RedisRepository, logger *zap.Logger) *UrlService {
	return &UrlService{repo: r, redisrepo: redisr, logger: logger}
}

func (s *UrlService)ServicePost(data models.UrlInfo) (models.UrlInfo, error) {
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

func (s *UrlService)ServiceGet(requestedCode string, ownerID int) (string, error) {
	var newData *models.UrlInfo
	newData, err := s.repo.RepositoryGet(requestedCode, ownerID)
	newData.AccessCount += 1
	return newData.Url, err
}

func (s *UrlService)ServicePut(requestedCode string, longUrl string, ownerID int) (models.UrlInfo, error) {
	if !strings.HasPrefix(longUrl, "http://") && !strings.HasPrefix(longUrl, "https://") {
		longUrl = "https://" + longUrl
	}
	updatedAt := time.Now()
	newData, err := s.repo.RepositoryUpdate(requestedCode, longUrl, updatedAt, ownerID)
	return newData, err
}

func (s *UrlService)ServiceDelete(requestedCode string, ownerID int) (error) {
	err := s.repo.RepositoryDelete(requestedCode, ownerID)
	return err
}

func (s *UrlService)ServiceGetStats(requestedCode string, ownerID int) (models.UrlInfo, error) {
	newData, err := s.repo.RepositoryGetStats(requestedCode, ownerID)
	return newData, err
}