package model

import "time"

// UserPrivacy 用户隐私设置模型
type UserPrivacy struct {
	ID             uint      `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID         uint      `gorm:"uniqueIndex;not null;comment:用户ID" json:"userId"`
	ProfileVisible int       `gorm:"type:tinyint;default:1;comment:资料可见性 1全部 2仅班级 3仅自己" json:"profileVisible"`
	RealnameVisible int      `gorm:"type:tinyint;default:1;comment:真实姓名可见性" json:"realnameVisible"`
	EmailVisible   int       `gorm:"type:tinyint;default:1;comment:邮箱可见性" json:"emailVisible"`
	StatsVisible   int       `gorm:"type:tinyint;default:1;comment:统计可见性" json:"statsVisible"`
	ClassVisible   int       `gorm:"type:tinyint;default:1;comment:班级可见性" json:"classVisible"`
	CreatedAt      time.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"comment:更新时间" json:"-"`
}

// TableName 表名
func (UserPrivacy) TableName() string {
	return "sys_user_privacy"
}
