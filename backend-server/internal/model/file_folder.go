package model

import "time"

// FileFolder 文件夹
type FileFolder struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:255;comment:文件夹名称" json:"name"`
	ParentID  *uint     `gorm:"index;comment:父文件夹ID(null=根目录)" json:"parentId"`
	Path      string    `gorm:"index:,length:255;size:1000;comment:物化路径" json:"path"`
	Type      string    `gorm:"size:20;default:'normal';comment:文件夹类型(normal/avatar)" json:"type"`
	UserID    uint      `gorm:"index;comment:所属用户" json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (FileFolder) TableName() string { return "sys_file_folders" }
