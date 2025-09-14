package repository

import (
	"github.com/boretsotets/url-shortening-service/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"context"
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
	"INSERT INTO urls (Url, ShortCode, CreatedAt, UpdatedAt) VALUES ($1, $2, $3, $4) RETURNING Id, Url, ShortCode, CreatedAt, UpdatedAt", 
	data.Url, data.ShortCode, data.CreatedAt, data.UpdatedAt).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt)
	return newData, err
}