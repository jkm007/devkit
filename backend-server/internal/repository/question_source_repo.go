package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

type QuestionSourceRepo struct {
	db *gorm.DB
}

func NewQuestionSourceRepo(db *gorm.DB) *QuestionSourceRepo {
	return &QuestionSourceRepo{db: db}
}

func (r *QuestionSourceRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.QuestionSource, int64, error) {
	var items []model.QuestionSource
	var total int64
	query := r.db.Model(&model.QuestionSource{})

	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name LIKE ?", "%"+escapeLike(name.(string))+"%")
	}
	if sourceType, ok := filters["sourceType"]; ok && sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}
	if examId, ok := filters["examId"]; ok && examId != "" {
		query = query.Where("exam_id = ?", examId)
	}
	if year, ok := filters["year"]; ok && year != "" {
		query = query.Where("year = ?", year)
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

func (r *QuestionSourceRepo) GetByID(id uint) (*model.QuestionSource, error) {
	var item model.QuestionSource
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *QuestionSourceRepo) Create(item *model.QuestionSource) error {
	return r.db.Create(item).Error
}

func (r *QuestionSourceRepo) Update(item *model.QuestionSource) error {
	return r.db.Save(item).Error
}

func (r *QuestionSourceRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.QuestionSource{}).Error
}
