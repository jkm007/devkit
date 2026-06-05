package database

import (
	"context"
	"fmt"

	"backend-server/config"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(cfg config.RedisConfig) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接 Redis 失败: %w", err)
	}

	return nil
}

// GetRedis 获取 Redis 客户端
func GetRedis() *redis.Client {
	return rdb
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() {
	if rdb != nil {
		rdb.Close()
	}
}
