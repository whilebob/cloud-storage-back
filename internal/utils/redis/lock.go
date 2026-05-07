package redis

import (
	"CloudStorage/internal/global"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLock struct {
	key    string
	value  string
	expire time.Duration
	cli    *redis.Client
}

func NewRedisLock(key string, value string, expire time.Duration) *RedisLock {
	return &RedisLock{
		key:    key,
		value:  value,
		expire: expire,
		cli:    global.Redis,
	}
}
