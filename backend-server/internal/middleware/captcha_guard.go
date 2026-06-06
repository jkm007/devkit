package middleware

import (
	"context"
	"strconv"
	"time"

	"backend-server/pkg/captcha"
	"backend-server/pkg/database"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

const riskScoreKeyPrefix = "risk:score:"

// CaptchaGuard 风险评分中间件
// 对受保护路径进行风险评估，高风险请求要求验证码
func CaptchaGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := getRiskCfg()

		if !cfg.Enabled {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		// 跳过认证相关路径（它们有自己的验证码校验）
		if len(path) >= 6 && path[:6] == "/auth/" {
			c.Next()
			return
		}

		// 只对受保护路径生效
		if !cfg.IsProtectedPath(path) {
			c.Next()
			return
		}

		ip := c.ClientIP()

		// 构建请求头 map
		headers := map[string]string{
			"Referer":          c.GetHeader("Referer"),
			"Accept-Language":  c.GetHeader("Accept-Language"),
			"User-Agent":       c.GetHeader("User-Agent"),
		}

		// 计算本次请求的风险分
		requestScore := calculateRiskScore(ip, path, headers, cfg)

		// 累加到 Redis 中的总分
		totalScore := accumulateRiskScore(ip, requestScore, cfg)

		// 检查是否需要直接拦截
		if cfg.BlockScore > 0 && totalScore >= cfg.BlockScore {
			response.Forbidden(c, "请求已被拦截，请稍后再试")
			c.Abort()
			return
		}

		// 检查是否需要验证码
		if totalScore < cfg.TriggerScore {
			c.Next()
			return
		}

		// 高风险：检查请求头中是否携带验证码
		captchaID := c.GetHeader("X-Captcha-Id")
		captchaCode := c.GetHeader("X-Captcha-Code")

		if captchaID == "" || captchaCode == "" {
			response.CaptchaRequired(c, "当前操作需要验证码验证")
			c.Abort()
			return
		}

		// 验证验证码
		valid, _ := captcha.Verify(captchaID, captchaCode, 0, nil)
		if !valid {
			response.CaptchaRequired(c, "验证码错误或已过期")
			c.Abort()
			return
		}

		// 验证通过，清零风险分
		clearRiskScore(ip)
		c.Next()
	}
}

// RiskConfigGetter 用于获取配置的函数类型（避免循环依赖）
type RiskConfigGetter struct {
	Enabled      bool
	TriggerScore int
	BlockScore   int
	DecayMinutes int
	DecayRate    float64
	Paths        []string
	Rules        []RiskRuleItem
}

// RiskRuleItem 规则项
type RiskRuleItem struct {
	Key       string
	Enabled   bool
	Score     int
	Threshold int
	Keywords  []string
}

// getRiskCfg 从 service 包获取配置（通过函数变量避免循环依赖）
var getRiskCfg func() *RiskConfigGetter

// SetRiskConfigGetter 设置配置获取函数（由 main.go 调用）
func SetRiskConfigGetter(fn func() *RiskConfigGetter) {
	getRiskCfg = fn
}

// IsProtectedPath 检查路径是否需要风险评估
// 匹配规则：路径以配置的前缀开头，且前缀后面紧跟 / 或路径结束
func (c *RiskConfigGetter) IsProtectedPath(path string) bool {
	for _, prefix := range c.Paths {
		if len(path) < len(prefix) {
			continue
		}
		if path[:len(prefix)] != prefix {
			continue
		}
		// 路径刚好等于前缀，或前缀后面是 /
		if len(path) == len(prefix) || path[len(prefix)] == '/' {
			return true
		}
	}
	return false
}

// calculateRiskScore 计算风险分数（在中间件内部实现，避免循环依赖）
func calculateRiskScore(ip, path string, headers map[string]string, cfg *RiskConfigGetter) int {
	score := 0

	for _, rule := range cfg.Rules {
		if !rule.Enabled {
			continue
		}

		switch rule.Key {
		case "frequency":
			score += evalFrequencyRule(rule, ip)
		case "no_referer":
			if headers["Referer"] == "" {
				score += rule.Score
			}
		case "no_lang":
			if headers["Accept-Language"] == "" {
				score += rule.Score
			}
		case "ua":
			score += evalUARule(rule, headers["User-Agent"])
		case "interval":
			score += evalIntervalRule(rule, ip)
		}
	}

	return score
}

// accumulateRiskScore 累加风险分到 Redis
func accumulateRiskScore(ip string, addScore int, cfg *RiskConfigGetter) int {
	rdb := database.GetRedis()
	ctx := context.Background()
	key := riskScoreKeyPrefix + ip

	// 获取当前分数
	current, _ := rdb.Get(ctx, key).Int()

	// 衰减：如果距离上次更新超过 decayMinutes，按比例衰减
	if cfg.DecayMinutes > 0 && cfg.DecayRate > 0 {
		ttl := rdb.TTL(ctx, key).Val()
		if ttl > 0 {
			totalTTL := time.Duration(cfg.DecayMinutes) * time.Minute
			elapsed := totalTTL - ttl
			decayPeriods := int(elapsed.Minutes()) / cfg.DecayMinutes
			if decayPeriods > 0 {
				for i := 0; i < decayPeriods; i++ {
					current = int(float64(current) * (1 - cfg.DecayRate))
				}
			}
		}
	}

	// 累加
	total := current + addScore

	// 存入 Redis，重置 TTL
	rdb.Set(ctx, key, total, time.Duration(cfg.DecayMinutes)*time.Minute)

	return total
}

// clearRiskScore 清零风险分
func clearRiskScore(ip string) {
	rdb := database.GetRedis()
	ctx := context.Background()
	rdb.Del(ctx, riskScoreKeyPrefix+ip)
}

// evalFrequencyRule 频率检测
func evalFrequencyRule(rule RiskRuleItem, ip string) int {
	rdb := database.GetRedis()
	ctx := context.Background()
	key := "risk:freq:" + ip

	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0
	}
	if count == 1 {
		rdb.Expire(ctx, key, 60*time.Second)
	}

	if int(count) > rule.Threshold {
		return rule.Score
	}
	return 0
}

// evalUARule UA 异常检测
func evalUARule(rule RiskRuleItem, ua string) int {
	if ua == "" {
		return rule.Score
	}
	for _, kw := range rule.Keywords {
		if containsIgnoreCase(ua, kw) {
			return rule.Score
		}
	}
	return 0
}

// evalIntervalRule 请求间隔检测
func evalIntervalRule(rule RiskRuleItem, ip string) int {
	rdb := database.GetRedis()
	ctx := context.Background()
	key := "risk:last:" + ip

	lastStr, err := rdb.Get(ctx, key).Result()
	if err != nil {
		rdb.Set(ctx, key, time.Now().UnixMilli(), 60*time.Second)
		return 0
	}

	lastTime, err := strconv.ParseInt(lastStr, 10, 64)
	if err != nil {
		return 0
	}

	now := time.Now().UnixMilli()
	rdb.Set(ctx, key, now, 60*time.Second)

	if now-lastTime < int64(rule.Threshold) {
		return rule.Score
	}
	return 0
}

// containsIgnoreCase 忽略大小写检查子串
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && containsLower(toLower(s), toLower(substr))
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
