package repository

import (
	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"context"
	"errors"
	"time"
)

type UrlRepository struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func NewUrlRepository(db *pgxpool.Pool, logger logger.Logger) *UrlRepository {
	return &UrlRepository{db: db, logger: logger}
}

func (r *UrlRepository) RepositoryPost(ctx context.Context, data models.UrlInfo) (models.UrlInfo, error) {
	var newData models.UrlInfo

	err := r.db.QueryRow(ctx,
		"INSERT INTO urls (Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, owner_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING Id, Url, ShortCode, CreatedAt, UpdatedAt, owner_id",
		data.Url, data.ShortCode, data.CreatedAt, data.UpdatedAt, data.AccessCount, data.OwnerID).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.OwnerID)
	return newData, err
}

func (r *UrlRepository) RepositoryGet(ctx context.Context, requestedCode string) (models.UrlInfo, error) {
	var newData models.UrlInfo
	err := r.db.QueryRow(ctx,
		"SELECT Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, owner_id FROM urls WHERE shortcode = $1",
		requestedCode).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return newData, appErr.ErrNotFound
		}
	}
	return newData, err
}

func (r *UrlRepository) RepositoryUpdate(ctx context.Context, requestedCode string, longurl string, updatedAt time.Time, ownerID int) (models.UrlInfo, error) {
	var newData models.UrlInfo
	err := r.db.QueryRow(ctx,
		"SELECT Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, owner_id FROM urls WHERE shortcode = $1",
		requestedCode).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return newData, appErr.ErrNotFound
		}
		return newData, err
	}
	if newData.OwnerID != ownerID {
		return newData, appErr.ErrForbidden
	}

	err = r.db.QueryRow(ctx,
		"UPDATE urls SET url = $1, updatedAt = $2 WHERE shortcode = $3 RETURNING Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, owner_id",
		longurl, updatedAt, requestedCode).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)
	return newData, err
}

func (r *UrlRepository) RepositoryDelete(ctx context.Context, requestedCode string, ownerID int) error {
	var newData models.UrlInfo
	err := r.db.QueryRow(ctx,
		"SELECT Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, owner_id FROM urls WHERE shortcode = $1",
		requestedCode).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return appErr.ErrNotFound
		}
		return err
	}
	if newData.OwnerID != ownerID {
		return appErr.ErrForbidden
	}

	var deletedUrl string
	err = r.db.QueryRow(ctx,
		"DELETE FROM urls WHERE shortcode = $1 RETURNING url",
		requestedCode).Scan(&deletedUrl)
	return err
}

func (r *UrlRepository) RepositoryGetStats(ctx context.Context, requestedCode string) (models.UrlInfo, error) {
	var newData models.UrlInfo
	err := r.db.QueryRow(ctx,
		"SELECT Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, owner_id FROM urls WHERE shortcode = $1",
		requestedCode).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return newData, appErr.ErrNotFound
	}

	return newData, err
}
