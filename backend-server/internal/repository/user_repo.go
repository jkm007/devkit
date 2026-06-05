package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// UserRepo 用户仓库
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓库
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// List 获取用户列表（分页）
func (r *UserRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{})

	// 应用筛选条件
	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name LIKE ?", "%"+escapeLike(name.(string))+"%")
	}
	if id, ok := filters["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if groupID, ok := filters["groupId"]; ok && groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}
	if startTime, ok := filters["startTime"]; ok && startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime, ok := filters["endTime"]; ok && endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}
	if remark, ok := filters["remark"]; ok && remark != "" {
		query = query.Where("remark LIKE ?", "%"+escapeLike(remark.(string))+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetByID 根据 ID 获取用户
func (r *UserRepo) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByName 根据用户名获取用户
func (r *UserRepo) GetByName(name string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("name = ?", name).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update 更新用户
func (r *UserRepo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete 删除用户（软删除）
func (r *UserRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.User{}).Error
}

// GetUserRoles 获取用户的角色列表
func (r *UserRepo) GetUserRoles(userID uint) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Table("sys_roles").
		Joins("JOIN sys_user_roles ON sys_user_roles.role_id = sys_roles.id").
		Where("sys_user_roles.user_id = ?", userID).
		Where("sys_roles.deleted_at IS NULL").
		Find(&roles).Error
	return roles, err
}

// GetUserRoleIDs 获取用户的角色 ID 列表
func (r *UserRepo) GetUserRoleIDs(userID uint) ([]uint, error) {
	var roleIDs []uint
	err := r.db.Model(&model.UserRole{}).Where("user_id = ?", userID).Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

// SyncUserRoles 同步用户角色（替换所有）
func (r *UserRepo) SyncUserRoles(userID uint, roleIDs []uint) error {
	// 删除旧的关联
	if err := r.db.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return err
	}

	// 创建新的关联
	for _, roleID := range roleIDs {
		userRole := &model.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
		if err := r.db.Create(userRole).Error; err != nil {
			return err
		}
	}

	return nil
}

// AddUserRole 添加单个用户角色（不删除已有角色）
func (r *UserRepo) AddUserRole(userID, roleID uint) error {
	// 检查是否已存在
	var count int64
	err := r.db.Model(&model.UserRole{}).Where("user_id = ? AND role_id = ?", userID, roleID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // 已存在，跳过
	}

	return r.db.Create(&model.UserRole{UserID: userID, RoleID: roleID}).Error
}

// GetUserIDsByRoleID 获取拥有指定角色的所有用户 ID
func (r *UserRepo) GetUserIDsByRoleID(roleID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Model(&model.UserRole{}).Where("role_id = ?", roleID).Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// GetUserIDsByGroupID 获取指定分组的所有用户 ID
func (r *UserRepo) GetUserIDsByGroupID(groupID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Model(&model.User{}).Where("group_id = ?", groupID).Pluck("id", &userIDs).Error
	return userIDs, err
}
