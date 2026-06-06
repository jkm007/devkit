package model

import "time"

// FileEntry 文件条目
type FileEntry struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FolderID    uint      `gorm:"index;comment:所属文件夹" json:"folderId"`
	FileAssetID uint      `gorm:"comment:关联文件资产" json:"fileAssetId"`
	Name        string    `gorm:"size:255;comment:文件名" json:"name"`
	Size        int64      `gorm:"comment:文件大小" json:"size"`
	ContentType string    `gorm:"size:128;comment:MIME类型" json:"contentType"`
	UserID      uint      `gorm:"index;comment:上传者" json:"userId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (FileEntry) TableName() string { return "sys_file_entries" }
