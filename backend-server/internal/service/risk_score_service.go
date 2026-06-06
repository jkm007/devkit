package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"backend-server/pkg/database"
)

// RiskScoreItem 风险评分项
type RiskScoreItem struct {
	IP          string    `json:"ip"`
	Score       int       `json:"score"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ExpireAt    time.Time `json:"expireAt"`
	TriggeredAt time.Time `json:"triggeredAt,omitempty"` // 触发验证码的时间
	BlockedAt   time.Time `json:"blockedAt,omitempty"`   // 拦截的时间
}

// RiskScoreService 风险评分服务
type RiskScoreService struct{}

// NewRiskScoreService 创建风险评分服务
func NewRiskScoreService() *RiskScoreService {
	return &RiskScoreService{}
}

// GetRiskScores 获取所有风险评分（按分数排序）
func (s *RiskScoreService) GetRiskScores(limit int) ([]RiskScoreItem, error) {
	rdb := database.GetRedis()
	ctx := context.Background()

	// 扫描所有 risk:score:* 键
	var cursor uint64
	var keys []string
	var scores []RiskScoreItem

	for {
		// SCAN 命令遍历
		result, nextCursor, err := rdb.Scan(ctx, cursor, "risk:score:*", 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, result...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	// 获取每个键的分数和 TTL
	for _, key := range keys {
		// 提取 IP
		ip := strings.TrimPrefix(key, "risk:score:")

		// 获取分数
		score, err := rdb.Get(ctx, key).Int()
		if err != nil {
			continue // 键可能已过期
		}

		// 获取 TTL
		ttl, err := rdb.TTL(ctx, key).Result()
		if err != nil || ttl < 0 {
			ttl = 0
		}

		// 解析更新时间（从 TTL 反推）
		cfg := GetRiskConfig()
		totalTTL := time.Duration(cfg.DecayMinutes) * time.Minute
		updatedAt := time.Now().Add(-totalTTL).Add(ttl)

		// 计算过期时间
		expireAt := time.Now().Add(ttl)

		scores = append(scores, RiskScoreItem{
			IP:        ip,
			Score:     score,
			UpdatedAt: updatedAt,
			ExpireAt:  expireAt,
		})
	}

	// 按分数降序排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// 限制返回数量
	if limit > 0 && len(scores) > limit {
		scores = scores[:limit]
	}

	return scores, nil
}

// GetRiskScoreByIP 获取指定 IP 的风险评分
func (s *RiskScoreService) GetRiskScoreByIP(ip string) (*RiskScoreItem, error) {
	rdb := database.GetRedis()
	ctx := context.Background()
	key := "risk:score:" + ip

	score, err := rdb.Get(ctx, key).Int()
	if err != nil {
		// 键不存在或已过期
		return &RiskScoreItem{IP: ip, Score: 0}, nil
	}

	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil || ttl < 0 {
		ttl = 0
	}

	cfg := GetRiskConfig()
	totalTTL := time.Duration(cfg.DecayMinutes) * time.Minute
	updatedAt := time.Now().Add(-totalTTL).Add(ttl)

	return &RiskScoreItem{
		IP:        ip,
		Score:     score,
		UpdatedAt: updatedAt,
		ExpireAt:  time.Now().Add(ttl),
	}, nil
}

// ClearRiskScore 清除指定 IP 的风险评分
func (s *RiskScoreService) ClearRiskScore(ip string) error {
	rdb := database.GetRedis()
	ctx := context.Background()

	// 删除所有相关键
	rdb.Del(ctx, "risk:score:"+ip)
	rdb.Del(ctx, "risk:freq:"+ip)
	rdb.Del(ctx, "risk:last:"+ip)

	return nil
}

// GetRiskScoreStats 获取风险评分统计
func (s *RiskScoreService) GetRiskScoreStats() (map[string]interface{}, error) {
	// 扫描所有键
	scores, err := s.GetRiskScores(0)
	if err != nil {
		return nil, err
	}

	// 统计
	totalCount := len(scores)
	triggerCount := 0  // 达到触发阈值
	blockCount := 0    // 达到拦截阈值
	highRiskCount := 0 // 高风险 (>70)

	cfg := GetRiskConfig()

	for _, item := range scores {
		if item.Score >= cfg.BlockScore {
			blockCount++
		}
		if item.Score >= cfg.TriggerScore {
			triggerCount++
		}
		if item.Score >= 70 {
			highRiskCount++
		}
	}

	return map[string]interface{}{
		"totalCount":     totalCount,
		"triggerCount":   triggerCount,
		"blockCount":     blockCount,
		"highRiskCount":  highRiskCount,
		"triggerScore":   cfg.TriggerScore,
		"blockScore":     cfg.BlockScore,
		"enabled":        cfg.Enabled,
	}, nil
}