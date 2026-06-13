package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// FavoriteNoteRepo 收藏和笔记数据访问
type FavoriteNoteRepo struct {
	db *gorm.DB
}

// NewFavoriteNoteRepo 创建收藏和笔记数据仓库
func NewFavoriteNoteRepo(db *gorm.DB) *FavoriteNoteRepo {
	return &FavoriteNoteRepo{db: db}
}

// IsFavorited 检查用户是否已收藏题目
func (r *FavoriteNoteRepo) IsFavorited(userID, questionID uint) bool {
	var count int64
	r.db.Model(&model.UserFavorite{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Count(&count)
	return count > 0
}

// AddFavorite 添加收藏
func (r *FavoriteNoteRepo) AddFavorite(userID, questionID uint) error {
	favorite := &model.UserFavorite{
		UserID:     userID,
		QuestionID: questionID,
	}
	return r.db.Create(favorite).Error
}

// RemoveFavorite 取消收藏
func (r *FavoriteNoteRepo) RemoveFavorite(userID, questionID uint) error {
	return r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&model.UserFavorite{}).Error
}

// ListFavorites 获取用户收藏列表
func (r *FavoriteNoteRepo) ListFavorites(userID uint, offset, limit int) ([]map[string]interface{}, int64, error) {
	var results []map[string]interface{}
	var total int64

	query := r.db.Table("user_favorites f").
		Select("f.id, f.question_id as questionId, q.title, q.question_type as questionType, "+
			"q.difficulty, kc.name as categoryName, f.created_at as createdAt").
		Joins("INNER JOIN qb_questions q ON f.question_id = q.id").
		Joins("LEFT JOIN qb_question_categories kc ON q.category_id = kc.id").
		Where("f.user_id = ?", userID)

	r.db.Table("user_favorites").Where("user_id = ?", userID).Count(&total)

	query.Order("f.created_at DESC").Offset(offset).Limit(limit).Scan(&results)

	return results, total, nil
}

// CreateNote 创建笔记
func (r *FavoriteNoteRepo) CreateNote(note *model.UserNote) error {
	return r.db.Create(note).Error
}

// UpdateNote 更新笔记
func (r *FavoriteNoteRepo) UpdateNote(note *model.UserNote) error {
	return r.db.Model(&model.UserNote{}).
		Where("id = ? AND user_id = ?", note.ID, note.UserID).
		Updates(map[string]interface{}{
			"content":    note.Content,
			"updated_at": note.UpdatedAt,
		}).Error
}

// DeleteNote 删除笔记
func (r *FavoriteNoteRepo) DeleteNote(noteID, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", noteID, userID).
		Delete(&model.UserNote{}).Error
}

// GetNoteByID 获取笔记详情
func (r *FavoriteNoteRepo) GetNoteByID(noteID, userID uint) (*model.UserNote, error) {
	var note model.UserNote
	err := r.db.Where("id = ? AND user_id = ?", noteID, userID).
		First(&note).Error
	return &note, err
}

// ListNotes 获取用户笔记列表
func (r *FavoriteNoteRepo) ListNotes(userID uint, offset, limit int) ([]map[string]interface{}, int64, error) {
	var results []map[string]interface{}
	var total int64

	query := r.db.Table("user_notes n").
		Select("n.id, n.question_id as questionId, q.title as questionTitle, "+
			"n.content, n.created_at as createdAt, n.updated_at as updatedAt").
		Joins("LEFT JOIN qb_questions q ON n.question_id = q.id").
		Where("n.user_id = ?", userID)

	r.db.Table("user_notes").Where("user_id = ?", userID).Count(&total)

	query.Order("n.updated_at DESC").Offset(offset).Limit(limit).Scan(&results)

	return results, total, nil
}

// GetNoteByQuestionID 获取某题目的笔记
func (r *FavoriteNoteRepo) GetNoteByQuestionID(questionID, userID uint) (*model.UserNote, error) {
	var note model.UserNote
	err := r.db.Where("question_id = ? AND user_id = ?", questionID, userID).
		First(&note).Error
	return &note, err
}
