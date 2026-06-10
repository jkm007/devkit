package middleware

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"strconv"
	"time"

	"backend-server/pkg/captcha"
	"backend-server/pkg/database"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 所有可用的验证码类型（随机选择）
var captchaTypes = []string{"numeric", "slider", "puzzle", "rotation", "point"}

// randomCaptchaType 随机返回一种验证码类型
func randomCaptchaType() string {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(captchaTypes))))
	if err != nil {
		return captchaTypes[0]
	}
	return captchaTypes[n.Int64()]
}

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
			log.Printf("[RiskScore] ip=%s path=%s | 白名单放行", ip, path)
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
		requestScore, details := calculateRiskScore(ip, path, headers, cfg)

		// 累加到 Redis 中的总分
		totalScore := accumulateRiskScore(ip, requestScore, cfg)

		log.Printf("[RiskScore] ip=%s path=%s | 本次=%d(%s) 累计=%d 阈值=%d",
			ip, path, requestScore, details, totalScore, cfg.TriggerScore)

		// 低风险：直接放行
		if totalScore < cfg.TriggerScore {
			c.Next()
			return
		}

		// 高风险：检查请求头中是否携带验证码
		captchaID := c.GetHeader("X-Captcha-Id")
		captchaCode := c.GetHeader("X-Captcha-Code")

		// 没有验证码头的请求：返回 403001 要求验证码（随机类型）
		if captchaID == "" || captchaCode == "" {
			captchaType := randomCaptchaType()
			response.CaptchaRequiredWithType(c, "当前操作需要验证码验证", captchaType)
			c.Abort()
			return
		}

		// 携带验证码头：始终允许验证（无论分数多高，给用户验证的机会）
		// 从 Header 读取 startTime（验证码渲染时刻）
		var startTime int64
		if startTimeStr := c.GetHeader("X-Captcha-Start-Time"); startTimeStr != "" {
			startTime, _ = strconv.ParseInt(startTimeStr, 10, 64)
		}

		// 解析点选验证码的坐标数据
		var points []captcha.Point
		if len(captchaCode) > 0 && captchaCode[0] == '[' {
			_ = json.Unmarshal([]byte(captchaCode), &points)
		}

		// 验证验证码
		valid, _ := captcha.Verify(captchaID, captchaCode, startTime, points)
		if !valid {
			// 验证失败，返回 403001（随机新类型）让前端重新获取验证码
			captchaType := randomCaptchaType()
			response.CaptchaRequiredWithType(c, "验证码错误或已过期", captchaType)
			c.Abort()
			return
		}

		// 验证成功，写入白名单（10分钟有效）
		rdb.Set(ctx, whitelistKey, "1", 10*time.Minute)

		// 验证通过，风险分减半而非清零
		halveRiskScore(ip)
		log.Printf("[RiskScore] ip=%s | 验证码通过，写入白名单(10min)，风险分减半", ip)
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
// 返回风险分和触发详情
func calculateRiskScore(ip, path string, headers map[string]string, cfg *RiskConfigGetter) (int, string) {
	score := 0
	details := make([]string, 0)

	for _, rule := range cfg.Rules {
		if !rule.Enabled {
			continue
		}

		switch rule.Key {
		case "frequency":
			s := evalFrequencyRule(rule, ip)
			if s > 0 {
				score += s
				details = append(details, "频率+"+strconv.Itoa(s))
			}
		case "no_referer":
			if headers["Referer"] == "" {
				score += rule.Score
				details = append(details, "无Referer+"+strconv.Itoa(rule.Score))
			}
		case "no_lang":
			if headers["Accept-Language"] == "" {
				score += rule.Score
				details = append(details, "无语言+"+strconv.Itoa(rule.Score))
			}
		case "ua":
			s := evalUARule(rule, headers["User-Agent"])
			if s > 0 {
				score += s
				details = append(details, "UA异常+"+strconv.Itoa(s))
			}
		case "interval":
			s := evalIntervalRule(rule, ip)
			if s > 0 {
				score += s
				details = append(details, "窗口频率+"+strconv.Itoa(s))
			}
		}
	}

	detail := "无"
	if len(details) > 0 {
		detail = joinStrings(details, ", ")
	}
	return score, detail
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += sep + p
	}
	return result
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

// evalIntervalRule 滑动窗口频率检测
// 使用 Redis 有序集合记录请求时间戳，统计窗口内的请求次数
// Threshold = 窗口内最大请求数，Keywords[0] = 窗口秒数（默认10）
func evalIntervalRule(rule RiskRuleItem, ip string) int {
	rdb := database.GetRedis()
	ctx := context.Background()
	key := "risk:window:" + ip

	// 窗口大小（秒），从 Keywords[0] 读取，默认 10 秒
	windowSeconds := 10
	if len(rule.Keywords) > 0 {
		if v, err := strconv.Atoi(rule.Keywords[0]); err == nil && v > 0 {
			windowSeconds = v
		}
	}

	now := time.Now().UnixMilli()
	windowStart := now - int64(windowSeconds*1000)

	// 原子操作：清除过期记录 + 添加当前请求 + 计数
	pipe := rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: strconv.FormatInt(now, 10) + ":" + strconv.Itoa(int(now%1000))})
	countCmd := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, time.Duration(windowSeconds+5)*time.Second)
	_, _ = pipe.Exec(ctx)

	count := int(countCmd.Val())

	// 阈值：窗口内最大请求数（Threshold 字段复用为 maxCount）
	maxCount := rule.Threshold
	if maxCount <= 0 {
		maxCount = 50
	}

	if count > maxCount {
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
