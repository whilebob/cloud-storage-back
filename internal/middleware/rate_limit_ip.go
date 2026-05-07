package middleware

import (
	"CloudStorage/internal/common"
	"CloudStorage/internal/global"
	redisUtil "CloudStorage/internal/utils/redis"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// IpRateLimit IP限流中间件
// duration: 时间窗口
// maxCount: 最大允许次数
func IpRateLimit(duration time.Duration, maxCount int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("%s:%s", global.RedisKeyRateLimit, ip)

		currentCount, err := redisUtil.RDU.IncrWithExpire(c.Request.Context(), key, duration)
		if err != nil {
			common.Fail(c, common.CodeServerError, "系统异常")
			c.Abort()
			return
		}

		if currentCount > maxCount {
			common.TooManyRequests(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
