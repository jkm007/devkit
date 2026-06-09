package model

import "time"

// RateLimitRule 限流规则
type RateLimitRule struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PathPattern   string    `gorm:"size:255;not null;index" json:"pathPattern"`   // 路径模式，支持 * 通配符
	Method        string    `gorm:"size:10;default:'*'" json:"method"`            // HTTP 方法：GET/POST/PUT/DELETE/* (所有)
	Rate          float64   `gorm:"not null;default:10" json:"rate"`              // 每秒请求数
	Burst         int       `gorm:"not null;default:20" json:"burst"`             // 突发容量
	Cooldown      int       `gorm:"not null;default:0" json:"cooldown"`           // 冷却时间（秒），触发限流后需等待多久恢复
	BlockDuration int       `gorm:"not null;default:0" json:"blockDuration"`      // 封禁时长（秒），超过触发次数后封禁 IP
	MaxViolations int       `gorm:"not null;default:0" json:"maxViolations"`      // 最大违规次数，超过后触发封禁（0=不封禁）
	ViolationScore int      `gorm:"not null;default:0" json:"violationScore"`     // 违规风险分，触发限流时累加到风险评分系统（0=不累加）
	Description   string    `gorm:"size:500" json:"description"`                  // 规则描述
	Enabled       bool      `gorm:"default:true" json:"enabled"`                  // 是否启用
	Priority      int       `gorm:"default:0;index" json:"priority"`              // 优先级，数值越大越先匹配
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (RateLimitRule) TableName() string {
	return "sys_rate_limit_rules"
}
