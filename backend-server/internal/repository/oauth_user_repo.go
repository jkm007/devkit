package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// OAuthUserRepo 第三方登录绑定仓库
type OAuthUserRepo struct {
	db *gorm.DB
}

// NewOAuthUserRepo 创建第三方登录绑定仓库
func NewOAuthUserRepo(db *gorm.DB) *OAuthUserRepo {
	return &OAuthUserRepo{db: db}
}

// Create 创建绑定
func (r *OAuthUserRepo) Create(oauth *model.OAuthUser) error {
	return r.db.Create(oauth).Error
}

// Update 更新绑定
func (r *OAuthUserRepo) Update(oauth *model.OAuthUser) error {
	return r.db.Save(oauth).Error
}

// ListByUser 获取用户的第三方绑定列表
func (r *OAuthUserRepo) ListByUser(userID uint) ([]model.OAuthUser, error) {
	var bindings []model.OAuthUser
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&bindings).Error
	return bindings, err
}

// GetByUserAndProvider 根据用户ID和提供商获取绑定
func (r *OAuthUserRepo) GetByUserAndProvider(userID uint, provider string) (*model.OAuthUser, error) {
	var binding model.OAuthUser
	if err := r.db.Where("user_id = ? AND provider = ?", userID, provider).First(&binding).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

// GetByProvider 根据提供商和第三方用户ID获取绑定
func (r *OAuthUserRepo) GetByProvider(provider, providerUserID string) (*model.OAuthUser, error) {
	var binding model.OAuthUser
	if err := r.db.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&binding).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

// Delete 删除绑定
func (r *OAuthUserRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.OAuthUser{}).Error
}

// DeleteByUserAndProvider 删除指定用户的指定提供商绑定
func (r *OAuthUserRepo) DeleteByUserAndProvider(userID uint, provider string) error {
	return r.db.Where("user_id = ? AND provider = ?", userID, provider).Delete(&model.OAuthUser{}).Error
}

// CountByUser 统计用户的绑定数量
func (r *OAuthUserRepo) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.OAuthUser{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetByProviderUserID 根据第三方用户ID查找绑定（跨所有 provider，用于 UnionID 去重）
func (r *OAuthUserRepo) GetByProviderUserID(providerUserID string) (*model.OAuthUser, error) {
	var binding model.OAuthUser
	if err := r.db.Where("provider_user_id = ?", providerUserID).First(&binding).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}
