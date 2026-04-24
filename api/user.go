package api

import (
	"cloud-storage/common"
	"cloud-storage/service"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var userService *service.UserService

func UserRouter(r *gin.Engine) {
	user := r.Group("/user", rateLimit(250, 10*time.Minute))
	{
		user.POST("/register", userService.Register)
		user.POST("/login", userService.Login)
		user.POST("/logout", userService.LogOut)
	}
}

func rateLimit(count int, min time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limitKey := fmt.Sprintf("%s:%s", common.UserLoginAndRegisterRateLimit, ip)

		// 1. 自增计数
		incr, err := common.RS.IncrWithExpire(context.Background(), limitKey, min)
		if err != nil {
			c.JSON(500, gin.H{"msg": "服务器限流异常"})
			c.Abort()
			return
		}

		// 2. 判断是否超过限制
		if incr > int64(count) {
			// 直接拦截，返回限流提示
			c.JSON(http.StatusTooManyRequests, common.ErrorWithCode(http.StatusTooManyRequests, "请求过于频繁，请稍后再试"))
			c.Abort() // 必须 abort，阻止进入接口
			return
		}

		// 3. 没超限，放行
		c.Next()
	}
}

func getIP(c *gin.Context) string {
	ip := c.GetHeader("X-Test-IP")
	if ip == "" {
		ip = "127.0.0.1"
	}
	return ip
}
