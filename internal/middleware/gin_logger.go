package middleware

import (
	"CloudStorage/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		ip := c.ClientIP()

		c.Next()

		cost := time.Since(start)
		status := c.Writer.Status()

		if status >= 400 {
			// 获取错误信息
			errMsg := ""
			if err := c.Errors.Last(); err != nil {
				errMsg = err.Error()
			}

			utils.Logger.Error("接口请求失败",
				zap.String("path", path),
				zap.String("method", method),
				zap.String("ip", ip),
				zap.Int("status", status),
				zap.Duration("cost", cost),
				zap.String("error", errMsg),
			)
		} else {
			// 记录正常访问日志
			utils.Logger.Info("访问接口",
				zap.String("path", path),
				zap.String("method", method),
				zap.String("ip", ip),
				zap.Int("status", status),
				zap.Duration("cost", cost),
			)
		}
	}
}
