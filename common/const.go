package common

import (
	"cloud-storage/utils/redis"
	"time"
)

// 跟 Redis 相关的常量
var (
	RS           = redis.NewSet()
	RateLimitKey = "rate_limit:"

	UserLoginAndRegisterRateLimit = "user:LoginAndRegisterRateLimit"
	UserRegisterKey               = "user:register"
	UserRegisterExpireTime        = 24 * time.Hour * 7
	UserLoginTokenKey             = "user:login:token"
	UserLoginTokenExpireTime      = 24 * time.Hour * 7

	FileMD5Key     = "file:md5"
	FileChunkKey   = "file:chunk:md5"
	LockChunkKey   = "lock:chunk"
	LockMergeKey   = "lock:merge"
	MergeStatusKey = "merge:status"
	FileStatusKey  = "file:status"

	DeleteLockKey = "lock:delete"
)

const (
	MinioBucketNameTemp   string = "temp"
	MinioBucketNameUpload string = "upload"
)
