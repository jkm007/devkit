package repository

import (
	"backend-server/internal/model"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

type RateLimitRuleRepo struct {
	db *gorm.DB
}

func NewRateLimitRuleRepo() *RateLimitRuleRepo {
	return &RateLimitRuleRepo{db: database.GetMySQL()}
}

// GetAll 获取所有限流规则
func (r *RateLimitRuleRepo) GetAll() ([]model.RateLimitRule, error) {
	var rules []model.RateLimitRule
	err := r.db.Order("priority DESC, id ASC").Find(&rules).Error
	return rules, err
}

// GetEnabled 获取所有启用的限流规则
func (r *RateLimitRuleRepo) GetEnabled() ([]model.RateLimitRule, error) {
	var rules []model.RateLimitRule
	err := r.db.Where("enabled = ?", true).Order("priority DESC, id ASC").Find(&rules).Error
	return rules, err
}

// GetByID 根据 ID 获取
func (r *RateLimitRuleRepo) GetByID(id uint) (*model.RateLimitRule, error) {
	var rule model.RateLimitRule
	err := r.db.First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// Create 创建
func (r *RateLimitRuleRepo) Create(rule *model.RateLimitRule) error {
	return r.db.Create(rule).Error
}

// Update 更新
func (r *RateLimitRuleRepo) Update(rule *model.RateLimitRule) error {
	return r.db.Save(rule).Error
}

// Delete 删除
func (r *RateLimitRuleRepo) Delete(id uint) error {
	return r.db.Delete(&model.RateLimitRule{}, id).Error
}

// UpdateEnabled 更新启用状态
func (r *RateLimitRuleRepo) UpdateEnabled(id uint, enabled bool) error {
	return r.db.Model(&model.RateLimitRule{}).Where("id = ?", id).Update("enabled", enabled).Error
}
