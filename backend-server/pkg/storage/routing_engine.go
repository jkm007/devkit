package storage

import (
	"backend-server/internal/model"
	"fmt"
	"sort"
)

// RoutingTag 路由标签
type RoutingTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RoutingResult 路由结果
type RoutingResult struct {
	Driver     string `json:"driver"`
	Bucket     string `json:"bucket"`
	PathPrefix string `json:"pathPrefix"`
	RuleID     int64  `json:"ruleId"`
	RuleName   string `json:"ruleName"`
}

// RoutingEngine 路由引擎
type RoutingEngine struct {
	rules []model.TagRouting
}

// NewRoutingEngine 创建路由引擎
func NewRoutingEngine(rules []model.TagRouting) *RoutingEngine {
	// 按优先级排序（降序）
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})
	return &RoutingEngine{rules: rules}
}

// Match 匹配路由规则
func (e *RoutingEngine) Match(tags []RoutingTag) *RoutingResult {
	// 先匹配非默认规则
	for _, rule := range e.rules {
		if !rule.IsDefault && rule.Status == 1 && e.matchRule(rule, tags) {
			return &RoutingResult{
				Driver:     rule.Driver,
				Bucket:     rule.Bucket,
				PathPrefix: rule.PathPrefix,
				RuleID:     rule.ID,
				RuleName:   rule.RuleName,
			}
		}
	}

	// 返回默认规则
	for _, rule := range e.rules {
		if rule.IsDefault && rule.Status == 1 {
			return &RoutingResult{
				Driver:     rule.Driver,
				Bucket:     rule.Bucket,
				PathPrefix: rule.PathPrefix,
				RuleID:     rule.ID,
				RuleName:   rule.RuleName,
			}
		}
	}

	return nil
}

// MatchWithTags 根据已有标签匹配（接受 map 格式）
func (e *RoutingEngine) MatchWithTags(tagMap map[string]string) *RoutingResult {
	tags := make([]RoutingTag, 0, len(tagMap))
	for k, v := range tagMap {
		tags = append(tags, RoutingTag{Key: k, Value: v})
	}
	return e.Match(tags)
}

// matchRule 匹配单个规则
func (e *RoutingEngine) matchRule(rule model.TagRouting, tags []RoutingTag) bool {
	conditions := rule.Conditions.Tags
	if len(conditions) == 0 {
		return true
	}

	// 构建标签 map
	tagMap := make(map[string]string)
	for _, t := range tags {
		tagMap[t.Key] = t.Value
	}

	switch rule.MatchType {
	case "all":
		// 所有条件都满足
		for _, cond := range conditions {
			if tagMap[cond.Key] != cond.Value {
				return false
			}
		}
		return true

	case "any":
		// 任一条件满足
		for _, cond := range conditions {
			if tagMap[cond.Key] == cond.Value {
				return true
			}
		}
		return false

	case "exact":
		// 精确匹配（标签完全一致）
		if len(tags) != len(conditions) {
			return false
		}
		for _, cond := range conditions {
			if tagMap[cond.Key] != cond.Value {
				return false
			}
		}
		return true

	default:
		return false
	}
}

// TestRule 测试规则匹配
func (e *RoutingEngine) TestRule(ruleID int64, tags []RoutingTag) (bool, error) {
	for _, rule := range e.rules {
		if rule.ID == ruleID {
			return e.matchRule(rule, tags), nil
		}
	}
	return false, fmt.Errorf("rule not found: %d", ruleID)
}

// GetRules 获取所有规则
func (e *RoutingEngine) GetRules() []model.TagRouting {
	return e.rules
}

// GetRuleByID 根据ID获取规则
func (e *RoutingEngine) GetRuleByID(id int64) *model.TagRouting {
	for _, rule := range e.rules {
		if rule.ID == id {
			return &rule
		}
	}
	return nil
}
