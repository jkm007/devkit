package model

import "time"

// PasswordHistory 密码历史模型
type PasswordHistory struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID    uint      `gorm:"not null;comment:用户ID" json:"userId"`
	Password  string    `gorm:"type:varchar(255);not null;comment:历史密码(bcrypt)" json:"-"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"createdAt"`
}

// TableName 表名
func (PasswordHistory) TableName() string {
	return "sys_password_history"
}
