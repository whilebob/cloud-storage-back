package api

import (
	"cloud-storage/service"

	"github.com/gin-gonic/gin"
)

var fileService *service.FileService

func FileRouter(r *gin.Engine) {
	file := r.Group("/file")
	{
		file.DELETE("/delete", fileService.Delete)
		file.GET("/list", fileService.GetList)
		file.GET("/download", fileService.Download)
		file.GET("/preview", fileService.Preview)
	}
}
