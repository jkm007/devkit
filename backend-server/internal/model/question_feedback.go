package model

import "time"

// QuestionFeedback 题目纠错反馈
type QuestionFeedback struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"userId"`
	QuestionID  uint      `gorm:"index;not null" json:"questionId"`
	FeedbackType string   `gorm:"type:varchar(50);not null" json:"feedbackType"` // answer_error / content_error / option_error / other
	Description string    `gorm:"type:text;not null" json:"description"`
	Suggestion  string    `gorm:"type:text" json:"suggestion"`
	Status      string    `gorm:"type:varchar(20);default:pending" json:"status"` // pending / processing / resolved / rejected
	AdminReply  string    `gorm:"type:text" json:"adminReply"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (QuestionFeedback) TableName() string {
	return "question_feedbacks"
}

// 反馈类型常量
const (
	FeedbackTypeErrorAnswer  = "answer_error"  // 答案错误
	FeedbackTypeErrorContent = "content_error" // 内容错误
	FeedbackTypeErrorOption  = "option_error"  // 选项错误
	FeedbackTypeErrorOther   = "other"         // 其他
)

// 反馈状态常量
const (
	FeedbackStatusPending     = "pending"
	FeedbackStatusProcessing  = "processing"
	FeedbackStatusResolved    = "resolved"
	FeedbackStatusRejected    = "rejected"
)
