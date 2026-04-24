package api

import (
	"cloud-storage/service"

	"github.com/gin-gonic/gin"
)

var uploadService *service.UploadService

func UploadRouter(r *gin.Engine) {

	upload := r.Group("/upload")
	{
		upload.POST("/file", uploadService.UploadFile)
		upload.POST("/chunk", uploadService.UploadChunk)
		upload.POST("/merge", uploadService.MergeChunks)
		upload.GET("/merge/status", uploadService.CheckMergeStatus)
	}
}
