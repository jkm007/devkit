package model

import (
	"time"

	"gorm.io/gorm"
)

// ExamCategory 考试大类模型
type ExamCategory struct {
	ID        uint           `gorm:"primaryKey;autoIncrement;comment:考试大类ID" json:"id"`
	ParentID  uint           `gorm:"default:0;index;comment:父级ID" json:"parentId"`
	Name      string         `gorm:"type:varchar(100);not null;comment:考试大类名称" json:"name"`
	Code      string         `gorm:"type:varchar(50);default:;comment:分类编码" json:"code"`
	Path      string         `gorm:"type:varchar(1000);default:;comment:路径" json:"path"`
	Level     int            `gorm:"default:1;comment:层级" json:"level"`
	SortOrder int            `gorm:"default:0;comment:排序" json:"sortOrder"`
	Status    int            `gorm:"type:tinyint;default:1;comment:状态 1:启用 0:禁用" json:"status"`
	CreatedBy uint           `gorm:"not null;comment:创建人ID" json:"createdBy"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"createTime"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (ExamCategory) TableName() string {
	return "qb_exam_categories"
}

// Exam 具体考试模型
type Exam struct {
	ID              uint           `gorm:"primaryKey;autoIncrement;comment:考试ID" json:"id"`
	ExamCategoryID  uint           `gorm:"not null;index;comment:考试大类ID" json:"examCategoryId"`
	Name            string         `gorm:"type:varchar(150);not null;comment:考试名称" json:"name"`
	Code            string         `gorm:"type:varchar(50);default:;comment:编码" json:"code"`
	Description     string         `gorm:"type:varchar(1000);default:;comment:描述" json:"description"`
	Status          int            `gorm:"type:tinyint;default:1;comment:状态 1:启用 0:禁用" json:"status"`
	SortOrder       int            `gorm:"default:0;comment:排序" json:"sortOrder"`
	CreatedBy       uint           `gorm:"not null;comment:创建人ID" json:"createdBy"`
	CreatedAt       time.Time      `gorm:"comment:创建时间" json:"createTime"`
	UpdatedAt       time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt       gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (Exam) TableName() string {
	return "qb_exams"
}

// Subject 科目/模块模型
type Subject struct {
	ID        uint           `gorm:"primaryKey;autoIncrement;comment:科目ID" json:"id"`
	ExamID    uint           `gorm:"not null;index;comment:考试ID" json:"examId"`
	Name      string         `gorm:"type:varchar(150);not null;comment:科目名称" json:"name"`
	Code      string         `gorm:"type:varchar(50);default:;comment:编码" json:"code"`
	SortOrder int            `gorm:"default:0;comment:排序" json:"sortOrder"`
	Status    int            `gorm:"type:tinyint;default:1;comment:状态 1:启用 0:禁用" json:"status"`
	CreatedBy uint           `gorm:"not null;comment:创建人ID" json:"createdBy"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"createTime"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (Subject) TableName() string {
	return "qb_subjects"
}

// Category 章节分类模型
type Category struct {
	ID        uint           `gorm:"primaryKey;autoIncrement;comment:分类ID" json:"id"`
	ExamID    uint           `gorm:"default:0;index;comment:考试ID" json:"examId"`
	SubjectID uint           `gorm:"default:0;index;comment:科目ID" json:"subjectId"`
	ParentID  uint           `gorm:"default:0;index;comment:父级ID" json:"parentId"`
	Name      string         `gorm:"type:varchar(100);not null;comment:分类名称" json:"name"`
	Path      string         `gorm:"type:varchar(1000);default:;comment:路径" json:"path"`
	Level     int            `gorm:"default:1;comment:层级" json:"level"`
	SortOrder int            `gorm:"default:0;comment:排序" json:"sortOrder"`
	Status    int            `gorm:"type:tinyint;default:1;comment:状态 1:启用 0:禁用" json:"status"`
	CreatedBy uint           `gorm:"not null;comment:创建人ID" json:"createdBy"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"createTime"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (Category) TableName() string {
	return "qb_categories"
}
