package redis

import (
	"cloud-storage/global"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Set Redis 集合操作
type Set struct {
	client *redis.Client
}

// NewSet 创建集合操作实例
func NewSet() *Set {
	return &Set{client: global.Redis}
}

// Add 向集合添加元素
func (s *Set) Add(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return global.Redis.SAdd(ctx, key, members...).Result()
}

// Remove 从集合移除元素
func (s *Set) Remove(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return global.Redis.SRem(ctx, key, members...).Result()
}

// Members 获取集合所有元素
func (s *Set) Members(ctx context.Context, key string) ([]string, error) {
	return global.Redis.SMembers(ctx, key).Result()
}

// Contains 检查元素是否在集合中
func (s *Set) Contains(ctx context.Context, key string, member interface{}) (bool, error) {
	result, err := global.Redis.SIsMember(ctx, key, member).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// Cardinality 获取集合大小
func (s *Set) Cardinality(ctx context.Context, key string) (int64, error) {
	return global.Redis.SCard(ctx, key).Result()
}

// Intersection 获取多个集合的交集
func (s *Set) Intersection(ctx context.Context, keys ...string) ([]string, error) {
	return global.Redis.SInter(ctx, keys...).Result()
}

// Union 获取多个集合的并集
func (s *Set) Union(ctx context.Context, keys ...string) ([]string, error) {
	return global.Redis.SUnion(ctx, keys...).Result()
}

// Difference 获取集合的差集
func (s *Set) Difference(ctx context.Context, key string, otherKeys ...string) ([]string, error) {
	keys := append([]string{key}, otherKeys...)
	return global.Redis.SDiff(ctx, keys...).Result()
}

func (s *Set) IncrWithExpire(ctx context.Context, key string, duration time.Duration) (int64, error) {
	count, err := global.Redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		s.Expire(ctx, key, duration)
	}

	return count, nil
}

func (s *Set) Expire(ctx context.Context, key string, duration time.Duration) {
	global.Redis.Expire(ctx, key, duration)
	return
}

func (s *Set) Get(ctx context.Context, key string) (string, error) {
	return global.Redis.Get(ctx, key).Result()
}

func (s *Set) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	return global.Redis.Set(ctx, key, value, duration).Err()
}
