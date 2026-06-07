package model

import "time"

// FileShare 文件分享
type FileShare struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	FileID       uint       `gorm:"index;comment:文件ID" json:"fileId"`
	FolderID     uint       `gorm:"index;comment:文件夹ID" json:"folderId"`
	ShareCode    string     `gorm:"uniqueIndex;size:16;comment:分享码" json:"shareCode"`
	UserID       uint       `gorm:"comment:分享者" json:"userId"`
	ExpireAt     *time.Time `gorm:"comment:过期时间(可选)" json:"expireAt"`
	AccessCount  int        `gorm:"default:0;comment:访问次数" json:"accessCount"`
	MaxAccess    int        `gorm:"default:0;comment:最大访问次数(0=无限)" json:"maxAccess"`
	IsPublic     bool       `gorm:"default:false;comment:是否公开" json:"isPublic"`
	CreatedAt    time.Time  `json:"createdAt"`
}

func (FileShare) TableName() string { return "sys_file_shares" }