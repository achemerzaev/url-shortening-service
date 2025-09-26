package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/boretsotets/url-shortening-service/internal/authorization"
	"github.com/boretsotets/url-shortening-service/internal/models"
	"github.com/boretsotets/url-shortening-service/internal/service"

	"encoding/json"
	"net/http"
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
	var newUser models.User
	err := json.NewDecoder(c.Request.Body).Decode(&newUser)
	if err != nil {
		h.logger.Error("json decoding error", zap.Error(err))
		c.String(http.StatusInternalServerError, "not found")
		return
	}

	_, tokens, err := h.service.ServiceRegister(newUser)
	if err != nil {
		h.logger.Error("error inserting user", zap.Error(err))
		c.String(http.StatusInternalServerError, "error inserting user")
		return
	}
	c.IndentedJSON(http.StatusCreated, tokens)

}

func (h *UserHandler) HandlerLogin(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var loginUser models.User
	err := json.NewDecoder(c.Request.Body).Decode(&loginUser)
	if err != nil {
		h.logger.Error("error decoding json", zap.Error(err))
		c.String(http.StatusBadRequest, "json error")
		return
	}

	tokens, err := h.service.ServiceLogin(loginUser)
	if err != nil {
		h.logger.Error("error logging in", zap.Error(err))
		c.String(http.StatusInternalServerError, "login error")
		return
	}

	c.IndentedJSON(http.StatusOK, tokens)
}

func (h *UserHandler) HandlerRefresh(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	var refreshtoken map[string]string
	err := json.NewDecoder(c.Request.Body).Decode(&refreshtoken)
	if err != nil {
		h.logger.Error("error decoding json", zap.Error(err))
		c.String(http.StatusInternalServerError, "error decoding json")
		return
	}

	userID, err := authorization.ValidateJWT(refreshtoken["refreshtoken"])
	if err != nil {
		h.logger.Error("error validating token", zap.Error(err))
		c.String(http.StatusInternalServerError, "error validating token")
		return
	}

	tokens, err := h.service.ServiceRefresh(userID, refreshtoken["refreshtoken"])
	if err != nil {
		h.logger.Error("error refreshing token", zap.Error(err))
		c.String(http.StatusInternalServerError, "error refreshing token")
		return
	}

	c.IndentedJSON(http.StatusCreated, tokens)
}
