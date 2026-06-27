package model

import (
	"time"
)

// ClassInvitation 班级邀请码
type ClassInvitation struct {
	ID        uint       `gorm:"primaryKey;autoIncrement;comment:邀请ID" json:"id"`
	ClassID   uint       `gorm:"index;not null;comment:班级ID" json:"classId"`
	Code      string     `gorm:"type:varchar(20);uniqueIndex;not null;comment:邀请码" json:"code"`
	ExpireAt  *time.Time `gorm:"comment:过期时间，空表示不过期" json:"expireAt"`
	MaxUses   int        `gorm:"default:0;comment:最大使用次数 0表示无限制" json:"maxUses"`
	UsedCount int        `gorm:"default:0;comment:已使用次数" json:"usedCount"`
	Status    int        `gorm:"default:1;comment:状态 1有效 0失效" json:"status"`
	CreatedBy uint       `gorm:"not null;comment:创建人ID" json:"createdBy"`
	CreatedAt time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"-"`
}

// TableName 表名
func (ClassInvitation) TableName() string {
	return "sys_class_invitations"
}
