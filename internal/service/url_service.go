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