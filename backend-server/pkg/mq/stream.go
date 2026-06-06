package mq

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"backend-server/pkg/database"
	"backend-server/pkg/logger"
)

// Stream 消息队列（基于 Redis Stream）
type Stream struct {
	streamKey string
	groupName string
}

// NewStream 创建 Stream 消息队列
func NewStream(streamKey, groupName string) *Stream {
	return &Stream{
		streamKey: streamKey,
		groupName: groupName,
	}
}

// Init 初始化消费者组
func (s *Stream) Init(ctx context.Context) error {
	rdb := database.GetRedis()
	// 创建消费者组（如果不存在）
	err := rdb.XGroupCreateMkStream(ctx, s.streamKey, s.groupName, "0").Err()
	if err != nil && err.Error() == "BUSYGROUP Consumer Group name already exists" {
		return nil
	}
	return err
}

// Publish 发布消息
func (s *Stream) Publish(ctx context.Context, values map[string]interface{}) (string, error) {
	rdb := database.GetRedis()
	return rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: s.streamKey,
		Values: values,
	}).Result()
}

// Subscribe 订阅消息
func (s *Stream) Subscribe(ctx context.Context, consumerName string, handler func(msg redis.XMessage) error) {
	rdb := database.GetRedis()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				results, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    s.groupName,
					Consumer: consumerName,
					Streams:  []string{s.streamKey, ">"},
					Count:    10,
					Block:    5 * time.Second,
				}).Result()
				if err != nil {
					if err != redis.Nil {
						logger.Error("读取消息失败", zap.Error(err))
					}
					continue
				}

				for _, stream := range results {
					for _, msg := range stream.Messages {
						if err := handler(msg); err != nil {
							logger.Error("处理消息失败",
								zap.String("id", msg.ID),
								zap.Error(err),
							)
							continue
						}
						// 确认消息
						rdb.XAck(ctx, s.streamKey, s.groupName, msg.ID)
					}
				}
			}
		}
	}()
}
