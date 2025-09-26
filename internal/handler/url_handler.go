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
		h.logger.Fatal("error retrieving clientID")
	}
	//var userUrl models.UrlInfo

	userUrl, exists := c.Get("jsonBody")
	if exists != true {
		h.logger.Fatal("error retrieving jsonBody")
	}
	/*
	err = json.NewDecoder(c.Request.Body).Decode(&userUrl)
	if err != nil {
		h.logger.Error("request decoding error", zap.Error(err))
		c.String(http.StatusBadRequest, "problem decoding json")
		return
	}
	*/
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
		h.logger.Error("database retrieval error", zap.Error(err))
		c.String(http.StatusNotFound, "not found")
		return
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
	/*
	err = json.NewDecoder(c.Request.Body).Decode(&newUrl)
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
	*/
	requestedCode := c.Param("shortcode")

	newData, err := h.service.ServicePut(requestedCode, newUrl.(*models.PutRequestJSON).Url, clientID.(int))
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
	requestedCode := c.Param("shortcode")
	statData, err := h.service.ServiceGetStats(requestedCode, clientID.(int))
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
