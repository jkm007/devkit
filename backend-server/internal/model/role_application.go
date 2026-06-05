package model

import (
	"time"

	"gorm.io/gorm"
)

// RoleApplication 角色申请模型
type RoleApplication struct {
	ID         uint           `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID     uint           `gorm:"not null;comment:申请人ID" json:"userId"`
	RoleID     uint           `gorm:"not null;comment:申请角色ID" json:"roleId"`
	Reason     string         `gorm:"type:varchar(500);default:;comment:申请理由" json:"reason"`
	Status     int            `gorm:"type:tinyint;default:0;comment:状态 0待审 1通过 2驳回" json:"status"`
	ReviewNote string         `gorm:"type:varchar(500);default:;comment:审核备注" json:"reviewNote"`
	ReviewedBy *uint          `gorm:"comment:审核人ID" json:"reviewedBy"`
	ReviewedAt *time.Time     `gorm:"comment:审核时间" json:"reviewedAt"`
	CreatedAt  time.Time      `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt  time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 表名
func (RoleApplication) TableName() string {
	return "sys_role_applications"
}
