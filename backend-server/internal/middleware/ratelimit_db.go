package middleware

import (
	"context"
	"fmt"
	"time"

	"backend-server/internal/service"
	"backend-server/pkg/database"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// DBRateLimiter 基于数据库规则的限流中间件
// 优先级高于全局限流，匹配到规则时使用规则的限流配置
func DBRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		cache := service.GetGlobalRuleCache()
		ip := c.ClientIP()
		path := c.Request.URL.Path
		method := c.Request.Method

		result := cache.MatchAndCheck(ip, path, method)
		if result.Matched && !result.Allowed {
			// 触发限流时，累加风险分
			if result.ViolationScore > 0 {
				addRateLimitViolationScore(ip, result.ViolationScore)
			}

			// 设置 Retry-After 头
			if result.RetryAfter > 0 {
				c.Header("Retry-After", fmt.Sprintf("%d", result.RetryAfter))
				c.Header("X-RateLimit-RetryAfter", fmt.Sprintf("%d", result.RetryAfter))
			}
			response.TooManyRequests(c, result.Reason)
			c.Abort()
			return
		}

		c.Next()
	}
}

// addRateLimitViolationScore 限流违规时累加风险分
func addRateLimitViolationScore(ip string, score int) {
	rdb := database.GetRedis()
	if rdb == nil {
		return
	}
	ctx := context.Background()
	key := riskScoreKeyPrefix + ip

	// 原子累加风险分
	rdb.IncrBy(ctx, key, int64(score))
	// 重置 TTL（30 分钟衰减）
	rdb.Expire(ctx, key, 30*time.Minute)
}
