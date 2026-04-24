package utils

import (
	"cloud-storage/common"
	"cloud-storage/config"
	"cloud-storage/global"
	"cloud-storage/model"
	"cloud-storage/utils/redis"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type MinioAndDB interface {
	FileToMinio(options FileUploadToMinioOptions) error
	SaveSmallFileToDB(options FileUploadToMinioOptions) error
}

type MinioAndDBStruct struct {
}

type FileUploadToMinioOptions struct {
	Username   string
	BucketName string
	ObjectName string
	FileHeader *multipart.FileHeader
	Context    *gin.Context
}

var (
	_     MinioAndDB = (*MinioAndDBStruct)(nil)
	queue            = redis.NewQueue()
)

func (m *MinioAndDBStruct) FileToMinio(options FileUploadToMinioOptions) error {
	bucketName := options.BucketName
	objectName := options.ObjectName
	file := options.FileHeader

	fileHandle, err := file.Open()
	defer fileHandle.Close()
	if err != nil {
		return err
	}

	size := file.Size

	_, err = global.Minio.PutObject(context.Background(), bucketName, objectName, fileHandle, size,
		minio.PutObjectOptions{
			//ContentType: file.Header.Get("Content-Type"),
			NumThreads: 4,
		})
	if err != nil {
		return err
	}

	return nil
}

func (m *MinioAndDBStruct) SaveSmallFileToDB(options FileUploadToMinioOptions) error {
	username := options.Username
	file := options.FileHeader
	c := options.Context

	fileName := file.Filename
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	//需要从配置文件加载 MinIO 域名
	minioURL := fmt.Sprintf("http://%s:%s/%s/%s/%s",
		config.AppConfig.Minio.Host,
		strconv.Itoa(config.AppConfig.Minio.Port),
		common.MinioBucketNameUpload,
		username,
		fileName,
	)

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		// 检查用户是否存在(一个用户多个文件)
		if err = tx.Model(&model.User{}).
			Where("username = ?", username).
			Error; err != nil {
			return errors.New("用户不存在")
		}
		// 保存文件到数据库
		if err = tx.Create(&model.File{
			FileName:   fileName,
			IsUploaded: true,
			Md5:        form.Value["md5"][0],
			Size:       file.Size,
			MinioURL:   minioURL,
			Username:   username,
		}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
