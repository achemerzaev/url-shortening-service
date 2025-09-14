package handler

import (
	"github.com/boretsotets/url-shortening-service/internal/service"
	"github.com/boretsotets/url-shortening-service/internal/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"encoding/json"
	"net/http"

)

type UrlHandler struct {
	service *service.UrlService
	logger *zap.Logger
}

func NewUrlHandler(s *service.UrlService, logger *zap.Logger) *UrlHandler {
	return &UrlHandler{service: s, logger: logger}
}

func (h *UrlHandler)HandlerPost(c *gin.Context) {
	h.logger.Info("HandlerPost starting",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	)
	c.Header("Content-Type", "application/json")
	var userUrl models.UrlInfo

	err := json.NewDecoder(c.Request.Body).Decode(&userUrl)
	if err != nil {
		h.logger.Error("request decoding error", zap.Error(err))
		c.String(http.StatusBadRequest, "problem decoding json")
		return
	}
	newUrl, err := h.service.ServicePost(userUrl)
	if err != nil {
		h.logger.Error("database insertion error", zap.Error(err))
		c.String(http.StatusInternalServerError, "problem inserting url")
		return
	}

	h.logger.Info("HandlerPost done, short url created",
		zap.String("original", newUrl.Url),
		zap.String("short", newUrl.ShortCode),
	)
	c.IndentedJSON(http.StatusOK, newUrl)
}