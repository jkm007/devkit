package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

type QuestionImportTaskRepo struct {
	db *gorm.DB
}

func NewQuestionImportTaskRepo(db *gorm.DB) *QuestionImportTaskRepo {
	return &QuestionImportTaskRepo{db: db}
}

func (r *QuestionImportTaskRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.QuestionImportTask, int64, error) {
	var items []model.QuestionImportTask
	var total int64
	query := r.db.Model(&model.QuestionImportTask{})

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

func (r *QuestionImportTaskRepo) GetByID(id uint) (*model.QuestionImportTask, error) {
	var item model.QuestionImportTask
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *QuestionImportTaskRepo) Create(item *model.QuestionImportTask) error {
	return r.db.Create(item).Error
}

func (r *QuestionImportTaskRepo) Update(item *model.QuestionImportTask) error {
	return r.db.Save(item).Error
}

func (r *QuestionImportTaskRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.QuestionImportTask{}).Error
}

// 导入明细
type QuestionImportItemRepo struct {
	db *gorm.DB
}

func NewQuestionImportItemRepo(db *gorm.DB) *QuestionImportItemRepo {
	return &QuestionImportItemRepo{db: db}
}

func (r *QuestionImportItemRepo) ListByTaskID(taskID uint) ([]model.QuestionImportItem, error) {
	var items []model.QuestionImportItem
	if err := r.db.Where("task_id = ?", taskID).Order("row_no ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *QuestionImportItemRepo) Create(item *model.QuestionImportItem) error {
	return r.db.Create(item).Error
}

func (r *QuestionImportItemRepo) BatchCreate(items []model.QuestionImportItem) error {
	return r.db.CreateInBatches(items, 100).Error
}
