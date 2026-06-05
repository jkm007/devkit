package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"backend-server/pkg/database"
)

// Set 设置缓存
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	rdb := database.GetRedis()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func Get(ctx context.Context, key string, dest interface{}) error {
	rdb := database.GetRedis()
	data, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete 删除缓存
func Delete(ctx context.Context, keys ...string) error {
	rdb := database.GetRedis()
	return rdb.Del(ctx, keys...).Err()
}

// Exists 检查缓存是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	rdb := database.GetRedis()
	result, err := rdb.Exists(ctx, key).Result()
	return result > 0, err
}

// SetHash 设置 Hash 缓存
func SetHash(ctx context.Context, key string, field string, value interface{}) error {
	rdb := database.GetRedis()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rdb.HSet(ctx, key, field, data).Err()
}

// GetHash 获取 Hash 缓存
func GetHash(ctx context.Context, key string, field string, dest interface{}) error {
	rdb := database.GetRedis()
	data, err := rdb.HGet(ctx, key, field).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// DeleteHash 删除 Hash 字段
func DeleteHash(ctx context.Context, key string, fields ...string) error {
	rdb := database.GetRedis()
	return rdb.HDel(ctx, key, fields...).Err()
}

// ZSetAdd 添加 ZSet 成员
func ZSetAdd(ctx context.Context, key string, score float64, member interface{}) error {
	rdb := database.GetRedis()
	data, err := json.Marshal(member)
	if err != nil {
		return err
	}
	return rdb.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: string(data),
	}).Err()
}

// ZSetRangeByScore 按分数范围获取 ZSet 成员
func ZSetRangeByScore(ctx context.Context, key string, min, max string) ([]string, error) {
	rdb := database.GetRedis()
	return rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
}

// ZSetRem 删除 ZSet 成员
func ZSetRem(ctx context.Context, key string, members ...interface{}) error {
	rdb := database.GetRedis()
	return rdb.ZRem(ctx, key, members...).Err()
}
