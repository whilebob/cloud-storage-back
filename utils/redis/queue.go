package redis

import (
	"cloud-storage/global"
	"context"
	"time"
)

// Queue Redis 消息队列操作
type Queue struct {
	Key string
}

// NewQueue 创建消息队列操作实例
func NewQueue() *Queue {
	return &Queue{}
}

// Enqueue 消息入队（使用 LPUSH）
func (q *Queue) Enqueue(ctx context.Context, key string, value interface{}) error {
	return global.Redis.LPush(ctx, key, value).Err()
}

// Dequeue 消息出队（使用 RPOP）
func (q *Queue) Dequeue(ctx context.Context, key string) (string, error) {
	result, err := global.Redis.RPop(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

// BlockingDequeue 阻塞式消息出队（使用 BLPOP）
func (q *Queue) BlockingDequeue(ctx context.Context, key string, timeout time.Duration) (interface{}, error) {
	result, err := global.Redis.BLPop(ctx, timeout, key).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Length 获取队列长度
func (q *Queue) Length(ctx context.Context, key string) (int64, error) {
	return global.Redis.LLen(ctx, key).Result()
}

// Clear 清空队列
func (q *Queue) Clear(ctx context.Context, key string) error {
	return global.Redis.Del(ctx, key).Err()
}

// ProcessMessages 异步处理消息队列中的消息
func (q *Queue) ProcessMessages(ctx context.Context, key string, processor func(interface{}) error) error {
	go func() {
		for {
			msg, err := q.BlockingDequeue(ctx, key, 0) // 0表示无限期阻塞
			if err != nil {
				// 处理错误，例如日志记录
				continue
			}

			// 处理消息
			err = processor(msg)
			if err != nil {
				// 处理错误，例如日志记录或重新入队
			}
		}
	}()
	return nil
}

// ProcessMessagesWithConcurrency 并发处理消息队列中的消息
func (q *Queue) ProcessMessagesWithConcurrency(ctx context.Context, key string, concurrency int, processor func(interface{}) error) error {
	for i := 0; i < concurrency; i++ {
		err := q.ProcessMessages(ctx, key, processor)
		if err != nil {
			return err
		}
	}
	return nil
}
