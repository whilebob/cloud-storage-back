package middleware

import (
	"CloudStorage/internal/common"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			common.Unauthorized(c)
			c.Abort()
			return
		}

		// 验证 Bearer token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			common.Unauthorized(c)
			c.Abort()
			return
		}

		tokenStr := parts[1]
		_, err := common.VerifyAccessToken(tokenStr)
		if err != nil {
			common.Unauthorized(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
