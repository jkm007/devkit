package model

import "time"

// StorageConfig 存储连接配置模型
// 用于管理多个同类型的存储连接配置（如多个 MinIO、多个 OSS 等）
type StorageConfig struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:100;not null;uniqueIndex:uk_name"`
	Driver      string    `json:"driver" gorm:"size:20;not null;index:idx_driver"`
	Endpoint    string    `json:"endpoint" gorm:"size:500"`
	AccessKey   string    `json:"accessKey" gorm:"size:500"`
	SecretKey   string    `json:"secretKey" gorm:"size:500"`
	Bucket      string    `json:"bucket" gorm:"size:200"`
	Region      string    `json:"region" gorm:"size:100"`
	UseSSL      bool      `json:"useSsl" gorm:"default:false"`
	CDNDomain   string    `json:"cdnDomain" gorm:"size:500"`
	IsDefault           bool      `json:"isDefault" gorm:"default:false"`
	PresignedURLExpiry  int       `json:"presignedUrlExpiry" gorm:"not null;default:3600"`
	Status              int8      `json:"status" gorm:"default:1;index:idx_status"`
	Description string    `json:"description" gorm:"size:500"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (StorageConfig) TableName() string {
	return "sys_storage_config"
}
