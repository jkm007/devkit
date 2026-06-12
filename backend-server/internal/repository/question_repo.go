package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

type QuestionRepo struct {
	db *gorm.DB
}

func NewQuestionRepo(db *gorm.DB) *QuestionRepo {
	return &QuestionRepo{db: db}
}

func (r *QuestionRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.Question, int64, error) {
	var items []model.Question
	var total int64
	query := r.db.Model(&model.Question{})

	if title, ok := filters["title"]; ok && title != "" {
		query = query.Where("title LIKE ?", "%"+escapeLike(title.(string))+"%")
	}
	if questionType, ok := filters["questionType"]; ok && questionType != "" {
		query = query.Where("question_type = ?", questionType)
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
	if difficulty, ok := filters["difficulty"]; ok && difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if resourceType, ok := filters["resourceType"]; ok && resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
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

func (r *QuestionRepo) GetByID(id uint) (*model.Question, error) {
	var item model.Question
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *QuestionRepo) Create(item *model.Question) error {
	return r.db.Create(item).Error
}

func (r *QuestionRepo) Update(item *model.Question) error {
	return r.db.Save(item).Error
}

func (r *QuestionRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Question{}).Error
}

func (r *QuestionRepo) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总题数
	var total int64
	if err := r.db.Model(&model.Question{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 按状态统计
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusCounts []StatusCount
	if err := r.db.Model(&model.Question{}).Where("deleted_at IS NULL").
		Select("status, count(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		return nil, err
	}
	stats["byStatus"] = statusCounts

	// 按题型统计
	type TypeCount struct {
		QuestionType string `json:"questionType"`
		Count        int64  `json:"count"`
	}
	var typeCounts []TypeCount
	if err := r.db.Model(&model.Question{}).Where("deleted_at IS NULL").
		Select("question_type, count(*) as count").Group("question_type").Scan(&typeCounts).Error; err != nil {
		return nil, err
	}
	stats["byType"] = typeCounts

	// 按难度统计
	type DifficultyCount struct {
		Difficulty int   `json:"difficulty"`
		Count      int64 `json:"count"`
	}
	var diffCounts []DifficultyCount
	if err := r.db.Model(&model.Question{}).Where("deleted_at IS NULL").
		Select("difficulty, count(*) as count").Group("difficulty").Scan(&diffCounts).Error; err != nil {
		return nil, err
	}
	stats["byDifficulty"] = diffCounts

	return stats, nil
}
