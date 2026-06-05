package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// PasswordHistoryRepo 密码历史仓库
type PasswordHistoryRepo struct {
	db *gorm.DB
}

// NewPasswordHistoryRepo 创建密码历史仓库
func NewPasswordHistoryRepo(db *gorm.DB) *PasswordHistoryRepo {
	return &PasswordHistoryRepo{db: db}
}

// Create 创建密码历史记录
func (r *PasswordHistoryRepo) Create(history *model.PasswordHistory) error {
	return r.db.Create(history).Error
}

// GetRecent 获取用户最近 N 条密码历史
func (r *PasswordHistoryRepo) GetRecent(userID uint, limit int) ([]model.PasswordHistory, error) {
	var histories []model.PasswordHistory
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&histories).Error
	return histories, err
}

// DeleteOld 删除超过指定天数的历史记录
func (r *PasswordHistoryRepo) DeleteOld(days int) (int64, error) {
	result := r.db.Where("created_at < DATE_SUB(NOW(), INTERVAL ? DAY)", days).Delete(&model.PasswordHistory{})
	return result.RowsAffected, result.Error
}
