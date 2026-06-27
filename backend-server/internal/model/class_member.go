package model

import (
	"time"
)

// ClassMemberRole 班级成员角色
type ClassMemberRole string

const (
	// ClassRoleStudent 同学：仅可查看班级题目
	ClassRoleStudent ClassMemberRole = "student"
	// ClassRoleMonitor 班级管理员：可管理同学、审批加入
	ClassRoleMonitor ClassMemberRole = "monitor"
	// ClassRoleTeacher 班主任/老师：班级全权限
	ClassRoleTeacher ClassMemberRole = "teacher"
)

// ClassMember 班级成员
type ClassMember struct {
	ID        uint            `gorm:"primaryKey;autoIncrement;comment:成员ID" json:"id"`
	ClassID   uint            `gorm:"index:idx_class_member_class_user,unique;not null;comment:班级ID" json:"classId"`
	UserID    uint            `gorm:"index:idx_class_member_class_user,unique;not null;comment:用户ID" json:"userId"`
	Role      ClassMemberRole `gorm:"type:varchar(20);not null;comment:角色 student/monitor/teacher" json:"role"`
	Status    int             `gorm:"default:1;comment:状态 1正常 0待审批" json:"status"`
	JoinedAt  time.Time       `gorm:"autoCreateTime;comment:加入时间" json:"joinedAt"`
	CreatedAt time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"-"`
}

// TableName 表名
func (ClassMember) TableName() string {
	return "sys_class_members"
}
