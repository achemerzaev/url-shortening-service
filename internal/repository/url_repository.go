package repository

import (
	"github.com/boretsotets/url-shortening-service/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"context"
	"time"
)

type UrlRepository struct {
	db *pgxpool.Pool
	logger *zap.Logger
}

func NewUrlRepository(db *pgxpool.Pool, logger *zap.Logger) *UrlRepository {
	return &UrlRepository{db: db, logger: logger}
}

func (r *UrlRepository)RepositoryPost(data models.UrlInfo) (models.UrlInfo, error) {
	var newData models.UrlInfo
	err := r.db.QueryRow(context.Background(), 
	"INSERT INTO urls (Url, ShortCode, CreatedAt, UpdatedAt, AccessCount) VALUES ($1, $2, $3, $4, $5) RETURNING Id, Url, ShortCode, CreatedAt, UpdatedAt", 
	data.Url, data.ShortCode, data.CreatedAt, data.UpdatedAt, data.AccessCount).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt)
	return newData, err
}

func (r *UrlRepository)RepositoryGet(requestedCode string) (*models.UrlInfo, error) {
	var newData models.UrlInfo
	err := r.db.QueryRow(context.Background(), 
	"SELECT Id, Url, ShortCode, CreatedAt, UpdatedAt FROM urls WHERE shortcode = $1", 
	requestedCode).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt)
	r.logger.Info("checking struct", zap.Int("id here ", newData.Id), zap.String("url here", newData.Url), zap.String("short code here", newData.ShortCode))

	if err == nil {
		var count int
		err = r.db.QueryRow(context.Background(), 
		"UPDATE urls SET AccessCount = AccessCount + 1 WHERE shortcode = $1 RETURNING AccessCount", 
		requestedCode).Scan(&count)
	}
	return &newData, err
}

func (r *UrlRepository)RepositoryUpdate(requestedCode string, longurl string, updatedAt time.Time) (models.UrlInfo, error) {
	var newData models.UrlInfo
	err := r.db.QueryRow(context.Background(), 
	"UPDATE urls SET url = $1, updatedAt = $2 WHERE shortcode = $3 RETURNING Id, Url, ShortCode, CreatedAt, UpdatedAt", 
	longurl, updatedAt, requestedCode).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt)
	return newData, err
}

func (r *UrlRepository)RepositoryDelete(requestedCode string) (error) {
	var checkDeletedUrl string
	err := r.db.QueryRow(context.Background(), 
	"DELETE FROM urls WHERE shortcode = $1 RETURNING url", 
	requestedCode).Scan(&checkDeletedUrl)
	return err
}

func (r *UrlRepository)RepositoryGetStats(requestedCode string) (models.UrlInfo, error) {
	var newData models.UrlInfo
	err := r.db.QueryRow(context.Background(), 
	"SELECT Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount FROM urls WHERE shortcode = $1", 
	requestedCode).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount)
	return newData, err
}