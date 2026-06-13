package model

import (
	"time"

	"gorm.io/gorm"
)

// WrongBook 错题本
type WrongBook struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint           `gorm:"index;not null;comment:用户ID" json:"userId"`
	QuestionID  uint           `gorm:"index;not null;comment:题目ID" json:"questionId"`
	CategoryID  uint           `gorm:"default:0;comment:所属分类ID" json:"categoryId"`
	WrongCount  int            `gorm:"default:1;comment:错误次数" json:"wrongCount"`
	LastWrongAt time.Time      `gorm:"comment:最后一次错误时间" json:"lastWrongAt"`
	IsMastered  bool           `gorm:"default:false;comment:是否已掌握" json:"isMastered"`
	MasteredAt  *time.Time     `gorm:"comment:掌握时间" json:"masteredAt,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (WrongBook) TableName() string {
	return "wrong_books"
}
