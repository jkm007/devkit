package model

import "time"

// FileAsset 文件资产（秒传映射）
type FileAsset struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FileHash    string    `gorm:"uniqueIndex;size:255;comment:文件哈希" json:"fileHash"`
	ObjectKey   string    `gorm:"size:500;comment:存储路径" json:"objectKey"`
	FileName    string    `gorm:"size:255;comment:原始文件名" json:"fileName"`
	FileSize    int64     `gorm:"comment:文件大小(字节)" json:"fileSize"`
	ContentType string    `gorm:"size:128;comment:MIME类型" json:"contentType"`
	StorageType string    `gorm:"size:32;default:local;comment:存储类型(local/minio/oss/cos/ceph)" json:"storageType"`
	Status      string    `gorm:"size:20;default:active;comment:状态(active/inaccessible)" json:"status"`
	RefCount    int       `gorm:"default:1;comment:引用计数" json:"refCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (FileAsset) TableName() string { return "sys_file_assets" }
