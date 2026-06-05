package model

import (
	"time"

	"gorm.io/gorm"
)

// Group 分组模型
type Group struct {
	ID        uint           `gorm:"primaryKey;autoIncrement;comment:分组ID" json:"id"`
	PID       uint           `gorm:"column:pid;default:0;comment:父分组ID" json:"pid"`
	Name      string         `gorm:"type:varchar(100);not null;comment:分组名称" json:"name"`
	Status    int            `gorm:"type:tinyint;default:1;comment:状态 1:启用 0:禁用" json:"status"`
	Remark    string         `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"createTime"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 表名
func (Group) TableName() string {
	return "sys_groups"
}

// GroupTree 分组树结构
type GroupTree struct {
	Group
	Children []*GroupTree `json:"children,omitempty"`
}

// GroupRole 分组角色关联
type GroupRole struct {
	ID      uint `gorm:"primaryKey;autoIncrement;comment:ID" json:"id"`
	GroupID uint `gorm:"not null;comment:分组ID" json:"groupId"`
	RoleID  uint `gorm:"not null;comment:角色ID" json:"roleId"`
}

// TableName 表名
func (GroupRole) TableName() string {
	return "sys_group_roles"
}
