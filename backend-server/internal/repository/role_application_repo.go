package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// RoleApplicationRepo 角色申请仓库
type RoleApplicationRepo struct {
	db *gorm.DB
}

// NewRoleApplicationRepo 创建角色申请仓库
func NewRoleApplicationRepo(db *gorm.DB) *RoleApplicationRepo {
	return &RoleApplicationRepo{db: db}
}

// Create 创建角色申请
func (r *RoleApplicationRepo) Create(app *model.RoleApplication) error {
	return r.db.Create(app).Error
}

// HasPending 检查用户是否已有指定角色的待审申请
func (r *RoleApplicationRepo) HasPending(userID, roleID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.RoleApplication{}).
		Where("user_id = ? AND role_id = ? AND status = ?", userID, roleID, 0).
		Count(&count).Error
	return count > 0, err
}

// GetPendingRoleIDs 获取用户待审申请中的角色 ID
func (r *RoleApplicationRepo) GetPendingRoleIDs(userID uint) ([]uint, error) {
	var roleIDs []uint
	err := r.db.Model(&model.RoleApplication{}).
		Where("user_id = ? AND status = ?", userID, 0).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

// Update 更新角色申请
func (r *RoleApplicationRepo) Update(app *model.RoleApplication) error {
	return r.db.Save(app).Error
}

// GetByID 根据ID获取角色申请
func (r *RoleApplicationRepo) GetByID(id uint) (*model.RoleApplication, error) {
	var app model.RoleApplication
	if err := r.db.Where("id = ?", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// ListByUser 按用户查询角色申请（分页）
func (r *RoleApplicationRepo) ListByUser(userID uint, page, pageSize int) ([]model.RoleApplication, int64, error) {
	var list []model.RoleApplication
	var total int64

	query := r.db.Model(&model.RoleApplication{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ListAll 查询所有角色申请（管理员，分页）
func (r *RoleApplicationRepo) ListAll(page, pageSize int, filters map[string]interface{}) ([]model.RoleApplication, int64, error) {
	var list []model.RoleApplication
	var total int64

	query := r.db.Model(&model.RoleApplication{})

	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if userID, ok := filters["userId"]; ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if roleID, ok := filters["roleId"]; ok && roleID != "" {
		query = query.Where("role_id = ?", roleID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
