package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// ==================== 考试大类 ====================

type ExamCategoryRepo struct {
	db *gorm.DB
}

func NewExamCategoryRepo(db *gorm.DB) *ExamCategoryRepo {
	return &ExamCategoryRepo{db: db}
}

func (r *ExamCategoryRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.ExamCategory, int64, error) {
	var items []model.ExamCategory
	var total int64
	query := r.db.Model(&model.ExamCategory{})

	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name LIKE ?", "%"+escapeLike(name.(string))+"%")
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
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

func (r *ExamCategoryRepo) GetAll() ([]model.ExamCategory, error) {
	var items []model.ExamCategory
	if err := r.db.Where("status = 1").Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ExamCategoryRepo) GetByID(id uint) (*model.ExamCategory, error) {
	var item model.ExamCategory
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ExamCategoryRepo) Create(item *model.ExamCategory) error {
	return r.db.Create(item).Error
}

func (r *ExamCategoryRepo) Update(item *model.ExamCategory) error {
	return r.db.Save(item).Error
}

func (r *ExamCategoryRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.ExamCategory{}).Error
}

func (r *ExamCategoryRepo) HasChildren(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.ExamCategory{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ExamCategoryRepo) HasExams(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Exam{}).Where("exam_category_id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== 具体考试 ====================

type ExamRepo struct {
	db *gorm.DB
}

func NewExamRepo(db *gorm.DB) *ExamRepo {
	return &ExamRepo{db: db}
}

func (r *ExamRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.Exam, int64, error) {
	var items []model.Exam
	var total int64
	query := r.db.Model(&model.Exam{})

	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name LIKE ?", "%"+escapeLike(name.(string))+"%")
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if examCategoryId, ok := filters["examCategoryId"]; ok && examCategoryId != "" {
		query = query.Where("exam_category_id = ?", examCategoryId)
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

func (r *ExamRepo) GetAll() ([]model.Exam, error) {
	var items []model.Exam
	if err := r.db.Where("status = 1").Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ExamRepo) GetByCategoryID(categoryId uint) ([]model.Exam, error) {
	var items []model.Exam
	if err := r.db.Where("exam_category_id = ? AND status = 1", categoryId).Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ExamRepo) GetByID(id uint) (*model.Exam, error) {
	var item model.Exam
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ExamRepo) Create(item *model.Exam) error {
	return r.db.Create(item).Error
}

func (r *ExamRepo) Update(item *model.Exam) error {
	return r.db.Save(item).Error
}

func (r *ExamRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Exam{}).Error
}

func (r *ExamRepo) HasSubjects(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Subject{}).Where("exam_id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== 科目 ====================

type SubjectRepo struct {
	db *gorm.DB
}

func NewSubjectRepo(db *gorm.DB) *SubjectRepo {
	return &SubjectRepo{db: db}
}

func (r *SubjectRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.Subject, int64, error) {
	var items []model.Subject
	var total int64
	query := r.db.Model(&model.Subject{})

	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name LIKE ?", "%"+escapeLike(name.(string))+"%")
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if examId, ok := filters["examId"]; ok && examId != "" {
		query = query.Where("exam_id = ?", examId)
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

func (r *SubjectRepo) GetByExamID(examId uint) ([]model.Subject, error) {
	var items []model.Subject
	if err := r.db.Where("exam_id = ? AND status = 1", examId).Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SubjectRepo) GetByID(id uint) (*model.Subject, error) {
	var item model.Subject
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SubjectRepo) Create(item *model.Subject) error {
	return r.db.Create(item).Error
}

func (r *SubjectRepo) Update(item *model.Subject) error {
	return r.db.Save(item).Error
}

func (r *SubjectRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Subject{}).Error
}

func (r *SubjectRepo) HasCategories(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Category{}).Where("subject_id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== 章节分类 ====================

type QuestionCategoryRepo struct {
	db *gorm.DB
}

func NewQuestionCategoryRepo(db *gorm.DB) *QuestionCategoryRepo {
	return &QuestionCategoryRepo{db: db}
}

func (r *QuestionCategoryRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.Category, int64, error) {
	var items []model.Category
	var total int64
	query := r.db.Model(&model.Category{})

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

func (r *QuestionCategoryRepo) GetAll() ([]model.Category, error) {
	var items []model.Category
	if err := r.db.Where("status = 1").Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *QuestionCategoryRepo) GetByID(id uint) (*model.Category, error) {
	var item model.Category
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *QuestionCategoryRepo) Create(item *model.Category) error {
	return r.db.Create(item).Error
}

func (r *QuestionCategoryRepo) Update(item *model.Category) error {
	return r.db.Save(item).Error
}

func (r *QuestionCategoryRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Category{}).Error
}

func (r *QuestionCategoryRepo) HasChildren(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Category{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
