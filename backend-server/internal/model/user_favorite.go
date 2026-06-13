package model

import (
	"time"

	"gorm.io/gorm"
)

// UserFavorite 用户收藏
type UserFavorite struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"index;not null;comment:用户ID" json:"userId"`
	QuestionID uint      `gorm:"index;not null;comment:题目ID" json:"questionId"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:收藏时间" json:"createdAt"`
}

// TableName 表名
func (UserFavorite) TableName() string {
	return "user_favorites"
}

// UserNote 用户笔记
type UserNote struct {
	ID         uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint           `gorm:"index;not null;comment:用户ID" json:"userId"`
	QuestionID uint           `gorm:"index;not null;comment:题目ID" json:"questionId"`
	Content    string         `gorm:"type:text;not null;comment:笔记内容" json:"content"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"createdAt"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (UserNote) TableName() string {
	return "user_notes"
}

// PracticeRecord 练习记录
type PracticeRecord struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"index;not null;comment:用户ID" json:"userId"`
	Mode      string    `gorm:"type:varchar(50);not null;comment:练习模式" json:"mode"`
	Total     int       `gorm:"not null;comment:总题数" json:"total"`
	Answered  int       `gorm:"not null;comment:已答题数" json:"answered"`
	Correct   int       `gorm:"default:0;comment:正确题数" json:"correct"`
	Elapsed   int       `gorm:"not null;comment:用时(秒)" json:"elapsed"`
	Answers   string    `gorm:"type:text;comment:答案JSON" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:完成时间" json:"createdAt"`
}

// TableName 表名
func (PracticeRecord) TableName() string {
	return "practice_records"
}
