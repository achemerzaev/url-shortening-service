package middleware

import (
	"github.com/boretsotets/url-shortening-service/internal/authorization"
	"github.com/boretsotets/url-shortening-service/internal/models"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"time"
	"net/http"
	"fmt"
	"strings"
	"strconv"
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
			Name: "http_request_duraion_seconds",
			Help: "Histogram of response time for handler in seconds",
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
		path := c.Request.URL.Path

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

func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// пре-процессинг
		start := time.Now()
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		method := c.Request.Method
		// другие мидлверы и хендлеры
		c.Next()

		// пост-процессинг
		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()
		reqID, _ := c.Get("requestID")

		logger.Info("request completed",
			zap.String("request_id", reqID.(string)),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Int("status", status),
			zap.Int("size", size),
			zap.Duration("latency", latency),
		)
	}
}

func AuthorizationMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		clientID, err := authorization.ValidateJWT(token)
		if err != nil {
			logger.Error("token validation error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, 
				gin.H{"error":"invalid access token"})
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
			v = &models.PostRequestJSON{}
			fmt.Println(">>> Correct case is choosed")
		case c.Request.Method == "PUT" && strings.HasPrefix(c.Request.URL.Path, "/shorten/"):
			v = &models.PutRequestJSON{}
		case c.Request.Method == "POST" && c.Request.URL.Path == "/register":
			v = &models.PostUserRegistration{}
		case c.Request.Method == "POST" && c.Request.URL.Path == "/login":
			v = &models.PostUserLogin{}
		case c.Request.Method == "POST" && strings.HasPrefix(c.Request.URL.Path, "/refresh"):
			v = &models.PostRefreshToken{}
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
				"error":"json validation error",
				"message":"invalid input"})
			return
		}

		c.Set("jsonBody", v)
		c.Next()
	}
}