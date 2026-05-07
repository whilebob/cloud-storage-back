package global

import (
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	Redis      *redis.Client
	DB         *gorm.DB
	Minio      *minio.Client
	MinioCore  *minio.Core
	AsyncLimit = make(chan struct{}, 20)
)
