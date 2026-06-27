package repository

import (
	"errors"

	"backend-server/internal/model"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// CategoryFavoriteRepo 分类收藏数据访问
type CategoryFavoriteRepo struct {
	db *gorm.DB
}

// NewCategoryFavoriteRepo 创建分类收藏仓库
func NewCategoryFavoriteRepo(db *gorm.DB) *CategoryFavoriteRepo {
	return &CategoryFavoriteRepo{db: db}
}

// List 获取用户的分类收藏列表
func (r *CategoryFavoriteRepo) List(userID uint, offset, limit int) ([]model.UserCategoryFavorite, int64, error) {
	var items []model.UserCategoryFavorite
	var total int64

	err := r.db.Model(&model.UserCategoryFavorite{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&items).Error

	return items, total, err
}

// GetByID 根据ID获取收藏记录
func (r *CategoryFavoriteRepo) GetByID(id uint) (*model.UserCategoryFavorite, error) {
	var item model.UserCategoryFavorite
	err := r.db.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Exists 检查是否已收藏
func (r *CategoryFavoriteRepo) Exists(userID uint, targetType string, targetID uint) bool {
	var count int64
	r.db.Model(&model.UserCategoryFavorite{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Count(&count)
	return count > 0
}

// Create 创建收藏记录
func (r *CategoryFavoriteRepo) Create(item *model.UserCategoryFavorite) error {
	return r.db.Create(item).Error
}

// Delete 删除收藏记录（已校验 user_id 归属）
func (r *CategoryFavoriteRepo) Delete(userID, id uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.UserCategoryFavorite{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("收藏记录不存在")
	}
	return nil
}

// IsDuplicateError 判断是否为唯一索引冲突
func (r *CategoryFavoriteRepo) IsDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	return false
}
