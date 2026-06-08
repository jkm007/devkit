package model

import "time"

// FileTag 文件标签关联模型
type FileTag struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	FileID    uint      `json:"fileId" gorm:"not null;uniqueIndex:uk_file_tag"`
	TagID     int64     `json:"tagId" gorm:"not null;uniqueIndex:uk_file_tag"`
	Source    string    `json:"source" gorm:"size:10;not null;default:auto"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`

	// 关联
	Tag *Tag `json:"tag,omitempty" gorm:"foreignKey:TagID"`
}

func (FileTag) TableName() string {
	return "sys_file_tag"
}
