package repository

import (
	"backend-server/internal/model"
	"gorm.io/gorm"
)

// TagRoutingRepo 标签路由规则仓库
type TagRoutingRepo struct {
	db *gorm.DB
}

// NewTagRoutingRepo 创建标签路由规则仓库
func NewTagRoutingRepo(db *gorm.DB) *TagRoutingRepo {
	return &TagRoutingRepo{db: db}
}

// GetAll 获取所有路由规则（按优先级降序）
func (r *TagRoutingRepo) GetAll() ([]model.TagRouting, error) {
	var rules []model.TagRouting
	err := r.db.Order("priority DESC, id ASC").Find(&rules).Error
	return rules, err
}

// GetEnabled 获取启用的路由规则（按优先级降序）
func (r *TagRoutingRepo) GetEnabled() ([]model.TagRouting, error) {
	var rules []model.TagRouting
	err := r.db.Where("status = ?", 1).Order("priority DESC, id ASC").Find(&rules).Error
	return rules, err
}

// GetByID 根据ID获取规则
func (r *TagRoutingRepo) GetByID(id int64) (*model.TagRouting, error) {
	var rule model.TagRouting
	err := r.db.Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetDefault 获取默认规则
func (r *TagRoutingRepo) GetDefault() (*model.TagRouting, error) {
	var rule model.TagRouting
	err := r.db.Where("is_default = ? AND status = ?", true, 1).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// Create 创建规则
func (r *TagRoutingRepo) Create(rule *model.TagRouting) error {
	return r.db.Create(rule).Error
}

// Update 更新规则
func (r *TagRoutingRepo) Update(rule *model.TagRouting) error {
	return r.db.Save(rule).Error
}

// Delete 删除规则
func (r *TagRoutingRepo) Delete(id int64) error {
	return r.db.Delete(&model.TagRouting{}, id).Error
}

// UpdateStatus 更新规则状态
func (r *TagRoutingRepo) UpdateStatus(id int64, status int8) error {
	return r.db.Model(&model.TagRouting{}).Where("id = ?", id).Update("status", status).Error
}

// UpdatePriority 更新规则优先级
func (r *TagRoutingRepo) UpdatePriority(id int64, priority int) error {
	return r.db.Model(&model.TagRouting{}).Where("id = ?", id).Update("priority", priority).Error
}

// BatchUpdatePriority 批量更新优先级
func (r *TagRoutingRepo) BatchUpdatePriority(priorities map[int64]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for id, priority := range priorities {
			if err := tx.Model(&model.TagRouting{}).Where("id = ?", id).Update("priority", priority).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
