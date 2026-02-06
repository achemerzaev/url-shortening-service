package app

import (
	"github.com/achemerzaev/url-shortening-service/internal/middleware"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"
)

func SetupRouter(h Handlers, logger logger.Logger) *gin.Engine {
	router := gin.New()
	router.Use(middleware.RequestIdMiddleware())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.PrometheusMiddleware())
	router.Use(middleware.JSONValidationMiddleware())

	router.POST("/register", h.User.HandlerRegister)
	router.POST("/login", h.User.HandlerLogin)
	router.POST("/refresh", h.User.HandlerRefresh)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	private := router.Group("/")
	private.Use(middleware.AuthorizationMiddleware(logger))
	{
		private.POST("/shorten", h.URL.HandlerPost)
		private.GET("/shorten/:shortcode", h.URL.HandlerGet)
		private.PUT("/shorten/:shortcode", h.URL.HandlerPut)
		private.DELETE("/shorten/:shortcode", h.URL.HandlerDelete)
		private.GET("/shorten/:shortcode/stats", h.URL.HandlerGetStats)
	}

	return router
}
