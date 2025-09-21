package repository

import (
	"go.uber.org/zap"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/boretsotets/url-shortening-service/internal/models"

	"context"
)

type UserRepository struct {
	db *pgxpool.Pool
	logger *zap.Logger
}

func NewUserRepository(db *pgxpool.Pool, logger *zap.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

func (r *UserRepository)RepoInsertUser(newUser models.User) (models.User, error){
	var checkInsertedInfo models.User
	err := r.db.QueryRow(context.Background(),
	"INSERT INTO url_users (Name, Email, Password) VALUES ($1, $2, $3) RETURNING Id, Name, Email", 
	newUser.Name, newUser.Email, newUser.Password).Scan(&checkInsertedInfo.Id, &checkInsertedInfo.Name, &checkInsertedInfo.Email)
	return checkInsertedInfo, err
}

func (r *UserRepository)RepoRetrieveUser(email string) (string, error) {
	var password string
	err := r.db.QueryRow(context.Background(),
	"SELECT Password FROM url_users WHERE Email = $1", 
	email).Scan(&password)
	return password, err
}