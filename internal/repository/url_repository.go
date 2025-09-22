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
	"INSERT INTO urls (Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, owner_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING Id, Url, ShortCode, CreatedAt, UpdatedAt, owner_id", 
	data.Url, data.ShortCode, data.CreatedAt, data.UpdatedAt, data.AccessCount, data.OwnerID).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.OwnerID)
	return newData, err
}

func (r *UrlRepository)RepositoryGet(requestedCode string, ownerID int) (models.UrlInfo, error) {
	var newData models.UrlInfo
	err := r.db.QueryRow(context.Background(), 
	"SELECT Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, owner_id FROM urls WHERE shortcode = $1 AND owner_id = $2", 
	requestedCode, ownerID).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)
	r.logger.Info("checking struct", zap.Int("id here ", newData.Id), zap.String("url here", newData.Url), zap.String("short code here", newData.ShortCode))
	return newData, err
}

func (r *UrlRepository)RepositoryUpdate(requestedCode string, longurl string, updatedAt time.Time, ownerID int) (models.UrlInfo, error) {
	var newData models.UrlInfo
	err := r.db.QueryRow(context.Background(), 
	"UPDATE urls SET url = $1, updatedAt = $2 WHERE shortcode = $3 AND owner_id = $4 RETURNING Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, owner_id", 
	longurl, updatedAt, requestedCode, ownerID).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)
	return newData, err
}

func (r *UrlRepository)RepositoryDelete(requestedCode string, ownerID int) (error) {
	var checkDeletedUrl string
	err := r.db.QueryRow(context.Background(), 
	"DELETE FROM urls WHERE shortcode = $1 AND owner_id = $2 RETURNING url", 
	requestedCode, ownerID).Scan(&checkDeletedUrl)
	return err
}

func (r *UrlRepository)RepositoryGetStats(requestedCode string, ownerID int) (models.UrlInfo, error) {
	var newData models.UrlInfo
	err := r.db.QueryRow(context.Background(), 
	"SELECT Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, owner_id FROM urls WHERE shortcode = $1 AND owner_id = $2", 
	requestedCode, ownerID).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)
	return newData, err
}