package handler

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"net/http"
)

type UserService interface {
	ServiceRegister(ctx context.Context, newUser models.PostUserRegistration) (models.User, models.Tokens, error)
	ServiceLogin(ctx context.Context, userinfo models.User) (models.Tokens, error)
	ServiceRefresh(ctx context.Context, refreshToken string) (models.Tokens, error)
}

type UserHandler struct {
	service UserService
	logger  logger.Logger
}

func NewUserHandler(s UserService, logger logger.Logger) *UserHandler {
	return &UserHandler{service: s, logger: logger}
}

func (h *UserHandler) HandlerRegister(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	newUser, _ := c.Get("jsonBody")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, tokens, err := h.service.ServiceRegister(ctx, *newUser.(*models.PostUserRegistration))
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.IndentedJSON(http.StatusCreated, tokens)

}

func (h *UserHandler) HandlerLogin(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	loginUser, _ := c.Get("jsonBody")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var userInfo models.User
	userInfo.Email, userInfo.Password = loginUser.(*models.PostUserLogin).Email,
		loginUser.(*models.PostUserLogin).Password

	tokens, err := h.service.ServiceLogin(ctx, userInfo)
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.IndentedJSON(http.StatusOK, tokens)
}

func (h *UserHandler) HandlerRefresh(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	refreshToken, _ := c.Get("jsonBody")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	tokens, err := h.service.ServiceRefresh(ctx, refreshToken.(*models.PostRefreshToken).RefreshToken)

	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.IndentedJSON(http.StatusCreated, tokens)
}

func ErrorHandler(c *gin.Context, err error, logger logger.Logger) {

	switch {
	case errors.Is(err, appErr.ErrNotFound):
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, appErr.ErrEmailExists):
		c.IndentedJSON(http.StatusConflict, gin.H{"error": "email already exists"})
	case errors.Is(err, appErr.ErrInvalidCredentials):
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	case errors.Is(err, appErr.ErrInvalidToken):
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "session expired, please log in again"})
	case errors.Is(err, appErr.ErrForbidden):
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "user dont own this resource"})
	case errors.Is(err, appErr.ErrGeneratingJWT):
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "authorization error on server side"})
	default:
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
	logger.Error("Error: ", err)
}
