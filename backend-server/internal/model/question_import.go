package model

import (
	"time"

	"gorm.io/gorm"
)

// QuestionImportTask 题目导入任务模型
type QuestionImportTask struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement;comment:任务ID" json:"id"`
	FileID             uint           `gorm:"not null;index;comment:文件ID" json:"fileId"`
	FileName           string         `gorm:"type:varchar(255);not null;comment:文件名" json:"fileName"`
	FileType           string         `gorm:"type:varchar(20);not null;comment:文件类型" json:"fileType"`
	Status             string         `gorm:"type:varchar(30);default:uploaded;index;comment:状态 uploaded/parsing/parsed/partial_failed/failed/confirmed/published/cancelled" json:"status"`
	TotalCount         int            `gorm:"default:0;comment:总题数" json:"totalCount"`
	SuccessCount       int            `gorm:"default:0;comment:成功数" json:"successCount"`
	FailedCount        int            `gorm:"default:0;comment:失败数" json:"failedCount"`
	ErrorReport        string         `gorm:"type:json;comment:错误报告" json:"errorReport"`
	TargetCategoryID   uint           `gorm:"default:0;comment:目标分类ID" json:"targetCategoryId"`
	TargetResourceType string         `gorm:"type:varchar(20);default:private;comment:目标资源类型" json:"targetResourceType"`
	TargetScopeType    string         `gorm:"type:varchar(20);default:user;comment:目标可见范围类型" json:"targetScopeType"`
	TargetScopeID      uint           `gorm:"default:0;comment:目标可见范围ID" json:"targetScopeId"`
	CreatedBy          uint           `gorm:"not null;index;comment:创建人ID" json:"createdBy"`
	ConfirmedAt        *time.Time     `gorm:"comment:确认时间" json:"confirmedAt"`
	CreatedAt          time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"createTime"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"-"`
	DeletedAt          gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (QuestionImportTask) TableName() string {
	return "qb_question_import_tasks"
}

// QuestionImportItem 题目导入明细模型
type QuestionImportItem struct {
	ID                  uint      `gorm:"primaryKey;autoIncrement;comment:明细ID" json:"id"`
	TaskID              uint      `gorm:"not null;index;comment:任务ID" json:"taskId"`
	RowNo               int       `gorm:"default:0;comment:行号" json:"rowNo"`
	QuestionNo          string    `gorm:"type:varchar(50);default:;comment:题号" json:"questionNo"`
	ParseStatus         string    `gorm:"type:varchar(20);not null;comment:解析状态 success/failed/skipped" json:"parseStatus"`
	QuestionID          uint      `gorm:"default:0;index;comment:题目ID" json:"questionId"`
	ErrorCode           string    `gorm:"type:varchar(50);default:;comment:错误码 parse_error/format_error/missing_field/duplicate/duplicate_found" json:"errorCode"`
	ErrorMessage        string    `gorm:"type:varchar(1000);default:;comment:错误信息" json:"errorMessage"`
	DuplicateQuestionID uint      `gorm:"default:0;comment:查重发现的疑似重复题ID" json:"duplicateQuestionId"`
	RawContent          string    `gorm:"type:json;comment:原始内容" json:"rawContent"`
	CreatedAt           time.Time `gorm:"autoCreateTime;comment:创建时间" json:"createTime"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"-"`
}

func (QuestionImportItem) TableName() string {
	return "qb_question_import_items"
}
