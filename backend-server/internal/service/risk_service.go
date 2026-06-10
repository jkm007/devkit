package service

import (
	"strconv"
	"strings"
	"sync"

	"backend-server/pkg/database"
)

// RiskConfig 风险评分配置
type RiskConfig struct {
	Enabled      bool
	TriggerScore int
	BlockScore   int
	DecayMinutes int
	DecayRate    float64
	Paths        []string   // 受保护路径前缀
	Rules        []RiskRule
}

// RiskRule 风险评分规则
type RiskRule struct {
	Key       string   // 规则标识: frequency, no_referer, no_lang, ua, interval
	Enabled   bool
	Score     int      // 触发时加分
	Threshold int      // 阈值（频率上限、最小间隔毫秒等）
	Keywords  []string // UA 关键词
}

// 配置缓存
var riskConfig *RiskConfig
var riskConfigMutex sync.RWMutex

// GetRiskConfig 获取风险评分配置（带缓存）
func GetRiskConfig() *RiskConfig {
	riskConfigMutex.RLock()
	if riskConfig != nil {
		riskConfigMutex.RUnlock()
		return riskConfig
	}
	riskConfigMutex.RUnlock()

	riskConfigMutex.Lock()
	defer riskConfigMutex.Unlock()

	if riskConfig != nil {
		return riskConfig
	}

	riskConfig = loadRiskConfigFromDB()
	return riskConfig
}

// RefreshRiskConfig 刷新风险评分配置缓存
func RefreshRiskConfig() {
	riskConfigMutex.Lock()
	riskConfig = loadRiskConfigFromDB()
	riskConfigMutex.Unlock()
}

// loadRiskConfigFromDB 从数据库加载风险评分配置
func loadRiskConfigFromDB() *RiskConfig {
	db := database.GetMySQL()

	rows, err := db.Raw("SELECT `key`, value FROM sys_system_settings WHERE group_key = 'risk_score' AND deleted_at IS NULL").Rows()
	if err != nil {
		return getDefaultRiskConfig()
	}
	defer rows.Close()

	// 临时存储 key-value
	kv := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		value = strings.Trim(value, "\"")
		kv[key] = value
	}

	if len(kv) == 0 {
		return getDefaultRiskConfig()
	}

	cfg := &RiskConfig{
		Enabled:      parseBool(kv["risk_enabled"], true),
		TriggerScore: parseInt(kv["risk_trigger_score"], 50),
		BlockScore:   parseInt(kv["risk_block_score"], 80),
		DecayMinutes: parseInt(kv["risk_decay_minutes"], 30),
		DecayRate:    parseFloat(kv["risk_decay_rate"], 0.5),
		Paths:        parseStringList(kv["risk_protected_paths"]),
	}

	// 频率检测
	cfg.Rules = append(cfg.Rules, RiskRule{
		Key:       "frequency",
		Enabled:   parseBool(kv["rule_frequency_enabled"], true),
		Score:     parseInt(kv["rule_frequency_score"], 30),
		Threshold: parseInt(kv["rule_frequency_threshold"], 30),
	})

	// 无 Referer
	cfg.Rules = append(cfg.Rules, RiskRule{
		Key:     "no_referer",
		Enabled: parseBool(kv["rule_no_referer_enabled"], true),
		Score:   parseInt(kv["rule_no_referer_score"], 20),
	})

	// 无 Accept-Language
	cfg.Rules = append(cfg.Rules, RiskRule{
		Key:     "no_lang",
		Enabled: parseBool(kv["rule_no_lang_enabled"], true),
		Score:   parseInt(kv["rule_no_lang_score"], 15),
	})

	// UA 异常
	cfg.Rules = append(cfg.Rules, RiskRule{
		Key:       "ua",
		Enabled:   parseBool(kv["rule_ua_enabled"], true),
		Score:     parseInt(kv["rule_ua_score"], 25),
		Keywords:  parseStringList(kv["rule_ua_keywords"]),
	})

	// 请求间隔
	cfg.Rules = append(cfg.Rules, RiskRule{
		Key:       "interval",
		Enabled:   parseBool(kv["rule_interval_enabled"], true),
		Score:     parseInt(kv["rule_interval_score"], 20),
		Threshold: parseInt(kv["rule_interval_min_ms"], 500),
	})

	return cfg
}

// IsProtectedPath 检查路径是否需要风险评估
func (c *RiskConfig) IsProtectedPath(path string) bool {
	for _, prefix := range c.Paths {
		if strings.HasPrefix(path, prefix) {
			if len(path) == len(prefix) || path[len(prefix)] == '/' {
				return true
			}
		}
	}
	return false
}

// 辅助函数
func parseBool(s string, defaultVal bool) bool {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return defaultVal
	}
	return v
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

func parseFloat(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

func parseStringList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// getDefaultRiskConfig 返回默认风险评分配置
func getDefaultRiskConfig() *RiskConfig {
	return &RiskConfig{
		Enabled:      true,
		TriggerScore: 50,
		BlockScore:   80,
		DecayMinutes: 30,
		DecayRate:    0.5,
		Paths:        []string{"/system/user", "/system/role", "/system/group", "/system/menu", "/system/settings", "/system/storage-buckets", "/system/storage-configs", "/system/scheduled-tasks", "/files", "/files/recycle", "/shares"},
		Rules: []RiskRule{
			{Key: "frequency", Enabled: true, Score: 30, Threshold: 30},
			{Key: "no_referer", Enabled: true, Score: 20},
			{Key: "no_lang", Enabled: true, Score: 15},
			{Key: "ua", Enabled: true, Score: 25, Keywords: []string{"curl", "python", "java", "go-http", "postman", "scrapy"}},
			{Key: "interval", Enabled: true, Score: 20, Threshold: 500},
		},
	}
}
