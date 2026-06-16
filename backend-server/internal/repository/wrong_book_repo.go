package repository

import (
	"backend-server/internal/model"
	"time"

	"gorm.io/gorm"
)

// WrongBookRepo 错题本数据访问
type WrongBookRepo struct {
	db *gorm.DB
}

// NewWrongBookRepo 创建错题本仓库
func NewWrongBookRepo(db *gorm.DB) *WrongBookRepo {
	return &WrongBookRepo{db: db}
}

// AddOrUpdate 添加或更新错题（重复错误累加次数）
func (r *WrongBookRepo) AddOrUpdate(userID, questionID uint, categoryID uint) error {
	var wb model.WrongBook
	err := r.db.Where("user_id = ? AND question_id = ? AND is_mastered = 0", userID, questionID).
		First(&wb).Error

	if err == gorm.ErrRecordNotFound {
		// 新错题
		return r.db.Create(&model.WrongBook{
			UserID:      userID,
			QuestionID:  questionID,
			CategoryID:  categoryID,
			WrongCount:  1,
			LastWrongAt: time.Now(),
		}).Error
	} else if err != nil {
		return err
	}

	// 已有错题：累加次数
	return r.db.Model(&wb).Updates(map[string]interface{}{
		"wrong_count":   wb.WrongCount + 1,
		"last_wrong_at": time.Now(),
		"is_mastered":   false,
	}).Error
}

// List 获取错题列表
func (r *WrongBookRepo) List(userID uint, categoryID uint, isMastered *bool, offset, limit int) ([]map[string]interface{}, int64, error) {
	var results []map[string]interface{}
	var total int64

	query := r.db.Table("wrong_books wb").
		Select("wb.id, wb.question_id as questionId, wb.category_id as categoryId, "+
			"wb.wrong_count as wrongCount, wb.last_wrong_at as lastWrongAt, "+
			"wb.is_mastered as isMastered, wb.mastered_at as masteredAt, "+
			"q.title, q.question_type as questionType, q.difficulty, "+
			"kc.name as categoryName").
		Joins("INNER JOIN qb_questions q ON wb.question_id = q.id").
		Joins("LEFT JOIN qb_categories kc ON q.category_id = kc.id").
		Where("wb.user_id = ?", userID)

	if categoryID > 0 {
		query = query.Where("wb.category_id = ?", categoryID)
	}
	if isMastered != nil {
		query = query.Where("wb.is_mastered = ?", *isMastered)
	} else {
		query = query.Where("wb.is_mastered = 0") // 默认只显示未掌握
	}

	countQuery := query
	countQuery.Count(&total)

	query.Order("wb.last_wrong_at DESC").Offset(offset).Limit(limit).Scan(&results)

	return results, total, nil
}

// GetByQuestionID 获取某道题的错题记录
func (r *WrongBookRepo) GetByQuestionID(userID, questionID uint) (*model.WrongBook, error) {
	var wb model.WrongBook
	err := r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		First(&wb).Error
	return &wb, err
}

// MarkMastered 标记已掌握
func (r *WrongBookRepo) MarkMastered(userID, questionID uint) error {
	now := time.Now()
	return r.db.Model(&model.WrongBook{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Updates(map[string]interface{}{
			"is_mastered": true,
			"mastered_at": now,
		}).Error
}

// BatchMarkMastered 批量标记已掌握
func (r *WrongBookRepo) BatchMarkMastered(userID uint, questionIDs []uint) error {
	now := time.Now()
	return r.db.Model(&model.WrongBook{}).
		Where("user_id = ? AND question_id IN ?", userID, questionIDs).
		Updates(map[string]interface{}{
			"is_mastered": true,
			"mastered_at": now,
		}).Error
}

// Delete 移除错题
func (r *WrongBookRepo) Delete(userID, questionID uint) error {
	return r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&model.WrongBook{}).Error
}

// GetRandomQuestions 获取随机错题（用于重做）
func (r *WrongBookRepo) GetRandomQuestions(userID uint, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	r.db.Table("wrong_books wb").
		Select("wb.question_id as question_id, q.title, q.question_type, q.difficulty, q.stem, q.content, q.answer, q.analysis, q.category_id").
		Joins("INNER JOIN qb_questions q ON wb.question_id = q.id").
		Where("wb.user_id = ? AND wb.is_mastered = 0", userID).
		Order("RAND()").Limit(limit).Scan(&results)

	return results, nil
}

// GetStats 获取错题统计
func (r *WrongBookRepo) GetStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总错题数
	var total int64
	r.db.Model(&model.WrongBook{}).Where("user_id = ? AND is_mastered = 0", userID).Count(&total)
	stats["total"] = total

	// 已掌握
	var mastered int64
	r.db.Model(&model.WrongBook{}).Where("user_id = ? AND is_mastered = 1", userID).Count(&mastered)
	stats["mastered"] = mastered

	// 本周新增
	var thisWeek int64
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	r.db.Model(&model.WrongBook{}).
		Where("user_id = ? AND created_at >= ?", userID, weekStart).
		Count(&thisWeek)
	stats["thisWeek"] = thisWeek

	// 分类分布
	var categoryDist []struct {
		CategoryName string `json:"categoryName"`
		CategoryID   uint   `json:"categoryId"`
		Count        int64  `json:"count"`
	}
	r.db.Table("wrong_books wb").
		Select("kc.name as category_name, wb.category_id, count(*) as count").
		Joins("LEFT JOIN qb_question_categories kc ON wb.category_id = kc.id").
		Where("wb.user_id = ? AND wb.is_mastered = 0", userID).
		Group("wb.category_id, kc.name").
		Order("count DESC").
		Scan(&categoryDist)
	stats["categoryDist"] = categoryDist

	return stats, nil
}
