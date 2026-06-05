package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，所有模型嵌入此结构体
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
