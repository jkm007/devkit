package model

import "time"

// StorageBucket 存储桶配置模型
type StorageBucket struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name       string    `json:"name" gorm:"size:100;not null;uniqueIndex:uk_name"`
	Driver     string    `json:"driver" gorm:"size:20;not null;index:idx_driver"`
	Endpoint   string    `json:"endpoint" gorm:"size:500"`
	Bucket     string    `json:"bucket" gorm:"size:200"`
	AccessKey  string    `json:"accessKey" gorm:"size:500"`
	SecretKey  string    `json:"secretKey" gorm:"size:500"`
	Region     string    `json:"region" gorm:"size:100"`
	UseSSL     bool      `json:"useSsl" gorm:"default:false"`
	CDNDomain  string    `json:"cdnDomain" gorm:"size:500"`
	PathPrefix string    `json:"pathPrefix" gorm:"size:500"`
	Purpose    string    `json:"purpose" gorm:"size:100;index:idx_purpose"`
	IsDefault  bool      `json:"isDefault" gorm:"default:false"`
	Status     int8      `json:"status" gorm:"default:1;index:idx_status"`
	Description string   `json:"description" gorm:"size:500"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (StorageBucket) TableName() string {
	return "sys_storage_bucket"
}
