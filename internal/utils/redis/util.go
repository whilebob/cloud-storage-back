package redis

import (
	"CloudStorage/internal/global"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisUtil struct {
	cli *redis.Client
}

var (
	RDU *RedisUtil
	ctx = context.Background()
)

func NewRedisUtil() *RedisUtil {
	return &RedisUtil{
		cli: global.Redis,
	}
}

func (r *RedisUtil) IncrWithExpire(ctx context.Context, key string, expire time.Duration) (int64, error) {
	script := `
		local count = redis.call("INCR", KEYS[1])
		redis.call("EXPIRE", KEYS[1], ARGV[1])
		return count
	`
	var keys []string
	keys = append(keys, key)
	val, err := r.cli.Eval(ctx, script, keys, int(expire.Seconds())).Result()
	if err != nil {
		return 0, err
	}
	return val.(int64), nil
}

func (r *RedisUtil) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	script := `
		return redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2])
	`
	_, err := r.cli.Eval(ctx, script, []string{key}, value, int(expire.Seconds())).Result()
	return err
}

func (r *RedisUtil) IncrBy(ctx context.Context, key string, increment int64) error {
	return r.cli.IncrBy(ctx, key, increment).Err()
}

func (r *RedisUtil) Get(ctx context.Context, key string) (string, error) {
	return r.cli.Get(ctx, key).Result()
}

func (r *RedisUtil) Del(ctx context.Context, key string) error {
	return r.cli.Del(ctx, key).Err()
}

func (r *RedisUtil) Expire(ctx context.Context, key string, expire time.Duration) error {
	return r.cli.Expire(ctx, key, expire).Err()
}

func (r *RedisUtil) Exist(ctx context.Context, key string) (bool, error) {
	count, err := r.cli.Exists(ctx, key).Result()
	return count > 0, err
}
