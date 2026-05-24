package repository

import (
	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"context"
	"errors"
	"time"
)

type URLRepository struct {
	db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) RepositoryPost(ctx context.Context, data models.URLInfo) (models.URLInfo, error) {
	var newData models.URLInfo

	err := r.db.QueryRow(ctx,
		"INSERT INTO urls (Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, ownerid) VALUES ($1, $2, $3, $4, $5, $6) RETURNING Id, Url, ShortCode, CreatedAt, UpdatedAt, ownerid",
		data.Url, data.ShortCode, data.CreatedAt, data.UpdatedAt, data.AccessCount, data.OwnerID).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.OwnerID)
	return newData, err
}

func (r *URLRepository) RepositoryGet(ctx context.Context, requestedCode string) (models.URLInfo, error) {
	var newData models.URLInfo
	err := r.db.QueryRow(ctx,
		"SELECT Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, ownerid FROM urls WHERE shortcode = $1",
		requestedCode).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return newData, appErr.ErrNotFound
		}
	}
	return newData, err
}

func (r *URLRepository) RepositoryUpdate(ctx context.Context, requestedCode string, longURL string, updatedAt time.Time, ownerID int) (models.URLInfo, error) {
	var newData models.URLInfo
	err := r.db.QueryRow(ctx,
		"UPDATE urls SET url = $1, updatedAt = $2 WHERE shortcode = $3 AND ownerid = $4 RETURNING Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, ownerid",
		longURL, updatedAt, requestedCode, ownerID).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			exists, err := r.existsByShortCode(ctx, requestedCode)
			if err != nil {
				return newData, err
			}
			if exists {
				return newData, appErr.ErrForbidden
			}
			return newData, appErr.ErrNotFound
		}
		return newData, err
	}

	return newData, err
}

func (r *URLRepository) RepositoryDelete(ctx context.Context, requestedCode string, ownerID int) error {
	var deletedUrl string
	err := r.db.QueryRow(ctx,
		"DELETE FROM urls WHERE shortcode = $1 AND ownerid = $2 RETURNING url",
		requestedCode, ownerID).Scan(&deletedUrl)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			exists, err := r.existsByShortCode(ctx, requestedCode)
			if err != nil {
				return err
			}
			if exists {
				return appErr.ErrForbidden
			}
			return appErr.ErrNotFound
		}
		return err
	}

	return err
}

func (r *URLRepository) RepositoryGetStats(ctx context.Context, requestedCode string) (models.URLInfo, error) {
	var newData models.URLInfo
	err := r.db.QueryRow(ctx,
		"SELECT Id, Url, ShortCode, CreatedAt, UpdatedAt, AccessCount, ownerid FROM urls WHERE shortcode = $1",
		requestedCode).Scan(&newData.Id, &newData.Url, &newData.ShortCode, &newData.CreatedAt, &newData.UpdatedAt, &newData.AccessCount, &newData.OwnerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return newData, appErr.ErrNotFound
	}

	return newData, err
}

func (r *URLRepository) existsByShortCode(ctx context.Context, requestedCode string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE shortcode = $1)", requestedCode).Scan(&exists)
	return exists, err
}
