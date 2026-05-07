package router

import (
	"CloudStorage/config"
	"CloudStorage/internal/handler"
	"CloudStorage/internal/middleware"
	"math"
	"time"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	userHandler := handler.NewUserHandler()

	gin.SetMode(config.AppConfig.Server.Mode)
	r := gin.Default()

	r.Use(middleware.GinLogger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.IpRateLimit(1*time.Minute, 100))

	api := r.Group("/api")

	user := api.Group("/user", middleware.IpRateLimit(5*time.Minute, int64(math.Pow10(6))))
	{
		user.POST("/register", userHandler.Register)
		user.POST("/login", userHandler.Login)
		user.GET("/logout", userHandler.Logout)
	}

	return r
}
