package handler

import (
	"github.com/boretsotets/url-shortening-service/internal/service"
	"github.com/boretsotets/url-shortening-service/internal/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"encoding/json"
	"net/http"
	"strings"

)

type UrlHandler struct {
	service *service.UrlService
	logger *zap.Logger
}

func NewUrlHandler(s *service.UrlService, logger *zap.Logger) *UrlHandler {
	return &UrlHandler{service: s, logger: logger}
}

func (h *UrlHandler)HandlerPost(c *gin.Context) {
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

	c.IndentedJSON(http.StatusOK, newUrl)
}

func (h *UrlHandler)HandlerGet(c *gin.Context) {
	requestedCode := c.Param("shortcode")
	newData, err := h.service.ServiceGet(requestedCode)
	if err != nil {
		h.logger.Error("database retrieval error", zap.Error(err))
		c.String(http.StatusNotFound, "not found")
		return
	}
	c.IndentedJSON(http.StatusOK, newData)
	return
}

func (h *UrlHandler)HandlerPut(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var newUrl map[string]string
	err := json.NewDecoder(c.Request.Body).Decode(&newUrl)
	if err != nil {
		h.logger.Error("json decoding error", zap.Error(err))
		c.String(http.StatusInternalServerError, "not found")
		return
	}
	if _, ok := newUrl["url"]; !ok {
		h.logger.Error("json decoding error", zap.Error(err))
		c.String(http.StatusBadRequest, "please provide new url in json")
		return
	}

	requestedCode := c.Param("shortcode")

	newData, err := h.service.ServicePut(requestedCode, newUrl["url"])
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			h.logger.Error("short url not found", zap.Error(err))
			c.String(http.StatusNotFound, "short url not found")	
		} else {
			h.logger.Error("url updating error", zap.Error(err))
			c.String(http.StatusBadRequest, "url updating error")	
		}
		return
	}
	
	c.IndentedJSON(http.StatusOK, newData)
}

func (h *UrlHandler)HandlerDelete(c *gin.Context) {
	requestedCode := c.Param("shortcode")
	err := h.service.ServiceDelete(requestedCode)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			h.logger.Error("short url not found", zap.Error(err))
			c.String(http.StatusNotFound, "short url not found")	
		} else {
			h.logger.Error("database delete problem", zap.Error(err))
			c.String(http.StatusInternalServerError, "database delete problem")	
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *UrlHandler)HandlerGetStats(c *gin.Context) {
	requestedCode := c.Param("shortcode")
	statData, err := h.service.ServiceGetStats(requestedCode)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			h.logger.Error("short url not found", zap.Error(err))
			c.String(http.StatusNotFound, "short url not found")	
		} else {
			h.logger.Error("error getting stats", zap.Error(err))
			c.String(http.StatusInternalServerError, "error getting statistics")	
		}
		return
	}

	c.IndentedJSON(http.StatusOK, statData)
}