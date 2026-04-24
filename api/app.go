package api

import (
	"cloud-storage/middleware"
	"cloud-storage/utils"

	"github.com/gin-gonic/gin"
)

// RouterAll 全部路由
func RouterAll() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 全局中间件（顺序固定）
	r.Use(utils.GinLogger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())
	r.Use(middleware.JWTAuth())

	UserRouter(r)
	FileRouter(r)
	UploadRouter(r)
	return r
}
