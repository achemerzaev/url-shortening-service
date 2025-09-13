package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/boretsotets/url-shortening-service/internal/models"

	"context"
	"log"
)

type UrlRepository struct {
	db *pgxpool.Pool
}

func NewUrlRepository(db *pgxpool.Pool) *UrlRepository {
	return &UrlRepository{db: db}
}

func (r *UrlRepository)RepositoryPost(data models.UrlInfo) (models.UrlInfo, error) {
	var newUrl models.UrlInfo
	err := r.db.QueryRow(context.Background(), 
	"INSERT INTO urls (Url, ShortCode, CreatedAt, UpdatedAt) VALUES ($1, $2, $3, $4) RETURNING Id, Url, ShortCode, CreatedAt, UpdatedAt", 
	data.Url, data.ShortCode, data.CreatedAt, data.UpdatedAt).Scan(&newUrl.Id, &newUrl.Url, &newUrl.ShortCode, &newUrl.CreatedAt, &newUrl.UpdatedAt)
	log.Println(data.Url, data.CreatedAt)
	log.Println(newUrl.Id, newUrl.Url, newUrl.ShortCode, newUrl.CreatedAt, newUrl.UpdatedAt, newUrl.AccessCount)
	return newUrl, err
}