package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

type KnowledgePointRepo struct {
	db *gorm.DB
}

func NewKnowledgePointRepo(db *gorm.DB) *KnowledgePointRepo {
	return &KnowledgePointRepo{db: db}
}

func (r *KnowledgePointRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.KnowledgePoint, int64, error) {
	var items []model.KnowledgePoint
	var total int64
	query := r.db.Model(&model.KnowledgePoint{})

	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name LIKE ?", "%"+escapeLike(name.(string))+"%")
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if examId, ok := filters["examId"]; ok && examId != "" {
		query = query.Where("exam_id = ?", examId)
	}
	if subjectId, ok := filters["subjectId"]; ok && subjectId != "" {
		query = query.Where("subject_id = ?", subjectId)
	}
	if categoryId, ok := filters["categoryId"]; ok && categoryId != "" {
		query = query.Where("category_id = ?", categoryId)
	}
	if parentId, ok := filters["parentId"]; ok && parentId != "" {
		query = query.Where("parent_id = ?", parentId)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort_order ASC, created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *KnowledgePointRepo) GetAll() ([]model.KnowledgePoint, error) {
	var items []model.KnowledgePoint
	if err := r.db.Where("status = 1").Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *KnowledgePointRepo) GetByID(id uint) (*model.KnowledgePoint, error) {
	var item model.KnowledgePoint
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KnowledgePointRepo) Create(item *model.KnowledgePoint) error {
	return r.db.Create(item).Error
}

func (r *KnowledgePointRepo) Update(item *model.KnowledgePoint) error {
	return r.db.Save(item).Error
}

func (r *KnowledgePointRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.KnowledgePoint{}).Error
}

func (r *KnowledgePointRepo) HasChildren(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.KnowledgePoint{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
