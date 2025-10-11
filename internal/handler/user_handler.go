package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/boretsotets/url-shortening-service/internal/authorization"
	"github.com/boretsotets/url-shortening-service/internal/models"
	"github.com/boretsotets/url-shortening-service/internal/service"

	"net/http"
	"strings"
)

type UserHandler struct {
	service *service.UserService
	logger  *zap.Logger
}

func NewUserHandler(s *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{service: s, logger: logger}
}

func (h *UserHandler) HandlerRegister(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	newUser, exists := c.Get("jsonBody")
	if !exists {
		h.logger.Info("error retrieving jsonBody from context")
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	_, tokens, err := h.service.ServiceRegister(*newUser.(*models.PostUserRegistration))
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			h.logger.Error("key duplicate error", zap.Error(err))
			c.String(http.StatusConflict, "user with this email already exists")	
		} else {
			h.logger.Error("error inserting user", zap.Error(err))
			c.String(http.StatusInternalServerError, "error inserting user")
		}
		return
	}
	c.IndentedJSON(http.StatusCreated, tokens)

}

func (h *UserHandler) HandlerLogin(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	loginUser, exists := c.Get("jsonBody")
	if !exists {
		h.logger.Info("error retrieving jsonBody from context")
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}
	var userInfo models.User
	userInfo.Email, userInfo.Password = loginUser.(*models.PostUserLogin).Email, 
	loginUser.(*models.PostUserLogin).Password

	tokens, err := h.service.ServiceLogin(userInfo)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			h.logger.Error("user does not exist", zap.Error(err))
			c.String(http.StatusNotFound, "username or password are not correct")
		} else {
			h.logger.Error("error logging in", zap.Error(err))
			c.String(http.StatusInternalServerError, "login error")	
		}
		return
	}

	c.IndentedJSON(http.StatusOK, tokens)
}

func (h *UserHandler) HandlerRefresh(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	refreshToken, exists := c.Get("jsonBody")
	if !exists {
		h.logger.Info("error retrieving jsonBody from context")
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	userID, err := authorization.ValidateJWT(refreshToken.(*models.PostRefreshToken).RefreshToken)
	if err != nil {
		h.logger.Error("Refresh token is not valid", zap.Error(err))
		c.String(http.StatusUnauthorized, "Refresh token is not valid")
		return
	}

	tokens, err := h.service.ServiceRefresh(userID, refreshToken.(*models.PostRefreshToken).RefreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "refresh token is not valid") {
			c.String(http.StatusUnauthorized, "Refresh token is not valid")
		} else if strings.Contains(err.Error(), "No rows") {
			c.String(http.StatusBadRequest, "User has no valid refresh tokens. Please, log in or register")
		} else {
			h.logger.Error("error refreshing token", zap.Error(err))
			c.String(http.StatusInternalServerError, "error refreshing token")	
		}
		return
	}

	c.IndentedJSON(http.StatusCreated, tokens)
}
