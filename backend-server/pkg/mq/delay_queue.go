package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"backend-server/pkg/database"
	"backend-server/pkg/logger"
)

// DelayQueue 延迟队列（基于 Redis ZSet）
type DelayQueue struct {
	queueKey string
}

// NewDelayQueue 创建延迟队列
func NewDelayQueue(queueKey string) *DelayQueue {
	return &DelayQueue{queueKey: queueKey}
}

// Push 推送延迟消息
func (q *DelayQueue) Push(ctx context.Context, payload string, executeAt time.Time) error {
	rdb := database.GetRedis()
	return rdb.ZAdd(ctx, q.queueKey, redis.Z{
		Score:  float64(executeAt.Unix()),
		Member: payload,
	}).Err()
}

// Consume 消费延迟消息
func (q *DelayQueue) Consume(ctx context.Context, handler func(payload string) error) {
	rdb := database.GetRedis()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				now := time.Now().Unix()
				results, err := rdb.ZRangeByScore(ctx, q.queueKey, &redis.ZRangeBy{
					Min:   "-inf",
					Max:   fmt.Sprintf("%d", now),
					Count: 10,
				}).Result()
				if err != nil {
					logger.Error("获取延迟消息失败", zap.Error(err))
					continue
				}

				for _, payload := range results {
					if err := handler(payload); err != nil {
						logger.Error("延迟消息处理失败",
							zap.String("payload", payload),
							zap.Error(err),
						)
						continue
					}
					// 处理成功后才删除
					rdb.ZRem(ctx, q.queueKey, payload)
				}
			}
		}
	}()
}
