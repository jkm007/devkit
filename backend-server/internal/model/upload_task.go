package model

import "time"

// UploadTask 上传任务
type UploadTask struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FileHash    string    `gorm:"index;size:64;comment:文件SHA-256哈希" json:"fileHash"`
	UploadID    string    `gorm:"uniqueIndex;size:128;comment:MinIO上传ID" json:"uploadId"`
	FileName    string    `gorm:"size:255;comment:原始文件名" json:"fileName"`
	FileSize    int64      `gorm:"comment:文件大小(字节)" json:"fileSize"`
	ContentType string    `gorm:"size:128;comment:MIME类型" json:"contentType"`
	ObjectKey   string    `gorm:"size:500;comment:最终存储路径" json:"objectKey"`
	TotalParts  int       `gorm:"comment:总分片数" json:"totalParts"`
	Status      string    `gorm:"size:20;default:uploading;comment:状态" json:"status"` // uploading, completed, aborted
	UserID      uint      `gorm:"index;comment:上传者ID" json:"userId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (UploadTask) TableName() string { return "sys_upload_tasks" }

// UploadedPart 已上传分片
type UploadedPart struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UploadTaskID uint      `gorm:"index;comment:关联上传任务" json:"uploadTaskId"`
	PartNumber   int       `gorm:"comment:分片序号" json:"partNumber"`
	ETag         string    `gorm:"size:64;comment:分片ETag" json:"etag"`
	Size         int64      `gorm:"comment:分片大小" json:"size"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (UploadedPart) TableName() string { return "sys_uploaded_parts" }
