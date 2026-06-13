package model

import (
	"time"

	"gorm.io/gorm"
)

// KnowledgePoint 知识点/考点模型
type KnowledgePoint struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;comment:知识点ID" json:"id"`
	ExamID      uint           `gorm:"default:0;index;comment:考试ID" json:"examId"`
	SubjectID   uint           `gorm:"default:0;index;comment:科目ID" json:"subjectId"`
	CategoryID  uint           `gorm:"default:0;index;comment:章节分类ID" json:"categoryId"`
	ParentID    uint           `gorm:"default:0;index;comment:父级ID" json:"parentId"`
	Name        string         `gorm:"type:varchar(150);not null;comment:知识点名称" json:"name"`
	Code        string         `gorm:"type:varchar(50);default:;comment:编码" json:"code"`
	Path        string         `gorm:"type:varchar(1000);default:;comment:路径" json:"path"`
	Level       int            `gorm:"default:1;comment:层级" json:"level"`
	Importance  int            `gorm:"type:tinyint;default:3;comment:重要程度 1-5" json:"importance"`
	Description string         `gorm:"type:varchar(1000);default:;comment:描述" json:"description"`
	SortOrder   int            `gorm:"default:0;comment:排序" json:"sortOrder"`
	Status      int            `gorm:"type:tinyint;default:1;comment:状态 1:启用 0:禁用" json:"status"`
	CreatedBy   uint           `gorm:"not null;comment:创建人ID" json:"createdBy"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"createTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (KnowledgePoint) TableName() string {
	return "qb_knowledge_points"
}
