package service

import (
	"github.com/gin-gonic/gin"
)

// service 接口
type User interface {
	Login(c *gin.Context)
	Register(c *gin.Context)
	LogOut(c *gin.Context)
}

type Upload interface {
	UploadFile(c *gin.Context)
	UploadChunk(c *gin.Context)
	MergeChunks(c *gin.Context)
}

type File interface {
	Download(c *gin.Context)
	Preview(c *gin.Context)
	GetList(c *gin.Context)
	Delete(c *gin.Context)
}
