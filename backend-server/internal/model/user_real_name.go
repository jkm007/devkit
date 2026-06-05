package model

import (
	"time"

	"gorm.io/gorm"
)

// UserRealName 实名认证模型
type UserRealName struct {
	ID           uint           `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID       uint           `gorm:"uniqueIndex;not null;comment:用户ID" json:"userId"`
	RealName     string         `gorm:"type:varchar(50);not null;comment:真实姓名" json:"realName"`
	IDCard       string         `gorm:"type:varchar(255);not null;comment:身份证号(AES加密)" json:"-"`
	IDCardHash   string         `gorm:"type:varchar(64);default:;comment:身份证号哈希" json:"-"`
	Status       int            `gorm:"type:tinyint;default:0;comment:状态 0待审 1已认证 2认证失败" json:"status"`
	RejectReason string         `gorm:"type:varchar(200);default:;comment:拒绝原因" json:"rejectReason"`
	ReviewedBy   *uint          `gorm:"comment:审核人ID" json:"reviewedBy"`
	ReviewedAt   *time.Time     `gorm:"comment:审核时间" json:"reviewedAt"`
	SubmittedAt  *time.Time     `gorm:"comment:提交时间" json:"submittedAt"`
	CreatedAt    time.Time      `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 表名
func (UserRealName) TableName() string {
	return "sys_user_real_names"
}
