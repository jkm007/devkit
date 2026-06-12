package model

import (
	"time"

	"gorm.io/gorm"
)

// QuestionShare 题目分享模型
type QuestionShare struct {
	ID              uint           `gorm:"primaryKey;autoIncrement;comment:分享ID" json:"id"`
	QuestionID      uint           `gorm:"not null;index;comment:题目ID" json:"questionId"`
	QuestionVersionID uint         `gorm:"not null;comment:题目版本ID" json:"questionVersionId"`
	ShareCode       string         `gorm:"type:varchar(64);not null;uniqueIndex;comment:分享码" json:"shareCode"`
	ShareType       string         `gorm:"type:varchar(20);not null;comment:分享类型" json:"shareType"`
	TargetID        uint           `gorm:"default:0;comment:目标ID" json:"targetId"`
	ExpireAt        *time.Time     `gorm:"comment:过期时间" json:"expireAt"`
	MaxAccess       int            `gorm:"default:0;comment:最大访问次数 0=不限" json:"maxAccess"`
	AccessCount     int            `gorm:"default:0;comment:已访问次数" json:"accessCount"`
	Status          int            `gorm:"type:tinyint;default:1;comment:状态 1有效 2过期 3禁用" json:"status"`
	CreatedBy       uint           `gorm:"not null;comment:创建人ID" json:"createdBy"`
	CreatedAt       time.Time      `gorm:"comment:创建时间" json:"createTime"`
	AccessedAt      *time.Time     `gorm:"comment:最后访问时间" json:"accessedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (QuestionShare) TableName() string {
	return "qb_question_shares"
}
