package repository

import (
	"errors"

	"backend-server/internal/model"

	"gorm.io/gorm"
)

// CategoryBindingRepo 分类绑定数据访问
type CategoryBindingRepo struct {
	db *gorm.DB
}

// NewCategoryBindingRepo 创建分类绑定仓库
func NewCategoryBindingRepo(db *gorm.DB) *CategoryBindingRepo {
	return &CategoryBindingRepo{db: db}
}

// GetDB 获取数据库连接
func (r *CategoryBindingRepo) GetDB() *gorm.DB {
	return r.db
}

// List 获取用户绑定的分类
func (r *CategoryBindingRepo) List(userID uint) ([]model.UserCategoryBinding, error) {
	var bindings []model.UserCategoryBinding
	err := r.db.Where("user_id = ?", userID).Order("is_primary DESC, bound_at ASC").Find(&bindings).Error
	return bindings, err
}

// Count 获取绑定数量
func (r *CategoryBindingRepo) Count(userID uint) (int64, error) {
	var count int64
	r.db.Model(&model.UserCategoryBinding{}).Where("user_id = ?", userID).Count(&count)
	return count, nil
}

// IsBound 检查是否已绑定
func (r *CategoryBindingRepo) IsBound(userID, categoryID uint) bool {
	var count int64
	r.db.Model(&model.UserCategoryBinding{}).
		Where("user_id = ? AND category_id = ?", userID, categoryID).Count(&count)
	return count > 0
}

// Create 绑定分类
func (r *CategoryBindingRepo) Create(userID, categoryID uint, isPrimary bool) error {
	// 检查数量限制
	count, _ := r.Count(userID)
	if count >= 3 {
		return errors.New("最多绑定 3 个分类")
	}

	// 检查重复
	if r.IsBound(userID, categoryID) {
		return errors.New("已绑定该分类")
	}

	// 如果是主分类，先取消其他主分类
	if isPrimary {
		r.db.Model(&model.UserCategoryBinding{}).
			Where("user_id = ?", userID).Update("is_primary", false)
	}

	return r.db.Create(&model.UserCategoryBinding{
		UserID:     userID,
		CategoryID: categoryID,
		IsPrimary:  isPrimary,
	}).Error
}

// SetPrimary 设为主分类
func (r *CategoryBindingRepo) SetPrimary(userID, id uint) error {
	// 取消所有主分类
	r.db.Model(&model.UserCategoryBinding{}).
		Where("user_id = ?", userID).Update("is_primary", false)

	// 设置新的主分类
	return r.db.Model(&model.UserCategoryBinding{}).
		Where("id = ? AND user_id = ?", id, userID).Update("is_primary", true).Error
}

// Delete 解绑
func (r *CategoryBindingRepo) Delete(userID, id uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.UserCategoryBinding{}).Error
}

// GetBoundCategoryIDs 获取已绑定的分类 ID 列表
func (r *CategoryBindingRepo) GetBoundCategoryIDs(userID uint) ([]uint, error) {
	var ids []uint
	r.db.Model(&model.UserCategoryBinding{}).
		Where("user_id = ?", userID).Pluck("category_id", &ids)
	return ids, nil
}
