package task

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"backend-server/pkg/database"
	"backend-server/pkg/logger"
)

// TaskFunc 任务函数类型
type TaskFunc func(ctx context.Context) error

// Scheduler 任务调度器
type Scheduler struct {
	tasks []taskEntry
}

type taskEntry struct {
	name     string
	interval time.Duration
	fn       TaskFunc
}

// NewScheduler 创建调度器
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// AddTask 添加定时任务
func (s *Scheduler) AddTask(name string, interval time.Duration, fn TaskFunc) {
	s.tasks = append(s.tasks, taskEntry{
		name:     name,
		interval: interval,
		fn:       fn,
	})
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) {
	for _, t := range s.tasks {
		go s.runTask(ctx, t)
	}
}

func (s *Scheduler) runTask(ctx context.Context, t taskEntry) {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	logger.Info("启动定时任务", zap.String("task", t.name))

	for {
		select {
		case <-ctx.Done():
			logger.Info("停止定时任务", zap.String("task", t.name))
			return
		case <-ticker.C:
			if err := t.fn(ctx); err != nil {
				logger.Error("定时任务执行失败",
					zap.String("task", t.name),
					zap.Error(err),
				)
			}
		}
	}
}

// DelayTask 延迟任务（基于 Redis ZSet）
type DelayTask struct {
	queueKey string
}

// NewDelayTask 创建延迟任务
func NewDelayTask(queueKey string) *DelayTask {
	return &DelayTask{queueKey: queueKey}
}

// Push 推送延迟任务
func (d *DelayTask) Push(ctx context.Context, payload string, executeAt time.Time) error {
	rdb := database.GetRedis()
	return rdb.ZAdd(ctx, d.queueKey, redis.Z{
		Score:  float64(executeAt.Unix()),
		Member: payload,
	}).Err()
}

// Consume 消费延迟任务
func (d *DelayTask) Consume(ctx context.Context, handler func(payload string) error) {
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
				// 取出到期的任务
				results, err := rdb.ZRangeByScore(ctx, d.queueKey, &redis.ZRangeBy{
					Min:    "-inf",
					Max:    fmt.Sprintf("%d", now),
					Count:  10,
					Offset: 0,
				}).Result()
				if err != nil {
					logger.Error("获取延迟任务失败", zap.Error(err))
					continue
				}

				for _, payload := range results {
					// 先删除再处理，避免重复消费
					rdb.ZRem(ctx, d.queueKey, payload)
					if err := handler(payload); err != nil {
						logger.Error("延迟任务处理失败",
							zap.String("payload", payload),
							zap.Error(err),
						)
					}
				}
			}
		}
	}()
}
