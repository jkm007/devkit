package model

import "time"

// FileShare 文件分享
type FileShare struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	FileID       uint       `gorm:"index;comment:文件ID" json:"fileId"`
	FolderID     uint       `gorm:"index;comment:文件夹ID" json:"folderId"`
	ShareCode    string     `gorm:"uniqueIndex;size:64;comment:分享码" json:"shareCode"`
	UserID       uint       `gorm:"index;comment:分享者" json:"userId"`
	Password     string     `gorm:"size:255;comment:访问密码(bcrypt哈希,空表示无密码)" json:"-"`
	HasPassword  bool       `gorm:"default:false;comment:是否设置了密码" json:"hasPassword"`
	Status       int        `gorm:"default:1;comment:状态(1=有效,2=已过期,3=已禁用)" json:"status"`
	ExpireAt     *time.Time `gorm:"comment:过期时间(可选)" json:"expireAt"`
	AccessCount  int        `gorm:"default:0;comment:访问次数" json:"accessCount"`
	MaxAccess    int        `gorm:"default:0;comment:最大访问次数(0=无限)" json:"maxAccess"`
	IsPublic     bool       `gorm:"default:false;comment:是否公开" json:"isPublic"`
	AccessedAt   *time.Time `gorm:"comment:最后访问时间" json:"accessedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
}

func (FileShare) TableName() string { return "sys_file_shares" }