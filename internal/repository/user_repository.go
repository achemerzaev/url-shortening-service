package repository

import (
	"go.uber.org/zap"
)

type UserRepository struct {
	db *pgxpool.Pool
	logger *zap.Logger
}

func NewUserRepository(db *pgxpool.Pool, logger *zap.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}
