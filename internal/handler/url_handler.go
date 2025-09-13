package handler

import (
	"github.com/boretsotets/url-shortening-service/internal/service"
	"github.com/boretsotets/url-shortening-service/internal/models"

	"github.com/gin-gonic/gin"

	"encoding/json"
	"net/http"

)

type UrlHandler struct {
	service *service.UrlService
}

func NewUrlHandler(s *service.UrlService) *UrlHandler {
	return &UrlHandler{service: s}
}

func (h *UrlHandler)HandlerPost(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var userUrl models.UrlInfo

	err := json.NewDecoder(c.Request.Body).Decode(&userUrl)
	if err != nil {
		c.String(http.StatusBadRequest, "problem decoding json")
		return
	}
	newUrl, err := h.service.ServicePost(userUrl)
	if err != nil {
		c.String(http.StatusInternalServerError, "problem inserting url")
		return
	}
	c.IndentedJSON(http.StatusOK, newUrl)
}