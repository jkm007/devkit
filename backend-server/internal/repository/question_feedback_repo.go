package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// QuestionFeedbackRepo 题目纠错数据访问
type QuestionFeedbackRepo struct {
	db *gorm.DB
}

// NewQuestionFeedbackRepo 创建题目纠错仓库
func NewQuestionFeedbackRepo(db *gorm.DB) *QuestionFeedbackRepo {
	return &QuestionFeedbackRepo{db: db}
}

// Create 创建纠错反馈
func (r *QuestionFeedbackRepo) Create(fb *model.QuestionFeedback) error {
	return r.db.Create(fb).Error
}

// List 获取用户的纠错列表
func (r *QuestionFeedbackRepo) List(userID uint, offset, limit int) ([]model.QuestionFeedback, int64, error) {
	var feedbacks []model.QuestionFeedback
	var total int64

	r.db.Model(&model.QuestionFeedback{}).Where("user_id = ?", userID).Count(&total)

	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&feedbacks).Error

	return feedbacks, total, err
}

// GetByID 获取纠错详情
func (r *QuestionFeedbackRepo) GetByID(userID, id uint) (*model.QuestionFeedback, error) {
	var fb model.QuestionFeedback
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&fb).Error
	return &fb, err
}

// Delete 删除纠错反馈
func (r *QuestionFeedbackRepo) Delete(userID, id uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.QuestionFeedback{}).Error
}

// AdminList 管理端列表
func (r *QuestionFeedbackRepo) AdminList(status string, offset, limit int) ([]model.QuestionFeedback, int64, error) {
	var feedbacks []model.QuestionFeedback
	var total int64

	query := r.db.Model(&model.QuestionFeedback{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	err := query.Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&feedbacks).Error

	return feedbacks, total, err
}

// AdminUpdate 管理端更新
func (r *QuestionFeedbackRepo) AdminUpdate(id uint, status, reply string) error {
	return r.db.Model(&model.QuestionFeedback{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"admin_reply": reply,
		}).Error
}
