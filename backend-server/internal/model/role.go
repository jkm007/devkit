package model

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;comment:角色ID" json:"id"`
	Name        string         `gorm:"type:varchar(50);not null;comment:角色名称" json:"name"`
	Status      int            `gorm:"type:tinyint;default:1;comment:状态 1:启用 0:禁用" json:"status"`
	Permissions string         `gorm:"type:text;comment:权限列表(JSON数组)" json:"permissions"`
	Remark      string         `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedAt   time.Time      `gorm:"comment:创建时间" json:"createTime"`
	UpdatedAt   time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 表名
func (Role) TableName() string {
	return "sys_roles"
}

// GetPermissions 获取权限列表
func (r *Role) GetPermissions() []string {
	// 从 JSON 字符串解析权限列表
	// 实际实现需要 json.Unmarshal
	return []string{}
}
