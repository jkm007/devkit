package model

import (
	"time"

	"gorm.io/gorm"
)

// Class 班级
// 支持创建者通过邀请码邀请其他用户加入
type Class struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;comment:班级ID" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null;comment:班级名称" json:"name"`
	Code        string         `gorm:"type:varchar(20);uniqueIndex;not null;comment:邀请码" json:"code"`
	Description string         `gorm:"type:varchar(500);comment:班级描述" json:"description"`
	Status      int            `gorm:"default:1;comment:状态 1启用 0禁用" json:"status"`
	CreatedBy   uint           `gorm:"index;not null;comment:创建人ID" json:"createdBy"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 表名
func (Class) TableName() string {
	return "sys_classes"
}
