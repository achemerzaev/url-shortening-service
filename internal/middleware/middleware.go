package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/google/uuid"

	"time"
)

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
			zap.String("method",method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Int("status", status),
			zap.Int("size", size),
			zap.Duration("latency", latency),
		)
	}
}