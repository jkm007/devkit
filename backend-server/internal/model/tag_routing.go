package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// TagCondition 标签条件
type TagCondition struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TagConditions 条件列表
type TagConditions struct {
	Tags []TagCondition `json:"tags"`
}

// Value 实现 driver.Valuer 接口
func (tc TagConditions) Value() (driver.Value, error) {
	return json.Marshal(tc)
}

// Scan 实现 sql.Scanner 接口
func (tc *TagConditions) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan TagConditions: %v", value)
	}
	return json.Unmarshal(bytes, tc)
}

// TagRouting 标签路由规则模型
type TagRouting struct {
	ID          int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	RuleName    string       `json:"ruleName" gorm:"size:100;not null"`
	Description string       `json:"description" gorm:"size:200"`
	Priority    int          `json:"priority" gorm:"default:0;index:idx_priority"`
	MatchType   string       `json:"matchType" gorm:"size:10;not null;default:all"`
	Conditions  TagConditions `json:"conditions" gorm:"type:json;not null"`
	Driver      string       `json:"driver" gorm:"size:20;not null"`
	Bucket      string       `json:"bucket" gorm:"size:100"`
	PathPrefix  string       `json:"pathPrefix" gorm:"size:200"`
	ExtraConfig string       `json:"extraConfig" gorm:"type:text"`
	IsDefault   bool         `json:"isDefault" gorm:"default:false;index:idx_default"`
	Status      int8         `json:"status" gorm:"default:1;index:idx_status"`
	CreatedAt   time.Time    `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (TagRouting) TableName() string {
	return "sys_tag_routing"
}
