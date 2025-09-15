package service

import (
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/internal/models"

	"go.uber.org/zap"

	"time"
	"math/rand"
)

type UrlService struct {
	repo *repository.UrlRepository
	logger *zap.Logger
}

func NewUrlService(r *repository.UrlRepository, logger *zap.Logger) *UrlService {
	return &UrlService{repo: r, logger: logger}
}

func (s *UrlService)ServicePost(data models.UrlInfo) (models.UrlInfo, error) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	data.ShortCode = GenerateShortCode()
	newInsertion, err := s.repo.RepositoryPost(data)
	return newInsertion, err
}

const chars = "1234567890abcdefghijklmnopqrstuvwxyz"

func GenerateShortCode() string {
    rand.Seed(time.Now().Unix())
    length := 6
    ran_str := make([]byte, length)

    // Generating Random string
    for i := range ran_str {
        ran_str[i] = chars[rand.Intn(len(chars))]
	}
	return string(ran_str)
}

func (s *UrlService)ServiceGet(requestedCode string) (models.UrlInfo, error) {
	newData, err := s.repo.RepositoryGet(requestedCode)
	return newData, err
}

func (s *UrlService)ServicePut(requestedCode string, longUrl string) (models.UrlInfo, error) {
	updatedAt := time.Now()
	newData, err := s.repo.RepositoryUpdate(requestedCode, longUrl, updatedAt)
	return newData, err
}

func (s *UrlService)ServiceDelete(requestedCode string) (error) {
	err := s.repo.RepositoryDelete(requestedCode)
	return err
}

func (s *UrlService)ServiceGetStats(requestedCode string) (models.UrlInfo, error) {
	newData, err := s.repo.RepositoryGetStats(requestedCode)
	return newData, err
}