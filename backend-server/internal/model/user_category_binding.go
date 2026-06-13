package model

import (
	"time"
)

// UserCategoryBinding 用户分类绑定
type UserCategoryBinding struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"index;not null;comment:用户ID" json:"userId"`
	CategoryID uint      `gorm:"index;not null;comment:分类ID" json:"categoryId"`
	IsPrimary  bool      `gorm:"default:false;comment:是否主分类" json:"isPrimary"`
	BoundAt    time.Time `gorm:"autoCreateTime;comment:绑定时间" json:"boundAt"`
}

// TableName 表名
func (UserCategoryBinding) TableName() string {
	return "user_category_bindings"
}
