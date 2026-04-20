package handler

import (
	"context"
	"time"

	"github.com/achemerzaev/url-shortening-service/internal/models"
	"github.com/achemerzaev/url-shortening-service/internal/service"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"github.com/gin-gonic/gin"

	"net/http"
)

type URLHandler struct {
	service *service.UrlService
	logger  logger.Logger
}

func NewUrlHandler(s *service.UrlService, logger logger.Logger) *URLHandler {
	return &URLHandler{service: s, logger: logger}
}

func (h *URLHandler) HandlerPost(c *gin.Context) {
	// database inserting error
	c.Header("Content-Type", "application/json")
	clientID, _ := c.Get("clientID")
	userUrl, _ := c.Get("jsonBody")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var newUrl models.UrlInfo
	newUrl.Url = userUrl.(*models.PostRequestJSON).Url
	newUrl.OwnerID = clientID.(int)
	newUrl, err := h.service.ServicePost(ctx, newUrl)
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.IndentedJSON(http.StatusOK, newUrl)
}

func (h *URLHandler) HandlerGet(c *gin.Context) {
	clientID, _ := c.Get("clientID")
	requestedCode := c.Param("shortcode")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	longUrl, err := h.service.ServiceGet(ctx, requestedCode, clientID.(int))

	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.Redirect(http.StatusFound, longUrl)
}

func (h *URLHandler) HandlerPut(c *gin.Context) {
	// needed short url and new long url
	c.Header("Content-Type", "application/json")
	clientID, _ := c.Get("clientID")
	newUrl, _ := c.Get("jsonBody")
	requestedCode := c.Param("shortcode")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	newData, err := h.service.ServicePut(ctx, requestedCode, newUrl.(*models.PutRequestJSON).Url, clientID.(int))
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.IndentedJSON(http.StatusOK, newData)
}

func (h *URLHandler) HandlerDelete(c *gin.Context) {
	clientID, _ := c.Get("clientID")
	requestedCode := c.Param("shortcode")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := h.service.ServiceDelete(ctx, requestedCode, clientID.(int))
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *URLHandler) HandlerGetStats(c *gin.Context) {
	clientID, _ := c.Get("clientID")
	requestedCode := c.Param("shortcode")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	statData, err := h.service.ServiceGetStats(ctx, requestedCode, clientID.(int))

	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}

	c.IndentedJSON(http.StatusOK, statData)
}
