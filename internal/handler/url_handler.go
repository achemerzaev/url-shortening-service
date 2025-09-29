package handler

import (
	"github.com/boretsotets/url-shortening-service/internal/models"
	"github.com/boretsotets/url-shortening-service/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"net/http"
	"strings"
)

type UrlHandler struct {
	service *service.UrlService
	logger  *zap.Logger
}

func NewUrlHandler(s *service.UrlService, logger *zap.Logger) *UrlHandler {
	return &UrlHandler{service: s, logger: logger}
}

func (h *UrlHandler) HandlerPost(c *gin.Context) {
	// database inserting error
	c.Header("Content-Type", "application/json")
	clientID, exists := c.Get("clientID")
	if exists != true {
		h.logger.Info("error retrieving clientID from context")
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	userUrl, exists := c.Get("jsonBody")
	if exists != true {
		h.logger.Info("error retrieving jsonBody from context")
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}
	
	var newUrl models.UrlInfo
	newUrl.Url = userUrl.(*models.PostRequestJSON).Url
	newUrl.OwnerID = clientID.(int)
	newUrl, err := h.service.ServicePost(newUrl)
	if err != nil {
		h.logger.Error("database insertion error", zap.Error(err))
		c.String(http.StatusInternalServerError, "problem inserting url")
		return
	}

	c.IndentedJSON(http.StatusOK, newUrl)
}

func (h *UrlHandler) HandlerGet(c *gin.Context) {
    clientID, exists := c.Get("clientID")
	if exists != true {
		h.logger.Fatal("error retrieving clientID")
	}
	requestedCode := c.Param("shortcode")

	longUrl, err := h.service.ServiceGet(requestedCode, clientID.(int))

	if err != nil {
		if strings.Contains(err.Error(), "redis") {
			h.logger.Warn("Error saving url in redis")
		} else if strings.Contains(err.Error(), "own this resource") {
			h.logger.Error("user has no access to resource")
			c.String(http.StatusForbidden, "user has no access to this resource")
			return
		} else { 
			h.logger.Error("database retrieval error", zap.Error(err))
			c.String(http.StatusNotFound, "not found")
			return
		}
	}
	c.Redirect(http.StatusFound, longUrl)
}

func (h *UrlHandler) HandlerPut(c *gin.Context) {
	// needed short url and new long url
	c.Header("Content-Type", "application/json")
	clientID, exists := c.Get("clientID")
	if exists != true {
		h.logger.Fatal("error retrieving clientID")
	}

	// var newUrl map[string]string
	newUrl, exists := c.Get("jsonBody")
	if exists != true {
		h.logger.Fatal("error retrieving jsonBody")
	}

	requestedCode := c.Param("shortcode")

	newData, err := h.service.ServicePut(requestedCode, newUrl.(*models.PutRequestJSON).Url, clientID.(int))
	if err != nil {
		if strings.Contains(err.Error(), "no rows")  {
			h.logger.Error("short url not found", zap.Error(err))
			c.String(http.StatusNotFound, "short url not found")
		} else if strings.Contains(err.Error(), "forbidden") {
			h.logger.Error("user has no access to this resource", zap.Error(err))
			c.String(http.StatusForbidden, "user has no access to this resource")
		} else {
			h.logger.Error("url updating error", zap.Error(err))
			c.String(http.StatusInternalServerError, "url updating error")
		}
		return
	}

	c.IndentedJSON(http.StatusOK, newData)
}

func (h *UrlHandler) HandlerDelete(c *gin.Context) {
	clientID, exists := c.Get("clientID")
	if exists != true {
		h.logger.Fatal("error retrieving clientID")
	}
	requestedCode := c.Param("shortcode")
	err := h.service.ServiceDelete(requestedCode, clientID.(int))
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			h.logger.Error("short url not found", zap.Error(err))
			c.String(http.StatusNotFound, "short url not found")
		} else if strings.Contains(err.Error(), "forbidden") {
			h.logger.Error("user has no access to this resource", zap.Error(err))
			c.String(http.StatusForbidden, "user has no access to this resource")
		} else {
			h.logger.Error("database delete error", zap.Error(err))
			c.String(http.StatusInternalServerError, "database delete problem")
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *UrlHandler) HandlerGetStats(c *gin.Context) {
	clientID, exists := c.Get("clientID")
	if exists != true {
		h.logger.Fatal("error retrieving clientID")
	}
	h.logger.Info("", zap.Int("client id here: ", clientID.(int)))
	requestedCode := c.Param("shortcode")
	statData, err := h.service.ServiceGetStats(requestedCode, clientID.(int))

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			h.logger.Error("short url not found", zap.Error(err))
			c.String(http.StatusNotFound, "short url not found")
		} else if strings.Contains(err.Error(), "forbidden") {
			h.logger.Error("user has no access to this resource", zap.Error(err))
			c.String(http.StatusForbidden, "user has no access to this resource")
		} else {
			h.logger.Error("error getting stats", zap.Error(err))
			c.String(http.StatusInternalServerError, "error getting statistics")
		}
		return
	}

	c.IndentedJSON(http.StatusOK, statData)
}
