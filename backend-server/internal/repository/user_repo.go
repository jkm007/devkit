package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	// 应用筛选条件（使用安全类型断言）
	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("name LIKE ?", "%"+escapeLike(name)+"%")
	}
	if id, ok := filters["id"].(string); ok && id != "" {
		query = query.Where("id = ?", id)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if groupID, ok := filters["groupId"].(string); ok && groupID != "" && groupID != "-1" {
		query = query.Where("group_id = ?", groupID)
	}
	if startTime, ok := filters["startTime"].(string); ok && startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime, ok := filters["endTime"].(string); ok && endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}
	if remark, ok := filters["remark"].(string); ok && remark != "" {
		query = query.Where("remark LIKE ?", "%"+escapeLike(remark)+"%")
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

// GetByIDs 批量获取用户（N+1 优化）
func (r *UserRepo) GetByIDs(ids []uint) (map[uint]*model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []model.User
	if err := r.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	result := make(map[uint]*model.User, len(users))
	for i := range users {
		result[users[i].ID] = &users[i]
	}
	return result, nil
}

// GetByName 根据用户名获取用户
func (r *UserRepo) GetByName(name string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("name = ?", name).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户（别名）
func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	return r.GetByName(username)
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepo) GetByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByPhone 根据手机号获取用户
func (r *UserRepo) GetByPhone(phone string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("phone = ?", phone).First(&user).Error; err != nil {
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

// DeleteWithCleanup 删除用户并清理关联数据
func (r *UserRepo) DeleteWithCleanup(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 软删除用户
		if err := tx.Where("id = ?", id).Delete(&model.User{}).Error; err != nil {
			return err
		}
		// 清理角色关联
		if err := tx.Where("user_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		// 清理隐私设置
		if err := tx.Where("user_id = ?", id).Delete(&model.UserPrivacy{}).Error; err != nil {
			return err
		}
		// 清理实名认证
		if err := tx.Where("user_id = ?", id).Delete(&model.UserRealName{}).Error; err != nil {
			return err
		}
		// 清理登录设备
		if err := tx.Where("user_id = ?", id).Delete(&model.LoginDevice{}).Error; err != nil {
			return err
		}
		// 清理 OAuth 绑定
		if err := tx.Where("user_id = ?", id).Delete(&model.OAuthUser{}).Error; err != nil {
			return err
		}
		// 清理密码历史
		if err := tx.Where("user_id = ?", id).Delete(&model.PasswordHistory{}).Error; err != nil {
			return err
		}
		// 清理角色申请
		if err := tx.Where("user_id = ?", id).Delete(&model.RoleApplication{}).Error; err != nil {
			return err
		}
		// 清理文件分享
		if err := tx.Where("user_id = ?", id).Delete(&model.FileShare{}).Error; err != nil {
			return err
		}
		// 递减文件资产引用计数（先查询再删除）
		var fileEntries []model.FileEntry
		if err := tx.Where("user_id = ?", id).Find(&fileEntries).Error; err == nil {
			for _, entry := range fileEntries {
				if entry.FileAssetID > 0 {
					tx.Model(&model.FileAsset{}).Where("id = ?", entry.FileAssetID).Update("ref_count", gorm.Expr("GREATEST(ref_count - 1, 0)"))
				}
			}
		}
		// 清理文件条目
		if err := tx.Where("user_id = ?", id).Delete(&model.FileEntry{}).Error; err != nil {
			return err
		}
		// 清理文件夹
		if err := tx.Where("user_id = ?", id).Delete(&model.FileFolder{}).Error; err != nil {
			return err
		}
		return nil
	})
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

// GetUserRoleIDsByUserIDs 批量获取多个用户的角色 ID 列表（N+1 优化）
func (r *UserRepo) GetUserRoleIDsByUserIDs(userIDs []uint) (map[uint][]uint, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	var userRoles []model.UserRole
	if err := r.db.Where("user_id IN ?", userIDs).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	result := make(map[uint][]uint, len(userIDs))
	for _, ur := range userRoles {
		result[ur.UserID] = append(result[ur.UserID], ur.RoleID)
	}
	return result, nil
}

// SyncUserRoles 同步用户角色（替换所有，事务保护）
func (r *UserRepo) SyncUserRoles(userID uint, roleIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧的关联
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}

		// 创建新的关联
		for _, roleID := range roleIDs {
			userRole := &model.UserRole{
				UserID: userID,
				RoleID: roleID,
			}
			if err := tx.Create(userRole).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// AddUserRole 添加单个用户角色（不删除已有角色，使用 ON CONFLICT 避免竞态）
func (r *UserRepo) AddUserRole(userID, roleID uint) error {
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.UserRole{UserID: userID, RoleID: roleID}).Error
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
