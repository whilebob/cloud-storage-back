package redis

import (
	"cloud-storage/global"
	"context"
	"errors"
	"fmt"
	"time"
)

// Lock Redis 分布式锁
type Lock struct {
	key   string
	value string
	ttl   time.Duration
}

// NewLock 创建分布式锁实例
func NewLock(key string, ttl time.Duration) *Lock {
	return &Lock{
		key:   key,
		value: fmt.Sprintf("%d", time.Now().UnixNano()), // 使用时间戳作为唯一值
		ttl:   ttl,
	}
}

// Acquire 获取锁
func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	// 使用 SET NX 命令获取锁
	result, err := global.Redis.SetNX(ctx, l.key, l.value, l.ttl).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// Release 释放锁
func (l *Lock) Release(ctx context.Context) error {
	// 使用 Lua 脚本原子性释放锁，避免误释放
	script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `
	result, err := global.Redis.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		return err
	}
	if result.(int64) == 0 {
		return errors.New("lock not held")
	}
	return nil
}

// TryLock 尝试获取锁，带超时
func (l *Lock) TryLock(ctx context.Context, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		acquired, err := l.Acquire(ctx)
		if err != nil {
			return false, err
		}
		if acquired {
			return true, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false, errors.New("lock acquisition timed out")
}
