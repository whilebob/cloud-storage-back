package utils

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinLogger 自动日志中间件（所有接口自动打印）
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		// 请求接口
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		cost := time.Since(start)
		// 状态码
		status := c.Writer.Status()
		ip := c.ClientIP()

		// 自动打印接口日志
		Logger.Info("API_REQUEST",
			zap.String("path", path),
			zap.String("method", method),
			zap.Int("status", status),
			zap.String("ip", ip),
			zap.Duration("cost", cost),
		)
	}
}
