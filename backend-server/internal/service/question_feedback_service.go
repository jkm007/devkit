package service

import (
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"time"
)

// QuestionFeedbackService 题目纠错服务
type QuestionFeedbackService struct {
	repo *repository.QuestionFeedbackRepo
}

// NewQuestionFeedbackService 创建题目纠错服务
func NewQuestionFeedbackService() *QuestionFeedbackService {
	return &QuestionFeedbackService{
		repo: repository.NewQuestionFeedbackRepo(database.GetMySQL()),
	}
}

// FeedbackResponse 纠错响应
type FeedbackResponse struct {
	ID           uint      `json:"id"`
	QuestionID   uint      `json:"questionId"`
	FeedbackType string    `json:"feedbackType"`
	Description  string    `json:"description"`
	Suggestion   string    `json:"suggestion"`
	Status       string    `json:"status"`
	AdminReply   string    `json:"adminReply"`
	CreatedAt    time.Time `json:"createdAt"`
}

// CreateFeedbackRequest 创建请求
type CreateFeedbackRequest struct {
	QuestionID  uint   `json:"questionId" binding:"required"`
	FeedbackType string `json:"feedbackType" binding:"required"`
	Description string `json:"description" binding:"required"`
	Suggestion  string `json:"suggestion"`
}

// Create 创建纠错反馈
func (s *QuestionFeedbackService) Create(userID uint, req *CreateFeedbackRequest) error {
	fb := &model.QuestionFeedback{
		UserID:      userID,
		QuestionID:  req.QuestionID,
		FeedbackType: req.FeedbackType,
		Description: req.Description,
		Suggestion:  req.Suggestion,
		Status:      model.FeedbackStatusPending,
	}
	return s.repo.Create(fb)
}

// List 获取纠错列表
func (s *QuestionFeedbackService) List(userID uint, page, pageSize int) ([]FeedbackResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	items, total, err := s.repo.List(userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	results := make([]FeedbackResponse, 0, len(items))
	for _, fb := range items {
		results = append(results, FeedbackResponse{
			ID:           fb.ID,
			QuestionID:   fb.QuestionID,
			FeedbackType: fb.FeedbackType,
			Description:  fb.Description,
			Suggestion:   fb.Suggestion,
			Status:       fb.Status,
			AdminReply:   fb.AdminReply,
			CreatedAt:    fb.CreatedAt,
		})
	}

	return results, total, nil
}

// GetByID 获取纠错详情
func (s *QuestionFeedbackService) GetByID(userID, id uint) (*FeedbackResponse, error) {
	fb, err := s.repo.GetByID(userID, id)
	if err != nil {
		return nil, err
	}

	return &FeedbackResponse{
		ID:           fb.ID,
		QuestionID:   fb.QuestionID,
		FeedbackType: fb.FeedbackType,
		Description:  fb.Description,
		Suggestion:   fb.Suggestion,
		Status:       fb.Status,
		AdminReply:   fb.AdminReply,
		CreatedAt:    fb.CreatedAt,
	}, nil
}

// Delete 删除纠错反馈
func (s *QuestionFeedbackService) Delete(userID, id uint) error {
	return s.repo.Delete(userID, id)
}

// AdminList 管理端列表
func (s *QuestionFeedbackService) AdminList(status string, page, pageSize int) ([]model.QuestionFeedback, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.repo.AdminList(status, offset, pageSize)
}

// AdminUpdate 管理端更新
func (s *QuestionFeedbackService) AdminUpdate(id uint, status, reply string) error {
	return s.repo.AdminUpdate(id, status, reply)
}
