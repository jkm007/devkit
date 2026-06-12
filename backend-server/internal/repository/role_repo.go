package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// RoleRepo 角色仓库
type RoleRepo struct {
	db *gorm.DB
}

// NewRoleRepo 创建角色仓库
func NewRoleRepo(db *gorm.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

// List 获取角色列表（分页）
func (r *RoleRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := r.db.Model(&model.Role{})

	// 应用筛选条件
	if name, ok := filters["name"]; ok && name != "" {
		if nameStr, ok := name.(string); ok {
			query = query.Where("name LIKE ?", "%"+escapeLike(nameStr)+"%")
		}
	}
	if id, ok := filters["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime, ok := filters["startTime"]; ok && startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime, ok := filters["endTime"]; ok && endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}
	if remark, ok := filters["remark"]; ok && remark != "" {
		if remarkStr, ok := remark.(string); ok {
			query = query.Where("remark LIKE ?", "%"+escapeLike(remarkStr)+"%")
		}
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// GetByID 根据 ID 获取角色
func (r *RoleRepo) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	if err := r.db.Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByName 根据名称获取角色
func (r *RoleRepo) GetByName(name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByIDs 批量根据 ID 列表获取角色（消除 N+1 查询）
func (r *RoleRepo) GetByIDs(ids []uint) ([]model.Role, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var roles []model.Role
	if err := r.db.Where("id IN ?", ids).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// ListAvailableForApply 获取可申请角色列表
func (r *RoleRepo) ListAvailableForApply(excludeIDs []uint, excludeNames []string) ([]model.Role, error) {
	var roles []model.Role
	query := r.db.Model(&model.Role{}).Where("status = ?", 1)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if len(excludeNames) > 0 {
		query = query.Where("name NOT IN ?", excludeNames)
	}
	if err := query.Order("created_at DESC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// Create 创建角色
func (r *RoleRepo) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

// Update 更新角色
func (r *RoleRepo) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

// Delete 删除角色（软删除）
func (r *RoleRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Role{}).Error
}
