package repository

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"context"
	"errors"
)

type UserRepository struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func NewUserRepository(db *pgxpool.Pool, logger logger.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

func (r *UserRepository) RepoInsertUser(ctx context.Context, newUser models.PostUserRegistration) (models.User, error) {
	var checkInsertedInfo models.User
	err := r.db.QueryRow(ctx,
		"INSERT INTO url_users (Name, Email, Password) VALUES ($1, $2, $3) RETURNING Id, Name, Email",
		newUser.Name, newUser.Email, newUser.Password).Scan(&checkInsertedInfo.Id, &checkInsertedInfo.Name, &checkInsertedInfo.Email)
	// duplicate key error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return checkInsertedInfo, appErr.ErrEmailExists
			}
		}
	}

	return checkInsertedInfo, err
}

func (r *UserRepository) RepoRetrieveUser(ctx context.Context, email string) (string, error) {
	var password string
	err := r.db.QueryRow(ctx,
		"SELECT Password FROM url_users WHERE Email = $1",
		email).Scan(&password)
	if errors.Is(err, pgx.ErrNoRows) {
		return password, appErr.ErrInvalidCredentials
	}

	return password, err
}
