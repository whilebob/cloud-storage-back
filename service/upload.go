package service

import (
	"cloud-storage/common"
	"cloud-storage/global"
	"cloud-storage/model"
	"cloud-storage/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type UploadService struct {
}

var _ Upload = (*UploadService)(nil)
var minioAndDB utils.MinioAndDB = &utils.MinioAndDBStruct{}

func (up *UploadService) UploadFile(c *gin.Context) {
	// 1. 校验token
	err, tokenStr := getString(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "无效的 token"))
		c.Abort()
		return
	}
	username := utils.ParseTokenToUsername(tokenStr)

	// 2. 获取文件与MD5
	file, err := c.FormFile("file")
	md5 := c.PostForm("md5")
	if md5 == "" {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "md5不能为空"))
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorWithCode(http.StatusBadRequest, "文件获取失败"))
		return
	}

	// 3. 秒传校验
	key := fmt.Sprintf("%s:%s:%s", common.FileMD5Key, username, md5)
	exist, err := global.Redis.Exists(context.Background(), key).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorWithCode(http.StatusInternalServerError, err.Error()))
		return
	}
	if exist == 1 {
		c.JSON(http.StatusFound, common.Error(http.StatusFound, "文件已存在"))
		return
	}

	objectName := fmt.Sprintf("%s/%s", username, file.Filename)

	minioErr := minioAndDB.FileToMinio(utils.FileUploadToMinioOptions{
		BucketName: common.MinioBucketNameUpload,
		ObjectName: objectName,
		FileHeader: file,
		Context:    c,
	})
	if minioErr != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorWithCode(http.StatusInternalServerError, minioErr.Error()))
		return
	}

	dbErr := minioAndDB.SaveSmallFileToDB(utils.FileUploadToMinioOptions{
		Username:   username,
		FileHeader: file,
		Context:    c,
	})
	if dbErr != nil {
		global.Minio.RemoveObject(context.Background(), common.MinioBucketNameUpload, objectName, minio.RemoveObjectOptions{})
		c.JSON(http.StatusInternalServerError, common.ErrorWithCode(http.StatusInternalServerError, dbErr.Error()))
		return
	}

	go func() {
		global.Redis.Set(context.Background(), key, md5, 7*24*time.Hour)
	}()

	c.JSON(http.StatusOK, common.SuccessWithData(gin.H{
		"message":  "文件上传成功",
		"filename": file.Filename,
	}))
}

func (up *UploadService) UploadChunk(c *gin.Context) {
	err, tokenStr := getString(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "无效的 token"))
		c.Abort()
		return
	}
	username := utils.ParseTokenToUsername(tokenStr)
	if username == "" {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "无效的 token"))
		c.Abort()
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "文件解析失败"))
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "文件解析失败"))
		return
	}
	fileHeader := form.File["file"][0]

	fileMd5 := form.Value["md5"][0]
	chunkMd5 := form.Value["chunkMd5"][0]
	chunkIndex := form.Value["chunkIndex"][0]
	totalChunks := form.Value["totalChunks"][0]

	if chunkIndex == "" || totalChunks == "" {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "参数不能为空"))
		return
	}
	chunkIndexInt, err := strconv.Atoi(chunkIndex)
	if err != nil || chunkIndexInt < 0 {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "分片索引格式错误"))
		return
	}
	totalChunksInt, err := strconv.Atoi(totalChunks)
	if err != nil || chunkIndexInt >= totalChunksInt {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "分片超出范围"))
		return
	}

	// 整个文件是否已上传（Redis 为准）
	ctx := c.Request.Context()
	fileStatusKey := fmt.Sprintf("%s:%s:%s", common.FileStatusKey, username, fileMd5)
	fileStatus, _ := global.Redis.Get(ctx, fileStatusKey).Result() //不存在才执行下面的逻辑
	//if err != nil {
	//	log.Printf("155行Redis获取文件状态失败: %v", err)
	//	c.JSON(http.StatusInternalServerError, common.Error(http.StatusInternalServerError, err.Error()))
	//	return
	//}
	if fileStatus == "1" {
		c.JSON(http.StatusOK, common.SuccessWithData("整个文件已上传"))
		return
	}

	chunkKey := fmt.Sprintf("%s:%s:%s", common.FileChunkKey, username, fileMd5)
	chunkField := fmt.Sprintf("chunk_%d", chunkIndexInt)
	objectName := fmt.Sprintf("%s/chunks/%s/part_%d", username, fileMd5, chunkIndexInt)

	//检查分片
	exist, _ := global.Redis.HExists(ctx, chunkKey, chunkField).Result()
	if exist {
		log.Printf("165行Redis检查分片是否存在失败: %v", err)
		c.JSON(http.StatusOK, common.SuccessWithData("分片已上传"))
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("异步上传panic: %v", r)
			}
		}()

		// 协程内部必须用独立上下文
		uploadCtx := context.Background()

		// 并发控制
		global.AsyncLimit <- struct{}{}
		defer func() { <-global.AsyncLimit }()

		// 分布式锁, 防止重复上传
		lockKey := fmt.Sprintf("%s:%s:%s:%d", common.LockChunkKey, username, fileMd5, chunkIndexInt)
		lockSuccess, _ := global.Redis.SetNX(uploadCtx, lockKey, "1", 30*time.Second).Result()
		if !lockSuccess {
			return
		}
		defer global.Redis.Del(uploadCtx, lockKey)

		err := minioAndDB.FileToMinio(utils.FileUploadToMinioOptions{
			Username:   username,
			BucketName: common.MinioBucketNameTemp,
			ObjectName: objectName,
			FileHeader: fileHeader,
			Context:    c,
		})
		if err != nil {
			log.Printf("MinIO上传失败: %v", err)
			// 上传失败 → 不写 Redis → 保证一致
			return
		}

		pipe := global.Redis.Pipeline()
		pipe.HSet(uploadCtx, chunkKey, chunkField, chunkMd5)
		pipe.Expire(uploadCtx, chunkKey, 7*24*time.Hour)
		_, err = pipe.Exec(uploadCtx)
		if err != nil {
			// Redis 失败 → 删除 MinIO 文件
			_ = global.Minio.RemoveObject(uploadCtx, common.MinioBucketNameTemp, objectName, minio.RemoveObjectOptions{})
			return
		}

		// 设置合并状态
		mergeStatusKey := fmt.Sprintf("%s:%s:%s", common.MergeStatusKey, username, fileMd5)
		_ = global.Redis.SetEx(uploadCtx, mergeStatusKey, "waiting", 7*24*time.Hour).Err()

		log.Printf("✅ 异步上传成功: %s %d", username, chunkIndexInt)
	}()

	// 立即返回前端（异步不阻塞）
	c.JSON(http.StatusOK, common.SuccessWithData("分片已接收，后台上传中"))
}

func (up *UploadService) MergeChunks(c *gin.Context) {
	err, tokenStr := getString(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "无效的 token"))
		c.Abort()
		return
	}
	username := utils.ParseTokenToUsername(tokenStr)
	if username == "" {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "无效的 token"))
		c.Abort()
		return
	}

	var req struct {
		Md5         string `json:"md5"`
		FileName    string `json:"filename"`
		TotalChunks int    `json:"totalChunks"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "参数错误: "+err.Error()))
		return
	}

	if req.Md5 == "" || req.TotalChunks == 0 || req.FileName == "" {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "参数不能为空"))
		return
	}

	fileMd5 := req.Md5
	fileName := req.FileName
	totalChunksInt := req.TotalChunks
	ctx := c.Request.Context()
	chunkKey := fmt.Sprintf("%s:%s:%s", common.FileChunkKey, username, fileMd5)

	// 1. 分布式锁：保证同一时间只能有一个合并在执行
	lockKey := fmt.Sprintf("%s:%s:%s", common.LockMergeKey, username, fileMd5)
	locked, _ := global.Redis.SetNX(ctx, lockKey, "1", 3*time.Minute).Result()
	if !locked {
		c.JSON(http.StatusOK, common.SuccessWithData("合并中，请稍后查询"))
		return
	}

	// 2. 异步合并（完全不阻塞前端）
	go func() {
		// 协程内部独立上下文
		mergeCtx := context.Background()
		mergeStatusKey := fmt.Sprintf("%s:%s:%s", common.MergeStatusKey, username, fileMd5)
		defer global.Redis.Del(mergeCtx, lockKey) // 合并结束释放锁

		// 标记合并中
		global.Redis.SetEx(mergeCtx, mergeStatusKey, "merging", 30*time.Minute)

		// panic 捕获
		defer func() {
			if r := recover(); r != nil {
				log.Printf("合并panic: %v", r)
				global.Redis.SetEx(mergeCtx, mergeStatusKey, "fail", 24*time.Hour)
			}
		}()

		// 并发控制
		global.AsyncLimit <- struct{}{}
		defer func() { <-global.AsyncLimit }()

		tmpBucket := common.MinioBucketNameTemp
		formalBucket := common.MinioBucketNameUpload
		dstObject := fmt.Sprintf("%s/%s", username, fileName)

		// 开始合并逻辑
		uploadID, err := global.MinioCore.NewMultipartUpload(mergeCtx, formalBucket, dstObject, minio.PutObjectOptions{})
		if err != nil {
			log.Printf("创建分片上传失败: %v", err)
			global.Redis.SetEx(mergeCtx, mergeStatusKey, "fail", 24*time.Hour)
			return
		}

		var completedParts []minio.CompletePart

		for i := 0; i < totalChunksInt; i++ {
			chunkObject := fmt.Sprintf("%s/chunks/%s/part_%d", username, fileMd5, i)

			reader, err := global.Minio.GetObject(mergeCtx, tmpBucket, chunkObject, minio.GetObjectOptions{})
			if err != nil {
				log.Printf("获取分片失败: %v", err)
				_ = global.MinioCore.AbortMultipartUpload(mergeCtx, formalBucket, dstObject, uploadID)
				global.Redis.SetEx(mergeCtx, mergeStatusKey, "fail", 24*time.Hour)
				return
			}

			stat, err := reader.Stat()
			if err != nil {
				_ = reader.Close()
				_ = global.MinioCore.AbortMultipartUpload(mergeCtx, formalBucket, dstObject, uploadID)
				global.Redis.SetEx(mergeCtx, mergeStatusKey, "fail", 24*time.Hour)
				return
			}

			part, err := global.MinioCore.PutObjectPart(
				mergeCtx,
				formalBucket,
				dstObject,
				uploadID,
				i+1,
				reader,
				stat.Size,
				minio.PutObjectPartOptions{},
			)
			_ = reader.Close()

			if err != nil {
				log.Printf("PutObjectPart失败: %v", err)
				_ = global.MinioCore.AbortMultipartUpload(mergeCtx, formalBucket, dstObject, uploadID)
				global.Redis.SetEx(mergeCtx, mergeStatusKey, "fail", 24*time.Hour)
				return
			}

			completedParts = append(completedParts, minio.CompletePart{
				PartNumber: i + 1,
				ETag:       part.ETag,
			})
		}

		// 完成合并
		_, err = global.MinioCore.CompleteMultipartUpload(
			mergeCtx,
			formalBucket,
			dstObject,
			uploadID,
			completedParts,
			minio.PutObjectOptions{},
		)
		if err != nil {
			log.Printf("合并失败: %v", err)
			global.Redis.SetEx(mergeCtx, mergeStatusKey, "fail", 24*time.Hour)
			return
		}

		// 合并成功
		global.Redis.SetEx(mergeCtx, mergeStatusKey, "success", 24*time.Hour)

		// 清理临时文件
		for i := 0; i < totalChunksInt; i++ {
			chunkObject := fmt.Sprintf("%s/chunks/%s/part_%d", username, fileMd5, i)
			_ = global.Minio.RemoveObject(mergeCtx, tmpBucket, chunkObject, minio.RemoveObjectOptions{})
		}
		_ = global.Redis.Del(mergeCtx, chunkKey).Err() // 删除分片记录

		fileStatusKey := fmt.Sprintf("%s:%s:%s", common.FileStatusKey, username, fileMd5)
		_, _ = global.Redis.Set(mergeCtx, fileStatusKey, "1", -1).Result() // 标记文件已上传, 永不过期

		//从minio获取文件大小
		stat, err := global.MinioCore.StatObject(mergeCtx, formalBucket, dstObject, minio.StatObjectOptions{})
		if err != nil {
			log.Printf("获取文件大小失败: %v", err)
			return
		}
		global.DB.Transaction(func(tx *gorm.DB) error {
			//先查用户是否存在
			if err := tx.Model(&model.User{}).Where("username = ?", username).Error; err != nil {
				return err
			}
			minioURL := fmt.Sprintf("%s/%s", username, fileName)
			// 再创建文件记录
			fileModel := &model.File{
				FileName:   fileName,
				IsUploaded: true,
				Md5:        fileMd5,
				MinioURL:   minioURL,
				Username:   username,
				Size:       stat.Size,
			}
			return tx.Create(fileModel).Error
		})

		log.Printf("✅ 用户 %s 文件 %s 合并成功", username, fileName)
	}()

	c.JSON(http.StatusOK, common.SuccessWithData("合并任务已开始，后台处理中"))
}

func (up *UploadService) CheckMergeStatus(c *gin.Context) {
	err, tokenStr := getString(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "无效的 token"))
		c.Abort()
		return
	}
	username := utils.ParseTokenToUsername(tokenStr)
	if username == "" {
		c.JSON(http.StatusUnauthorized, common.Error(http.StatusUnauthorized, "无效的 token"))
		c.Abort()
		return
	}

	fileMd5 := c.Query("md5")
	if fileMd5 == "" {
		c.JSON(http.StatusBadRequest, common.Error(http.StatusBadRequest, "md5 不能为空"))
		return
	}

	mergeStatusKey := fmt.Sprintf("%s:%s:%s", common.MergeStatusKey, username, fileMd5)
	status, _ := global.Redis.Get(c.Request.Context(), mergeStatusKey).Result()

	c.JSON(http.StatusOK, common.SuccessWithData(map[string]interface{}{
		"status": status,
		"msg":    getStatusMsg(status),
	}))
}

// 状态文案
func getStatusMsg(status string) string {
	switch status {
	case "merging":
		return "文件合并中..."
	case "success":
		return "合并完成"
	case "fail":
		return "合并失败"
	case "waiting":
		return "合并等待中..."
	default:
		return "未查询到合并任务"
	}
}
func getString(c *gin.Context) (error, string) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return errors.New("未提供认证 token"), ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	var tokenStr string
	if len(parts) == 2 && parts[0] == "Bearer" {
		tokenStr = parts[1]
	} else {
		tokenStr = parts[0]
	}
	return nil, tokenStr
}
