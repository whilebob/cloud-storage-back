package common

import (
	"cloud-storage/config"
	"cloud-storage/global"
	gormLogger "cloud-storage/gorm-logger"
	"cloud-storage/model"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitAll() {
	initMinio()
	initRedis()
	initDB()
	migrate()
}

func initMinio() {
	endpoint := fmt.Sprintf("%s:%d", config.AppConfig.Minio.Host, config.AppConfig.Minio.Port)
	client, err := minio.New(
		endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(config.AppConfig.Minio.AccessKey, config.AppConfig.Minio.SecretKey, ""),
			Secure: false, // 是否使用 HTTPS，本地默认 false
			Transport: &http.Transport{
				MaxIdleConns:        runtime.NumCPU() * 100, // 总最大空闲连接
				MaxIdleConnsPerHost: runtime.NumCPU() * 10,  // 每个 host 最大空闲连接
				IdleConnTimeout:     90 * time.Second,       // 空闲超时关闭
				TLSHandshakeTimeout: 10 * time.Second,       // TLS 握手超时
				DisableCompression:  true,                   // 对象存储不需要压缩
			},
		})
	if err != nil {
		log.Fatal(err)
	}
	global.Minio = client

	global.MinioCore, err = minio.NewCore(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AppConfig.Minio.AccessKey, config.AppConfig.Minio.SecretKey, ""),
		Secure: false, // 是否使用 HTTPS，本地默认 false
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := makeMinioBucket(); err != nil {
		panic(fmt.Sprintf("创建MinIO桶失败: %s", err.Error()))
	}
}

func initRedis() {
	client := redis.NewClient(&redis.Options{
		Password:        config.AppConfig.Redis.Password,
		PoolSize:        runtime.NumCPU() * 10,
		MinIdleConns:    runtime.NumCPU() * 2,
		Addr:            fmt.Sprintf("%s:%d", config.AppConfig.Redis.Host, config.AppConfig.Redis.Port),
		DB:              config.AppConfig.Redis.DB,
		DisableIdentity: true,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatal(err)
	}
	global.Redis = client
}

func initDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.AppConfig.MySQL.Username,
		config.AppConfig.MySQL.Password,
		config.AppConfig.MySQL.Host,
		config.AppConfig.MySQL.Port,
		config.AppConfig.MySQL.DBName,
	)
	db, err1 := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: false,
		Logger:                                   DBLogger,
		NamingStrategy:                           schema.NamingStrategy{
			//SingularTable: true,
			//TablePrefix:   "tb_",
		},
	})

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	// 设置连接池（高并发必须配置！）
	sqlDB.SetMaxOpenConns(runtime.NumCPU() * 10) // 最大打开连接
	sqlDB.SetMaxIdleConns(runtime.NumCPU() * 2)  // 最大空闲连接
	sqlDB.SetConnMaxLifetime(30 * time.Second)   // 连接复用时间
	sqlDB.SetConnMaxIdleTime(10 * time.Second)   // 空闲最长时间
	global.DB = db
	if err1 != nil {
		log.Fatal(err1)
	}

	tx := global.DB.Exec("SELECT 1")
	if tx.Error != nil {
		panic(tx.Error)
	}
}

func migrate() {
	global.DB.Exec("SET FOREIGN_KEY_CHECKS = 0")

	err := global.DB.AutoMigrate(&model.User{})

	if err != nil {
		panic("迁移表失败user：" + err.Error())
	}

	err = global.DB.AutoMigrate(&model.File{})
	if err != nil {
		panic("迁移表失败file：" + err.Error())
	}

	err = global.DB.AutoMigrate(&model.Chunk{})
	if err != nil {
		panic("迁移表失败chunk：" + err.Error())
	}
}

func makeMinioBucket() error {
	var bucketName = []string{
		MinioBucketNameTemp,
		MinioBucketNameUpload,
	}
	for _, bucket := range bucketName {
		err := global.Minio.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{
			ForceCreate: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

var DBLogger logger.Interface

func InitDBLogger() {
	file := mkLogger()
	log.Println(file)
	config := gormLogger.FileLogConfig{
		Config: logger.Config{
			SlowThreshold:             0,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		},
	}
	filename := gormLogger.PathOption{Name: "/logs/db.log"}
	DBLogger = gormLogger.NewLogger(config, filename)
}
func mkLogger() *os.File {
	wd, _ := os.Getwd()
	logDir := filepath.Join(wd, "logs")
	_ = os.Mkdir(logDir, 0755)

	logFilePath := filepath.Join(logDir, "db.log")
	file, _ := os.Open(logFilePath)
	return file
}
