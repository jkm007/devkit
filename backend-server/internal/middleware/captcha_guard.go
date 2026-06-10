package middleware

import (
	"context"
	"strconv"
	"time"

	"backend-server/pkg/captcha"
	"backend-server/pkg/database"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const riskScoreKeyPrefix = "risk:score:"

// CaptchaGuard 风险评分中间件
// 对受保护路径进行风险评估，高风险请求要求验证码
func CaptchaGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if getRiskCfg == nil {
			c.Next()
			return
		}
		cfg := getRiskCfg()
		if cfg == nil {
			c.Next()
			return
		}

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

		// 检查验证码白名单（10分钟内已验证过则直接放行，不累加风险分）
		rdb := database.GetRedis()
		ctx := context.Background()
		whitelistKey := "captcha:whitelist:" + ip
		if exists, _ := rdb.Exists(ctx, whitelistKey).Result(); exists > 0 {
			c.Next()
			return
		}

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

		// 从 Header 读取 startTime（验证码渲染时刻）
		var startTime int64
		if startTimeStr := c.GetHeader("X-Captcha-Start-Time"); startTimeStr != "" {
			startTime, _ = strconv.ParseInt(startTimeStr, 10, 64)
		}

		// 验证验证码
		valid, _ := captcha.Verify(captchaID, captchaCode, startTime, nil)
		if !valid {
			// 验证失败，累积风险分（使用配置中的最大单条规则分作为惩罚分）
			punishScore := 0
			for _, rule := range cfg.Rules {
				if rule.Enabled && rule.Score > punishScore {
					punishScore = rule.Score
				}
			}
			if punishScore > 0 {
				accumulateRiskScore(ip, punishScore, cfg)
			}
			response.CaptchaRequired(c, "验证码错误或已过期")
			c.Abort()
			return
		}

		// 验证成功，写入白名单（10分钟有效）
		rdb.Set(ctx, whitelistKey, "1", 10*time.Minute)

		// 验证通过，风险分减半而非清零
		halveRiskScore(ip)
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

// accumulateRiskScore 累加风险分到 Redis（原子操作）
func accumulateRiskScore(ip string, addScore int, cfg *RiskConfigGetter) int {
	rdb := database.GetRedis()
	ctx := context.Background()
	key := riskScoreKeyPrefix + ip

	// 使用 INCRBY 原子累加，避免竞态条件
	total, err := rdb.IncrBy(ctx, key, int64(addScore)).Result()
	if err != nil {
		// 降级：直接返回 addScore
		return addScore
	}

	// 重置 TTL
	rdb.Expire(ctx, key, time.Duration(cfg.DecayMinutes)*time.Minute)

	return int(total)
}

// halveScript Lua 脚本：原子地获取值、TTL，减半后写回（Bug 2/3 修复）
var halveScript = redis.NewScript(`
local val = tonumber(redis.call('GET', KEYS[1]))
if not val or val <= 0 then
    return 0
end
local ttl = redis.call('TTL', KEYS[1])
if ttl <= 0 then
    ttl = 1800
end
local halved = math.floor(val / 2)
if halved <= 0 then
    redis.call('DEL', KEYS[1])
else
    redis.call('SET', KEYS[1], halved, 'EX', ttl)
end
return halved
`)

// halveRiskScore 风险分减半（验证通过时适度降低，而非直接清零）
func halveRiskScore(ip string) {
	rdb := database.GetRedis()
	ctx := context.Background()
	key := riskScoreKeyPrefix + ip

	halveScript.Run(ctx, rdb, []string{key})
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
