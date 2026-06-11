package middleware

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"backend-server/config"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter 关联限流器与最后访问时间
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64 // UnixNano，避免数据竞争
}

// RateLimiter 限流中间件（令牌桶算法）
func RateLimiter(ctx context.Context, cfg config.RateLimitConfig) gin.HandlerFunc {
	return NewIPRateLimiter(ctx, rate.Limit(cfg.Rate), cfg.Burst)
}

// NewIPRateLimiter 创建基于 IP 的限流中间件
func NewIPRateLimiter(ctx context.Context, r rate.Limit, burst int) gin.HandlerFunc {
	limiters := &sync.Map{}

	// 后台清理过期的 IP 记录，防止内存泄漏
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				now := time.Now()
				limiters.Range(func(key, value interface{}) bool {
					entry := value.(*ipLimiter)
					if now.Sub(time.Unix(0, entry.lastSeen.Load())) > 10*time.Minute {
						limiters.Delete(key)
					}
					return true
				})
			}
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		val, _ := limiters.LoadOrStore(ip, &ipLimiter{
			limiter: rate.NewLimiter(r, burst),
		})
		entry := val.(*ipLimiter)
		entry.lastSeen.Store(time.Now().UnixNano())

		if !entry.limiter.Allow() {
			response.TooManyRequests(c, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}
