package model

import "time"

// Notification 通知消息模型
type Notification struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint       `gorm:"index;default:0;comment:目标用户ID 0=公告(全员)" json:"userId"`
	Type      string     `gorm:"type:varchar(50);not null;index;comment:通知类型" json:"type"`
	Title     string     `gorm:"type:varchar(200);not null;comment:标题" json:"title"`
	Content   string     `gorm:"type:text;comment:内容" json:"content"`
	Link      string     `gorm:"type:varchar(500);default:;comment:跳转链接" json:"link"`
	IsRead    bool       `gorm:"default:false;comment:是否已读" json:"isRead"`
	SenderID  uint       `gorm:"default:0;comment:发送者ID 0=系统" json:"senderId"`
	CreatedAt time.Time  `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"comment:更新时间" json:"-"`
}

// TableName 表名
func (Notification) TableName() string {
	return "sys_notifications"
}

// NotificationRead 用户已读记录（用于公告类通知）
type NotificationRead struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	NotificationID uint      `gorm:"not null;uniqueIndex:uk_notif_user;comment:通知ID" json:"notificationId"`
	UserID         uint      `gorm:"not null;uniqueIndex:uk_notif_user;comment:用户ID" json:"userId"`
	ReadAt         time.Time `gorm:"comment:阅读时间" json:"readAt"`
}

// TableName 表名
func (NotificationRead) TableName() string {
	return "sys_notification_reads"
}
