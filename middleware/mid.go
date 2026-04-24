package middleware

import (
	"cloud-storage/common"
	"cloud-storage/utils"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		//c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// JWTAuth JWT认证中间件（应用到所有路由）
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于登录和注册接口，跳过认证
		if strings.HasPrefix(c.Request.URL.Path, "/user/") {
			c.Next()
			return
		}

		// 从请求头获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "未提供认证 token"))
			c.Abort()
			return
		}

		// 检查token格式（前端使用 Bearer ${token} 格式）
		parts := strings.SplitN(authHeader, " ", 2)
		var tokenStr string
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenStr = parts[1]
		} else {
			tokenStr = parts[0]
		}

		claims, err := utils.ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "无效的token"))
			c.Abort()
			return
		}

		if utils.ShouldRefreshToken(claims) {
			newToken, err := utils.RefreshToken(tokenStr)
			if err == nil {
				// 将新token设置到响应头
				c.Writer.Header().Set("Authorization", "Bearer "+newToken)
			}
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

// RateLimit 请求频率限制中间件（应用到所有路由）
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/upload/chunk") || strings.HasPrefix(path, "/upload/merge") {
			c.Next()
			return
		}
		key := common.RateLimitKey + ip
		count, err := common.RS.IncrWithExpire(context.Background(), key, time.Minute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, "请求频率限制失败"))
			c.Abort()
			return
		}
		if count <= 100 {
			c.Next()
			return
		}
		c.JSON(http.StatusTooManyRequests, common.Error(http.StatusTooManyRequests, "请求频率"))
		c.Abort()
	}
}
