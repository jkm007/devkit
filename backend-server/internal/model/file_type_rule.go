package model

import "time"

// FileTypeRule 文件类型检测规则模型
type FileTypeRule struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Extension   string    `json:"extension" gorm:"size:20;not null;uniqueIndex:uk_extension"`
	FileType    string    `json:"fileType" gorm:"size:50;not null"`
	Description string    `json:"description" gorm:"size:200;default:''"`
	Status      int8      `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (FileTypeRule) TableName() string {
	return "sys_file_type_rules"
}
