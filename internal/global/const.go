package global

import "time"

// Minio 桶名称
const (
	MinioBucketNameTemp   = "temp"
	MinioBucketNameUpload = "upload"
)

// Redis 相关前缀
const (
	RedisKeyRateLimit          = "rate_limit:"
	RedisRegisterKeyUser       = "register_user:"
	RedisRegisterKeyUserExpire = "register_user_expire:"
)

const (
	RedisRegisterKeyUserExpireTime = 5 * time.Minute //应该设置为用不过期的时间
)
