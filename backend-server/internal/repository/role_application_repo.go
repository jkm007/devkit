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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
