package model

import (
	"time"

	"gorm.io/gorm"
)

// QuestionVersion 题目版本快照表模型
type QuestionVersion struct {
	ID                    uint           `gorm:"primaryKey;autoIncrement;comment:版本ID" json:"id"`
	QuestionID            uint           `gorm:"not null;index;comment:题目ID" json:"questionId"`
	Version               int            `gorm:"not null;default:1;comment:版本号" json:"version"`
	VersionStatus         string         `gorm:"type:varchar(20);default:active;comment:版本状态 active/archived/deprecated" json:"versionStatus"`
	ChangeLog             string         `gorm:"type:varchar(500);default:;comment:本次发布变更说明" json:"changeLog"`

	// 快照字段 - 用户端展示/答题/判分必需
	Title                 string         `gorm:"type:varchar(500);not null;comment:快照：题目标题/摘要" json:"title"`
	QuestionType          string         `gorm:"type:varchar(50);not null;comment:快照：题型" json:"questionType"`
	Stem                  string         `gorm:"type:json;not null;comment:快照：题干富媒体内容块" json:"stem"`
	Content               string         `gorm:"type:json;comment:快照：题型结构内容" json:"content"`
	Answer                string         `gorm:"type:json;comment:快照：编辑态答案" json:"answer"`
	Analysis              string         `gorm:"type:json;comment:快照：解析富媒体内容块" json:"analysis"`
	Materials             string         `gorm:"type:json;comment:快照：材料" json:"materials"`
	ScoreRule             string         `gorm:"type:json;comment:快照：判分规则" json:"scoreRule"`

	// 分类和属性快照
	ExamID                uint           `gorm:"default:0;comment:快照：具体考试" json:"examId"`
	SubjectID             uint           `gorm:"default:0;comment:快照：科目/模块" json:"subjectId"`
	CategoryID            uint           `gorm:"default:0;comment:快照：章节分类" json:"categoryId"`
	SourceID              uint           `gorm:"default:0;comment:快照：题目来源" json:"sourceId"`
	Difficulty            int            `gorm:"type:tinyint;default:1;comment:快照：1简单 2中等 3困难" json:"difficulty"`
	ResourceType          string         `gorm:"type:varchar(20);default:private;comment:快照：public/private/group/user" json:"resourceType"`

	// 可见策略快照
	AnalysisVisiblePolicy string         `gorm:"type:varchar(30);default:after_answer;comment:快照：解析可见策略" json:"analysisVisiblePolicy"`
	AnswerVisiblePolicy   string         `gorm:"type:varchar(30);default:after_answer;comment:快照：答案可见策略" json:"answerVisiblePolicy"`

	// 组合题快照
	ParentID              uint           `gorm:"default:0;comment:快照：父题ID（组合题）" json:"parentId"`
	ParentVersionID       uint           `gorm:"default:0;comment:组合题父版本ID" json:"parentVersionId"`
	IsGroup               int            `gorm:"type:tinyint;default:0;comment:快照：是否组合题父题" json:"isGroup"`
	SubIndex              int            `gorm:"default:0;comment:快照：子题排序" json:"subIndex"`

	// 指纹
	StemHash              string         `gorm:"type:varchar(64);default:;comment:题干指纹" json:"stemHash"`
	ContentHash           string         `gorm:"type:varchar(64);default:;comment:内容指纹" json:"contentHash"`
	AnswerHash            string         `gorm:"type:varchar(64);default:;comment:答案指纹" json:"answerHash"`

	// 发布信息
	PublishedBy           uint           `gorm:"default:0;comment:发布人ID" json:"publishedBy"`
	AttachmentSnapshot    string         `gorm:"type:json;comment:附件快照" json:"attachmentSnapshot"`

	CreatedAt             time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"createTime"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"-"`
	DeletedAt             gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (QuestionVersion) TableName() string {
	return "qb_question_versions"
}
