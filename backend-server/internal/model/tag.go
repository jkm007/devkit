package model

import "time"

// Tag 标签定义模型
type Tag struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TagKey      string    `json:"tagKey" gorm:"size:50;not null;uniqueIndex:uk_key_value"`
	TagValue    string    `json:"tagValue" gorm:"size:50;not null;uniqueIndex:uk_key_value"`
	TagName     string    `json:"tagName" gorm:"size:100;not null"`
	Icon        string    `json:"icon" gorm:"size:50"`
	Color       string    `json:"color" gorm:"size:20"`
	Description string    `json:"description" gorm:"size:200"`
	IsSystem    bool      `json:"isSystem" gorm:"default:false"`
	SortOrder   int       `json:"sortOrder" gorm:"default:0"`
	Status      int8      `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Tag) TableName() string {
	return "sys_tag"
}
