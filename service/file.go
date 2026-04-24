package service

import (
	"cloud-storage/common"
	"cloud-storage/config"
	"cloud-storage/global"
	"cloud-storage/model"
	"cloud-storage/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

type FileService struct {
}

var _ File = (*FileService)(nil)

type FileDTO struct {
	FileName   string    `json:"file_name"` // 字段名和数据库列名一致
	CreateTime time.Time `json:"create_time"`
	//MinioURL   string    `json:"minio_url"`
	Size int64 `json:"size"`
}

func (f *FileService) Delete(c *gin.Context) {
	err, tokenStr := getString(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, err.Error()))
		return
	}
	username := utils.ParseTokenToUsername(tokenStr)
	fileName := c.Query("filename")
	ctx := context.Background()

	var file model.File
	err = global.DB.Model(&model.File{}).
		Where("username = ? AND file_name = ?", username, fileName).
		First(&file).Error
	if err != nil {
		c.JSON(http.StatusOK, common.ErrorWithCode(http.StatusInternalServerError, err.Error()))
		return
	}
	md5 := file.Md5

	//分布式锁
	lockKey := fmt.Sprintf("%s:%s:%s", common.DeleteLockKey, username, md5)
	locked, err := global.Redis.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
	if err != nil || !locked {
		c.JSON(http.StatusOK, common.Error(http.StatusInternalServerError, "文件正在删除中，请勿重复操作"))
		return
	}
	defer global.Redis.Del(ctx, lockKey)

	fileSize := file.Size
	const SmallFileSize = 1024 * 1024 * 10

	if fileSize > int64(SmallFileSize) {
		err := deleteLargeFile(username, fileName, md5)
		if err != nil {
			c.JSON(http.StatusOK, common.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		c.JSON(http.StatusOK, common.SuccessWithData("删除大文件成功"))
		return
	}

	delErr := deleteSmallFile(username, fileName, md5)
	if delErr != nil {
		c.JSON(http.StatusOK, common.Error(http.StatusInternalServerError, delErr.Error()))
		return
	}
	c.JSON(http.StatusOK, common.SuccessWithData("删除小文件成功"))
}

func (f *FileService) GetList(c *gin.Context) {
	err, tokenStr := getString(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, err.Error()))
		return
	}
	username := utils.ParseTokenToUsername(tokenStr)

	page := c.Query("page")
	size := c.Query("size")

	if page == "" || size == "" {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "分页参数不能为空"))
		return
	}
	pageNum, _ := strconv.Atoi(page)
	pageSize, _ := strconv.Atoi(size)
	if pageNum <= 0 || pageSize <= 0 {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "分页参数格式错误"))
		return
	}

	files, err := getFiles(username, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, common.SuccessWithData(files))
}

func (f *FileService) Download(c *gin.Context) {
	err, tokenStr := getString(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, err.Error()))
		return
	}
	username := utils.ParseTokenToUsername(tokenStr)
	fileName := c.Query("filename")

	finalURL, err := createTempURL(username, fileName, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.SuccessWithData(finalURL))
}

func (f *FileService) Preview(c *gin.Context) {
	err, tokenStr := getString(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, err.Error()))
		return
	}
	username := utils.ParseTokenToUsername(tokenStr)
	fileName := c.Query("filename")
	finalURL, err := createTempURL(username, fileName, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.SuccessWithData(finalURL))
}

func getFiles(username string, pageNum int, pageSize int) ([]FileDTO, error) {
	var fileDTOs []FileDTO

	err := global.DB.Model(&model.File{}).
		Where("username = ? AND is_uploaded = ?", username, true).
		Select("file_name, created_at AS create_time, size").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&fileDTOs).Error

	return fileDTOs, err
}

func deleteSmallFile(username, fileName, md5 string) error {
	ctx := context.Background()

	tx := global.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return errors.New("事务开启失败")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Delete(&model.File{}, "md5 = ? AND username = ?", md5, username).Error; err != nil {
		tx.Rollback()
		return errors.New("删除记录失败")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.New("事务提交失败")
	}

	deleteKey := fmt.Sprintf("%s:%s:%s", common.FileStatusKey, username, md5)
	if err := global.Redis.Del(ctx, deleteKey).Err(); err != nil {
		return errors.New("删除Redis标记失败")
	}

	objectName := fmt.Sprintf("%s/%s", username, fileName)
	err := global.Minio.RemoveObject(ctx, common.MinioBucketNameUpload, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return errors.New("删除MinIO文件失败")
	}

	return nil
}

func deleteLargeFile(username, fileName, md5 string) error {
	err := deleteSmallFile(username, fileName, md5)
	if err != nil {
		return err
	}
	deleteMergeKey := fmt.Sprintf("%s:%s:%s", common.MergeStatusKey, username, md5)
	if err := global.Redis.Del(context.Background(), deleteMergeKey).Err(); err != nil {
		return err
	}
	return nil
}

func createTempURL(username, filename string, preview bool) (string, error) {
	ctx := context.Background()

	objectName := fmt.Sprintf("%s/%s", username, filename)

	disposition := "attachment"
	if preview {
		disposition = "inline"
	}

	ext := strings.ToLower(filepath.Ext(filename))
	contentType := getContentType(ext)

	tempURL, err := global.Minio.PresignedGetObject(ctx,
		common.MinioBucketNameUpload,
		objectName,
		10*time.Minute,
		url.Values{
			"response-content-disposition": {disposition + "; filename=\"" + filename + "\""},
			"response-content-type":        {contentType},
		},
	)
	if err != nil {
		return "", errors.New("生成临时URL失败")
	}

	proxyPrefix := fmt.Sprintf("http://%s:%d/oss", config.AppConfig.HOST, config.AppConfig.PORT)
	finalURL := strings.ReplaceAll(tempURL.String(), "http://127.0.0.1:9000", proxyPrefix)
	finalURL = strings.ReplaceAll(finalURL, "http://localhost:9000", proxyPrefix)

	return finalURL, nil
}

func getContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".md":
		return "text/markdown"
	default:
		return "application/octet-stream"
	}
}
