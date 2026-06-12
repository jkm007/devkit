package model

import (
	"time"

	"gorm.io/gorm"
)

// Question 题目主表模型
type Question struct {
	ID                   uint           `gorm:"primaryKey;autoIncrement;comment:题目ID" json:"id"`
	Title                string         `gorm:"type:varchar(500);not null;comment:题目标题/摘要" json:"title"`
	QuestionType         string         `gorm:"type:varchar(50);not null;index;comment:题型" json:"questionType"`
	Stem                 string         `gorm:"type:json;not null;comment:题干富媒体内容块" json:"stem"`
	Content              string         `gorm:"type:json;comment:题型结构内容" json:"content"`
	Answer               string         `gorm:"type:json;comment:编辑态答案" json:"answer"`
	Analysis             string         `gorm:"type:json;comment:解析富媒体内容块" json:"analysis"`
	Materials            string         `gorm:"type:json;comment:材料" json:"materials"`
	ScoreRule            string         `gorm:"type:json;comment:判分规则" json:"scoreRule"`
	ExamID               uint           `gorm:"default:0;index;comment:考试ID" json:"examId"`
	SubjectID            uint           `gorm:"default:0;index;comment:科目ID" json:"subjectId"`
	CategoryID           uint           `gorm:"default:0;index;comment:章节分类ID" json:"categoryId"`
	SourceID             uint           `gorm:"default:0;comment:来源ID" json:"sourceId"`
	Difficulty           int            `gorm:"type:tinyint;default:1;comment:难度 1简单 2中等 3困难" json:"difficulty"`
	ResourceType         string         `gorm:"type:varchar(20);default:private;comment:资源类型" json:"resourceType"`
	Status               string         `gorm:"type:varchar(20);default:draft;index;comment:状态" json:"status"`
	CurrentVersionID     uint           `gorm:"default:0;comment:当前版本ID" json:"currentVersionId"`
	SourceImportTaskID   uint           `gorm:"default:0;comment:导入任务ID" json:"sourceImportTaskId"`
	ParentID             uint           `gorm:"default:0;index;comment:父题ID" json:"parentId"`
	IsGroup              int            `gorm:"type:tinyint;default:0;comment:是否组合题" json:"isGroup"`
	SubIndex             int            `gorm:"default:0;comment:子题排序" json:"subIndex"`
	StemHash             string         `gorm:"type:varchar(64);default:;comment:题干指纹" json:"stemHash"`
	ContentHash          string         `gorm:"type:varchar(64);default:;comment:内容指纹" json:"contentHash"`
	AnswerHash           string         `gorm:"type:varchar(64);default:;comment:答案指纹" json:"answerHash"`
	AnalysisVisiblePolicy string       `gorm:"type:varchar(30);default:after_answer;comment:解析可见策略" json:"analysisVisiblePolicy"`
	AnswerVisiblePolicy  string         `gorm:"type:varchar(30);default:after_answer;comment:答案可见策略" json:"answerVisiblePolicy"`
	CreatedBy            uint           `gorm:"not null;index;comment:创建人ID" json:"createdBy"`
	ReviewedBy           uint           `gorm:"default:0;comment:审核人ID" json:"reviewedBy"`
	ReviewedAt           *time.Time     `gorm:"comment:审核时间" json:"reviewedAt"`
	RejectReason         string         `gorm:"type:varchar(500);default:;comment:驳回原因" json:"rejectReason"`
	PublishedAt          *time.Time     `gorm:"comment:发布时间" json:"publishedAt"`
	DeletedBy            uint           `gorm:"default:0;comment:删除人ID" json:"deletedBy"`
	RecycleExpireAt      *time.Time     `gorm:"comment:回收站过期时间" json:"recycleExpireAt"`
	CreatedAt            time.Time      `gorm:"comment:创建时间" json:"createTime"`
	UpdatedAt            time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt            gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (Question) TableName() string {
	return "qb_questions"
}
