package model

import (
	"time"
)

// UserCategoryBinding 用户分类绑定
// Note: category_id 字段存储的是 Subject ID（L3 科目ID），
// 而非 Category ID（L4 章节分类ID）。这是移动端分类系统的
// 设计决策：用户绑定到科目级别，科目是最合适的刷题粒度。
type UserCategoryBinding struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"index;not null;comment:用户ID" json:"userId"`
	CategoryID uint      `gorm:"index;not null;comment:科目ID(Subject)" json:"categoryId"`
	IsPrimary  bool      `gorm:"default:false;comment:是否主分类" json:"isPrimary"`
	BoundAt    time.Time `gorm:"autoCreateTime;comment:绑定时间" json:"boundAt"`
}

// TableName 表名
func (UserCategoryBinding) TableName() string {
	return "user_category_bindings"
}
