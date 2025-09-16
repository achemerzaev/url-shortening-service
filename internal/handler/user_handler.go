package handler

import (
	"go.uber.org/zap"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *service.UrlService
	logger *zap.Logger
}

func NewUserHandler(s *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{service: s, logger: logger}
}
