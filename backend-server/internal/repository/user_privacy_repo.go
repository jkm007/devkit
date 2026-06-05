package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// UserPrivacyRepo 用户隐私设置仓库
type UserPrivacyRepo struct {
	db *gorm.DB
}

// NewUserPrivacyRepo 创建用户隐私设置仓库
func NewUserPrivacyRepo(db *gorm.DB) *UserPrivacyRepo {
	return &UserPrivacyRepo{db: db}
}

// GetByUserID 根据用户ID获取隐私设置
func (r *UserPrivacyRepo) GetByUserID(userID uint) (*model.UserPrivacy, error) {
	var privacy model.UserPrivacy
	if err := r.db.Where("user_id = ?", userID).First(&privacy).Error; err != nil {
		return nil, err
	}
	return &privacy, nil
}

// Create 创建隐私设置
func (r *UserPrivacyRepo) Create(privacy *model.UserPrivacy) error {
	return r.db.Create(privacy).Error
}

// Update 更新隐私设置
func (r *UserPrivacyRepo) Update(privacy *model.UserPrivacy) error {
	return r.db.Save(privacy).Error
}
