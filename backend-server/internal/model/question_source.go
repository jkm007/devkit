package model

import (
	"time"

	"gorm.io/gorm"
)

// QuestionSource 题目来源模型
type QuestionSource struct {
	ID         uint           `gorm:"primaryKey;autoIncrement;comment:来源ID" json:"id"`
	SourceType string         `gorm:"type:varchar(30);not null;index;comment:来源类型" json:"sourceType"`
	Name       string         `gorm:"type:varchar(200);not null;comment:来源名称" json:"name"`
	ExamID     uint           `gorm:"default:0;index;comment:考试ID" json:"examId"`
	SubjectID  uint           `gorm:"default:0;index;comment:科目ID" json:"subjectId"`
	Year       int            `gorm:"default:0;comment:年份" json:"year"`
	Region     string         `gorm:"type:varchar(100);default:;comment:地区" json:"region"`
	PaperName  string         `gorm:"type:varchar(200);default:;comment:试卷名称" json:"paperName"`
	QuestionNo string         `gorm:"type:varchar(50);default:;comment:原题号" json:"questionNo"`
	Copyright  string         `gorm:"type:varchar(500);default:;comment:版权说明" json:"copyright"`
	CreatedBy  uint           `gorm:"not null;comment:创建人ID" json:"createdBy"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"createTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (QuestionSource) TableName() string {
	return "qb_question_sources"
}
