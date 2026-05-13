package handler

import (
	"context"
	"time"

	"github.com/achemerzaev/url-shortening-service/internal/dto"
	"github.com/achemerzaev/url-shortening-service/internal/models"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"github.com/gin-gonic/gin"

	"net/http"
)

type URLService interface {
	ServicePost(ctx context.Context, data models.URLInfo) (models.URLInfo, error)
	ServiceGet(ctx context.Context, requestedCode string, ownerID int) (string, error)
	ServicePut(ctx context.Context, requestedCode string, longURL string, ownerID int) (models.URLInfo, error)
	ServiceDelete(ctx context.Context, requestedCode string, ownerID int) error
	ServiceGetStats(ctx context.Context, requestedCode string, ownerID int) (models.URLInfo, error)
}

type URLHandler struct {
	service URLService
	logger  logger.Logger
}

func NewUrlHandler(s URLService, logger logger.Logger) *URLHandler {
	return &URLHandler{service: s, logger: logger}
}

// @Summary Create short URL
// @Description Creates short link for provided long link
// @Tags urls
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.HandlerPostRequest true "payload"
// @Success 201 {object} dto.HandlerPostResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /shorten [post]
func (h *URLHandler) HandlerPost(c *gin.Context) {
	// database inserting error
	c.Header("Content-Type", "application/json")
	clientID, _ := c.Get("clientID")
	userUrl, _ := c.Get("jsonBody")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var newUrl models.URLInfo
	newUrl.Url = userUrl.(*dto.HandlerPostRequest).URL
	newUrl.OwnerID = clientID.(int)
	newUrl, err := h.service.ServicePost(ctx, newUrl)
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.JSON(http.StatusOK, newUrl)
}

// @Summary Get short URL
// @Description Retrieves short link for provided long link
// @Tags urls
// @Produce json
// @Security BearerAuth
// @Param shortcode path string true "short code"
// @Success 201 {object} dto.HandlerGetResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /shorten/{shortcode} [get]
func (h *URLHandler) HandlerGet(c *gin.Context) {
	clientID, _ := c.Get("clientID")
	requestedCode := c.Param("shortcode")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	longURL, err := h.service.ServiceGet(ctx, requestedCode, clientID.(int))

	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.Redirect(http.StatusFound, longURL)
}

// @Summary Change long URL
// @Description Changes long link for provided short link
// @Tags urls
// @Produce json
// @Security BearerAuth
// @Param request body dto.HandlerPutRequest true "payload"
// @Param shortcode path string true "short code"
// @Success 201 {object} dto.HandlerPutResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /shorten/{shortcode} [put]
func (h *URLHandler) HandlerPut(c *gin.Context) {
	// needed short url and new long url
	c.Header("Content-Type", "application/json")
	clientID, _ := c.Get("clientID")
	newUrl, _ := c.Get("jsonBody")
	requestedCode := c.Param("shortcode")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	newData, err := h.service.ServicePut(ctx, requestedCode, newUrl.(*dto.HandlerPutRequest).URL, clientID.(int))
	if err != nil {
		ErrorHandler(c, err, h.logger)
		return
	}
	c.JSON(http.StatusOK, newData)
}

// @Summary Delete short url
// @Description Deletes short url
// @Tags urls
// @Security BearerAuth
// @Param shortcode path string true "short code"
// @Success 204
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /shorten/{shortcode} [delete]
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

// @Summary Get statistics
// @Description Get statistics for short URL
// @Tags urls
// @Produce json
// @Security BearerAuth
// @Param shortcode path string true "short code"
// @Success 201 {object} dto.HandlerGetStatsResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /shorten/{shortcode}/stats [get]
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

	c.JSON(http.StatusOK, statData)
}
