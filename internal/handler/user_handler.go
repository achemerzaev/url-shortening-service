package handler

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/achemerzaev/url-shortening-service/internal/dto"
	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"net/http"
)

type UserService interface {
	ServiceRegister(ctx context.Context, newUser models.User) (models.User, models.Tokens, error)
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

// @Summary Register user
// @Description Registers user and returns jwt tokens
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.HandlerRegisterRequest true "payload"
// @Success 201 {object} dto.HandlerRegisterResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /register [post]
func (h *UserHandler) HandlerRegister(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	newUser, _ := c.Get("jsonBody")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	newerUser := newUser.(*dto.HandlerRegisterRequest)
	convertedUser := models.User{
		Name:     newerUser.Name,
		Email:    newerUser.Email,
		Password: newerUser.Password,
	}

	_, tokens, err := h.service.ServiceRegister(ctx, convertedUser)
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.JSON(http.StatusCreated, tokens)

}

// @Summary Login user
// @Description User login with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.HandlerLoginRequest true "payload"
// @Success 201 {object} dto.HandlerLoginResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /login [post]
func (h *UserHandler) HandlerLogin(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	loginUser, _ := c.Get("jsonBody")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var userInfo models.User
	userInfo.Email, userInfo.Password = loginUser.(*dto.HandlerLoginRequest).Email,
		loginUser.(*dto.HandlerLoginRequest).Password

	tokens, err := h.service.ServiceLogin(ctx, userInfo)
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

// @Summary Update Refresh token
// @Description Creates new refresh token for a user
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.HandlerRefreshRequest true "payload"
// @Success 201 {object} dto.HandlerRefreshResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /refresh [post]
func (h *UserHandler) HandlerRefresh(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	refreshToken, _ := c.Get("jsonBody")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	tokens, err := h.service.ServiceRefresh(ctx, refreshToken.(*dto.HandlerRefreshRequest).RefreshToken)

	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.JSON(http.StatusCreated, tokens)
}

func ErrorHandler(c *gin.Context, err error, logger logger.Logger) {

	switch {
	case errors.Is(err, appErr.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, appErr.ErrEmailExists):
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
	case errors.Is(err, appErr.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	case errors.Is(err, appErr.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired, please log in again"})
	case errors.Is(err, appErr.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "user dont own this resource"})
	case errors.Is(err, appErr.ErrGeneratingJWT):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization error on server side"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
	logger.Error("Error: ", err)
}
