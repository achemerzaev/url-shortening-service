package service

import (
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/internal/models"

	"time"
)

type UrlService struct {
	repo *repository.UrlRepository
}

func NewUrlService(r *repository.UrlRepository) *UrlService {
	return &UrlService{repo: r}
}

func (s *UrlService)ServicePost(data models.UrlInfo) (models.UrlInfo, error) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	data.ShortCode = "shortcode123"
	newInsertion, err := s.repo.RepositoryPost(data)
	return newInsertion, err
}