package middleware

import (
	"github.com/achemerzaev/url-shortening-service/internal/authorization"
	"github.com/achemerzaev/url-shortening-service/internal/dto"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duraion_seconds",
			Help:    "Histogram of response time for handler in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		path := c.FullPath()

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, strconv.Itoa(c.Writer.Status())).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

func RequestIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()
		c.Set("requestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

func LoggerMiddleware(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// пре-процессинг
		// start := time.Now()
		// path := c.Request.URL.Path
		// clientIP := c.ClientIP()
		// method := c.Request.Method
		// другие мидлверы и хендлеры
		// c.Next()

		// пост-процессинг
		// latency := time.Since(start)
		// status := c.Writer.Status()
		// size := c.Writer.Size()
		// reqID, _ := c.Get("requestID")

		// logger.Info("request completed",
		// 	"request_id: ", reqID.(string),
		// 	"method: ", method,
		// 	"path: ", path,
		// 	"ip: ", clientIP,
		// 	"status: ", status,
		// 	"size: ", size,
		// 	"latency: ", latency,
		// )
	}
}

func AuthorizationMiddleware(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		parts := strings.SplitN(token, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			logger.Error("Authorization header invalid format")
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "invalid access token"})
			return
		}

		clientID, err := authorization.ValidateJWT(parts[1])
		if err != nil {
			logger.Error("token validation error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "invalid access token"})
			return
		}
		c.Set("clientID", clientID)
		c.Next()
	}
}

func JSONValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var v interface{}

		switch {
		case c.Request.Method == "POST" && c.Request.URL.Path == "/shorten":
			v = &dto.HandlerPostRequest{}
		case c.Request.Method == "PUT" && strings.HasPrefix(c.Request.URL.Path, "/shorten/"):
			v = &dto.HandlerPutRequest{}
		case c.Request.Method == "POST" && c.Request.URL.Path == "/register":
			v = &dto.HandlerRegisterRequest{}
		case c.Request.Method == "POST" && c.Request.URL.Path == "/login":
			v = &dto.HandlerLoginRequest{}
		case c.Request.Method == "POST" && strings.HasPrefix(c.Request.URL.Path, "/refresh"):
			v = &dto.HandlerRefreshRequest{}
		default:
			c.Next()
			return
		}

		/*

			dec := json.NewDecoder(c.Request.Body)
			dec.DisallowUnknownFields()
			if err := dec.Decode(v); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":"json validation error",
						"message":"invalid input"})
					return
			}
		*/

		if err := c.ShouldBindJSON(v); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{
					"error":   "json validation error",
					"message": "invalid input"})
			return
		}

		c.Set("jsonBody", v)
		c.Next()
	}
}
