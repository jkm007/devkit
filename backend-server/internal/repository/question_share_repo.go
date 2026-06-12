package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

type QuestionShareRepo struct {
	db *gorm.DB
}

func NewQuestionShareRepo(db *gorm.DB) *QuestionShareRepo {
	return &QuestionShareRepo{db: db}
}

func (r *QuestionShareRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.QuestionShare, int64, error) {
	var items []model.QuestionShare
	var total int64
	query := r.db.Model(&model.QuestionShare{})

	if questionId, ok := filters["questionId"]; ok && questionId != "" {
		query = query.Where("question_id = ?", questionId)
	}
	if shareType, ok := filters["shareType"]; ok && shareType != "" {
		query = query.Where("share_type = ?", shareType)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if createdBy, ok := filters["createdBy"]; ok && createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *QuestionShareRepo) GetByID(id uint) (*model.QuestionShare, error) {
	var item model.QuestionShare
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *QuestionShareRepo) GetByCode(code string) (*model.QuestionShare, error) {
	var item model.QuestionShare
	if err := r.db.Where("share_code = ?", code).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *QuestionShareRepo) Create(item *model.QuestionShare) error {
	return r.db.Create(item).Error
}

func (r *QuestionShareRepo) Update(item *model.QuestionShare) error {
	return r.db.Save(item).Error
}

func (r *QuestionShareRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.QuestionShare{}).Error
}
