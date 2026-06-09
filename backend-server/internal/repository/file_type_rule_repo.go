package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// FileTypeRuleRepo 文件类型规则仓库
type FileTypeRuleRepo struct {
	db *gorm.DB
}

// NewFileTypeRuleRepo 创建文件类型规则仓库
func NewFileTypeRuleRepo(db *gorm.DB) *FileTypeRuleRepo {
	return &FileTypeRuleRepo{db: db}
}

// GetAll 获取所有规则
func (r *FileTypeRuleRepo) GetAll() ([]model.FileTypeRule, error) {
	var rules []model.FileTypeRule
	if err := r.db.Order("file_type ASC, extension ASC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// GetAllEnabled 获取所有启用的规则
func (r *FileTypeRuleRepo) GetAllEnabled() ([]model.FileTypeRule, error) {
	var rules []model.FileTypeRule
	if err := r.db.Where("status = ?", 1).Order("file_type ASC, extension ASC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// GetByID 根据 ID 获取规则
func (r *FileTypeRuleRepo) GetByID(id int64) (*model.FileTypeRule, error) {
	var rule model.FileTypeRule
	if err := r.db.Where("id = ?", id).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetByExtension 根据扩展名获取规则
func (r *FileTypeRuleRepo) GetByExtension(ext string) (*model.FileTypeRule, error) {
	var rule model.FileTypeRule
	if err := r.db.Where("extension = ?", ext).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// Create 创建规则
func (r *FileTypeRuleRepo) Create(rule *model.FileTypeRule) error {
	return r.db.Create(rule).Error
}

// Update 更新规则
func (r *FileTypeRuleRepo) Update(rule *model.FileTypeRule) error {
	return r.db.Save(rule).Error
}

// Delete 删除规则
func (r *FileTypeRuleRepo) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.FileTypeRule{}).Error
}
