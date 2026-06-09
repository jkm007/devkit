package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID               uint           `gorm:"primaryKey;autoIncrement;comment:用户ID" json:"id"`
	Name             string         `gorm:"type:varchar(50);not null;comment:用户名" json:"name"`
	Nickname         string         `gorm:"type:varchar(50);default:;comment:昵称" json:"nickname"`
	Email            string         `gorm:"type:varchar(100);default:;comment:邮箱" json:"email"`
	Phone            string         `gorm:"type:varchar(20);default:;comment:手机号" json:"phone"`
	Avatar           string         `gorm:"type:varchar(1000);default:;comment:头像URL" json:"avatar"`
	Gender           int            `gorm:"type:tinyint;default:0;comment:性别 0未知 1男 2女" json:"gender"`
	Birthday         *time.Time     `gorm:"comment:生日" json:"birthday"`
	Bio              string         `gorm:"type:varchar(500);default:;comment:个人简介" json:"bio"`
	Password         string         `gorm:"type:varchar(255);not null;comment:密码" json:"-"`
	Status           int            `gorm:"type:tinyint;default:1;comment:状态 1:启用 0:禁用" json:"status"`
	GroupID          uint           `gorm:"comment:所属组ID" json:"groupId"`
	RealName         string         `gorm:"type:varchar(50);default:;comment:真实姓名" json:"realName"`
	IDCard           string         `gorm:"type:varchar(255);default:;comment:身份证号(AES加密)" json:"-"`
	IsReal           int            `gorm:"type:tinyint;default:0;comment:是否已实名 0否 1是" json:"isReal"`
	RegisterSource   string         `gorm:"type:varchar(20);default:web;comment:注册来源" json:"registerSource"`
	LastLoginAt      *time.Time     `gorm:"comment:最后登录时间" json:"lastLoginAt"`
	LastLoginIP      string         `gorm:"type:varchar(50);default:;comment:最后登录IP" json:"lastLoginIP"`
	LastLoginDevice  string         `gorm:"type:varchar(200);default:;comment:最后登录设备" json:"-"`
	LoginFailCount   int            `gorm:"default:0;comment:连续登录失败次数" json:"-"`
	LockUntil        *time.Time     `gorm:"comment:锁定截止时间" json:"-"`
	PasswordChangedAt *time.Time    `gorm:"comment:密码修改时间" json:"passwordChangedAt"`
	Remark           string         `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedAt        time.Time      `gorm:"comment:创建时间" json:"createTime"`
	UpdatedAt        time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt        gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 表名
func (User) TableName() string {
	return "sys_users"
}
