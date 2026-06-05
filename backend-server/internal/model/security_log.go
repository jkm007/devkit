package model

import "time"

// SecurityLog 安全日志模型
type SecurityLog struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID      uint      `gorm:"not null;comment:用户ID" json:"userId"`
	Username    string    `gorm:"->;-:migration;column:username" json:"username"`
	EventType   string    `gorm:"type:varchar(30);not null;comment:事件类型" json:"eventType"`
	EventDetail string    `gorm:"type:varchar(500);default:;comment:事件详情" json:"eventDetail"`
	IP          string    `gorm:"type:varchar(50);default:;comment:IP地址" json:"ip"`
	UserAgent   string    `gorm:"type:varchar(500);default:;comment:User-Agent" json:"userAgent"`
	Status      int       `gorm:"type:tinyint;default:1;comment:状态 0失败 1成功" json:"status"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"createdAt"`
}

// TableName 表名
func (SecurityLog) TableName() string {
	return "sys_security_logs"
}
