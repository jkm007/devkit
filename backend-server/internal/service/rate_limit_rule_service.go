package service

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"golang.org/x/time/rate"
)

// RateLimitRuleService 限流规则服务
type RateLimitRuleService struct {
	repo *repository.RateLimitRuleRepo
}

func NewRateLimitRuleService() *RateLimitRuleService {
	return &RateLimitRuleService{
		repo: repository.NewRateLimitRuleRepo(),
	}
}

// ==================== CRUD ====================

// GetAll 获取所有规则
func (s *RateLimitRuleService) GetAll() ([]model.RateLimitRule, error) {
	return s.repo.GetAll()
}

// GetByID 根据 ID 获取
func (s *RateLimitRuleService) GetByID(id uint) (*model.RateLimitRule, error) {
	return s.repo.GetByID(id)
}

// Create 创建规则
func (s *RateLimitRuleService) Create(rule *model.RateLimitRule) error {
	err := s.repo.Create(rule)
	if err == nil {
		GetGlobalRuleCache().Reload()
	}
	return err
}

// Update 更新规则
func (s *RateLimitRuleService) Update(rule *model.RateLimitRule) error {
	err := s.repo.Update(rule)
	if err == nil {
		GetGlobalRuleCache().Reload()
	}
	return err
}

// Delete 删除规则
func (s *RateLimitRuleService) Delete(id uint) error {
	err := s.repo.Delete(id)
	if err == nil {
		GetGlobalRuleCache().Reload()
	}
	return err
}

// UpdateEnabled 更新启用状态
func (s *RateLimitRuleService) UpdateEnabled(id uint, enabled bool) error {
	err := s.repo.UpdateEnabled(id, enabled)
	if err == nil {
		GetGlobalRuleCache().Reload()
	}
	return err
}

// ==================== 规则缓存（供中间件使用） ====================

// CachedRule 缓存的限流规则
type CachedRule struct {
	PathPattern    string
	Method         string
	Limiter        *rate.Limiter
	Cooldown       time.Duration // 冷却时间
	BlockDuration  time.Duration // 封禁时长
	MaxViolations  int           // 最大违规次数
	ViolationScore int           // 违规风险分
	Description    string
}

// ipState IP 状态跟踪
type ipState struct {
	limiter      *rate.Limiter
	lastSeen     atomic.Int64
	violations   atomic.Int32  // 违规次数
	blockedUntil atomic.Int64  // 封禁截止时间 (UnixNano)
	cooldownEnd  atomic.Int64  // 冷却截止时间 (UnixNano)
}

// RuleCache 规则缓存
type RuleCache struct {
	rules   atomic.Value // []CachedRule
	mu      sync.Mutex
	limiter sync.Map // key -> *ipState
	started bool
	startMu sync.Mutex
}

var globalRuleCache = &RuleCache{}

func init() {
	globalRuleCache.rules.Store([]CachedRule{})
}

// ensureStarted 确保后台 goroutine 已启动（延迟到数据库就绪后）
func (rc *RuleCache) ensureStarted() {
	rc.startMu.Lock()
	defer rc.startMu.Unlock()
	if rc.started {
		return
	}
	rc.started = true

	// 启动时加载一次
	go rc.Reload()
	// 定时刷新（每 30 秒）
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			rc.Reload()
		}
	}()
	// 定时清理过期 IP（每 5 分钟）
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			rc.limiter.Range(func(key, value interface{}) bool {
				entry := value.(*ipState)
				if now.Sub(time.Unix(0, entry.lastSeen.Load())) > 10*time.Minute {
					rc.limiter.Delete(key)
				}
				return true
			})
		}
	}()
}

// Reload 从数据库重新加载规则
func (rc *RuleCache) Reload() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// 检查数据库是否就绪
	db := database.GetMySQL()
	if db == nil {
		return
	}

	var rules []model.RateLimitRule
	if err := db.Where("enabled = ?", true).Order("priority DESC, id ASC").Find(&rules).Error; err != nil {
		return
	}

	cached := make([]CachedRule, 0, len(rules))
	for _, r := range rules {
		cached = append(cached, CachedRule{
			PathPattern:    r.PathPattern,
			Method:         strings.ToUpper(r.Method),
			Limiter:        rate.NewLimiter(rate.Limit(r.Rate), r.Burst),
			Cooldown:       time.Duration(r.Cooldown) * time.Second,
			BlockDuration:  time.Duration(r.BlockDuration) * time.Second,
			MaxViolations:  r.MaxViolations,
			ViolationScore: r.ViolationScore,
			Description:    r.Description,
		})
	}
	rc.rules.Store(cached)
}

// GetRules 获取当前缓存的规则
func (rc *RuleCache) GetRules() []CachedRule {
	rc.ensureStarted()
	return rc.rules.Load().([]CachedRule)
}

// RateLimitCheckResult 限流检查结果
type RateLimitCheckResult struct {
	Matched        bool   // 是否匹配到规则
	Allowed        bool   // 是否允许通过
	Reason         string // 拒绝原因
	RetryAfter     int    // 建议重试时间（秒）
	ViolationScore int    // 违规风险分（触发限流时累加到风险评分）
}

// MatchAndCheck 匹配规则并检查限流
func (rc *RuleCache) MatchAndCheck(ip, path, method string) RateLimitCheckResult {
	rules := rc.GetRules()
	method = strings.ToUpper(method)
	now := time.Now().UnixNano()

	for _, rule := range rules {
		if matchMethod(rule.Method, method) && matchPath(rule.PathPattern, path) {
			key := ip + ":" + rule.PathPattern + ":" + rule.Method
			val, _ := rc.limiter.LoadOrStore(key, &ipState{
				limiter: rule.Limiter,
			})
			entry := val.(*ipState)
			entry.lastSeen.Store(now)

			// 检查是否被封禁
			if rule.BlockDuration > 0 && rule.MaxViolations > 0 {
				blockedUntil := entry.blockedUntil.Load()
				if blockedUntil > 0 && now < blockedUntil {
					retryAfter := int((blockedUntil - now) / int64(time.Second)) + 1
					return RateLimitCheckResult{
						Matched:        true,
						Allowed:        false,
						Reason:         "请求已被封禁，请稍后再试",
						RetryAfter:     retryAfter,
						ViolationScore: rule.ViolationScore,
					}
				}
				// 封禁已过期，重置违规计数
				if blockedUntil > 0 && now >= blockedUntil {
					entry.blockedUntil.Store(0)
					entry.violations.Store(0)
				}
			}

			// 检查冷却时间
			if rule.Cooldown > 0 {
				cooldownEnd := entry.cooldownEnd.Load()
				if cooldownEnd > 0 && now < cooldownEnd {
					retryAfter := int((cooldownEnd - now) / int64(time.Second)) + 1
					return RateLimitCheckResult{
						Matched:        true,
						Allowed:        false,
						Reason:         "请求过于频繁，冷却中",
						RetryAfter:     retryAfter,
						ViolationScore: rule.ViolationScore,
					}
				}
			}

			// 检查令牌桶
			if !entry.limiter.Allow() {
				// 触发限流，设置冷却时间
				if rule.Cooldown > 0 {
					entry.cooldownEnd.Store(now + int64(rule.Cooldown))
				}

				// 增加违规次数
				if rule.BlockDuration > 0 && rule.MaxViolations > 0 {
					violations := entry.violations.Add(1)
					if violations >= int32(rule.MaxViolations) {
						// 触发封禁，使用 CAS 避免多个 goroutine 重复触发
						if entry.violations.CompareAndSwap(violations, 0) {
								entry.blockedUntil.Store(now + int64(rule.BlockDuration))
								return RateLimitCheckResult{
									Matched:        true,
									Allowed:        false,
									Reason:         "请求已被封禁，请稍后再试",
									RetryAfter:     int(rule.BlockDuration.Seconds()),
									ViolationScore: rule.ViolationScore,
								}
							}
						}
					}

				retryAfter := 1
				if rule.Cooldown > 0 {
					retryAfter = int(rule.Cooldown.Seconds())
				}
				return RateLimitCheckResult{
					Matched:        true,
					Allowed:        false,
					Reason:         "请求过于频繁，请稍后再试",
					RetryAfter:     retryAfter,
					ViolationScore: rule.ViolationScore,
				}
			}

			return RateLimitCheckResult{Matched: true, Allowed: true}
		}
	}
	return RateLimitCheckResult{Matched: false, Allowed: true}
}

// matchMethod 匹配 HTTP 方法
func matchMethod(pattern, method string) bool {
	if pattern == "*" || pattern == "" {
		return true
	}
	return pattern == method
}

// matchPath 匹配路径（支持 * 通配符）
func matchPath(pattern, path string) bool {
	if pattern == "*" {
		return true
	}
	// 精确匹配
	if pattern == path {
		return true
	}
	// 前缀匹配：/auth/*
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix+"/") || path == prefix
	}
	// 后缀匹配：*.js
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}
	return false
}

// GetGlobalRuleCache 获取全局规则缓存（供中间件使用）
func GetGlobalRuleCache() *RuleCache {
	return globalRuleCache
}
