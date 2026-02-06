package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/achemerzaev/url-shortening-service/internal/models"
	"github.com/achemerzaev/url-shortening-service/internal/service"
	"github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"net/http"
)

type UserHandler struct {
	service *service.UserService
	logger  logger.Logger
}

func NewUserHandler(s *service.UserService, logger logger.Logger) *UserHandler {
	return &UserHandler{service: s, logger: logger}
}

func (h *UserHandler) HandlerRegister(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	newUser, _ := c.Get("jsonBody")

	_, tokens, err := h.service.ServiceRegister(*newUser.(*models.PostUserRegistration))
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.IndentedJSON(http.StatusCreated, tokens)

}

func (h *UserHandler) HandlerLogin(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	loginUser, _ := c.Get("jsonBody")

	var userInfo models.User
	userInfo.Email, userInfo.Password = loginUser.(*models.PostUserLogin).Email,
		loginUser.(*models.PostUserLogin).Password

	tokens, err := h.service.ServiceLogin(userInfo)
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.IndentedJSON(http.StatusOK, tokens)
}

func (h *UserHandler) HandlerRefresh(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	refreshToken, _ := c.Get("jsonBody")

	tokens, err := h.service.ServiceRefresh(refreshToken.(*models.PostRefreshToken).RefreshToken)

	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.IndentedJSON(http.StatusCreated, tokens)
}

func ErrorHandler(c *gin.Context, err *errors.AppError, logger logger.Logger) {

	status := map[string]int{
		"EMAIL_EXISTS":        http.StatusConflict,
		"INVALID_CREDENTIALS": http.StatusUnauthorized,
		"INVALID_TOKEN":       http.StatusUnauthorized,
		"FORBIDDEN":           http.StatusForbidden,
		"NOT_FOUND":           http.StatusNotFound,
		"JWT_ERROR":           http.StatusInternalServerError,
	}[err.Code]

	if status == 0 {
		status = http.StatusInternalServerError
	}

	c.IndentedJSON(status, gin.H{"error": err.Message})
	logger.Error("Error code: ", err.Code, " error: ", err.Err)
}
