package service

import (
	"go.uber.org/zap"
)

type UserService struct {
	repo *repository.UrlRepository
	logger *zap.Logger
}

func NewUserService(r *repository.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{repo: r, logger: logger}
}