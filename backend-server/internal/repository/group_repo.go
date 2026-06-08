package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// GroupRepo 分组仓库
type GroupRepo struct {
	db *gorm.DB
}

// NewGroupRepo 创建分组仓库
func NewGroupRepo(db *gorm.DB) *GroupRepo {
	return &GroupRepo{db: db}
}

// List 获取分组列表
func (r *GroupRepo) List() ([]model.Group, error) {
	var groups []model.Group
	if err := r.db.Order("created_at ASC").Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// GetByID 根据 ID 获取分组
func (r *GroupRepo) GetByID(id uint) (*model.Group, error) {
	var group model.Group
	if err := r.db.Where("id = ?", id).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// Create 创建分组
func (r *GroupRepo) Create(group *model.Group) error {
	return r.db.Create(group).Error
}

// Update 更新分组
func (r *GroupRepo) Update(group *model.Group) error {
	return r.db.Save(group).Error
}

// Delete 删除分组（软删除）
func (r *GroupRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Group{}).Error
}

// GetChildren 获取子分组
func (r *GroupRepo) GetChildren(pid uint) ([]model.Group, error) {
	var groups []model.Group
	if err := r.db.Where("pid = ?", pid).Order("created_at ASC").Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// DeleteWithChildren 删除分组及其子分组（事务保护）
func (r *GroupRepo) DeleteWithChildren(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return r.deleteWithChildrenTx(tx, id)
	})
}

// deleteWithChildrenTx 递归删除分组（在事务内）
func (r *GroupRepo) deleteWithChildrenTx(tx *gorm.DB, id uint) error {
	// 递归删除子分组
	var children []model.Group
	if err := tx.Where("pid = ?", id).Find(&children).Error; err != nil {
		return err
	}
	for _, child := range children {
		if err := r.deleteWithChildrenTx(tx, child.ID); err != nil {
			return err
		}
	}
	// 删除分组角色关联
	if err := tx.Where("group_id = ?", id).Delete(&model.GroupRole{}).Error; err != nil {
		return err
	}
	// 删除当前分组
	return tx.Where("id = ?", id).Delete(&model.Group{}).Error
}

// GetGroupRoleIDs 获取分组的角色 ID 列表
func (r *GroupRepo) GetGroupRoleIDs(groupID uint) ([]uint, error) {
	var roleIDs []uint
	err := r.db.Model(&model.GroupRole{}).Where("group_id = ?", groupID).Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

// SyncGroupRoles 同步分组角色（事务保护）
func (r *GroupRepo) SyncGroupRoles(groupID uint, roleIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧的关联
		if err := tx.Where("group_id = ?", groupID).Delete(&model.GroupRole{}).Error; err != nil {
			return err
		}

		// 创建新的关联
		for _, roleID := range roleIDs {
			groupRole := &model.GroupRole{
				GroupID: groupID,
				RoleID:  roleID,
			}
			if err := tx.Create(groupRole).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetGroupIDsByRoleID 获取关联了指定角色的所有分组 ID
func (r *GroupRepo) GetGroupIDsByRoleID(roleID uint) ([]uint, error) {
	var groupIDs []uint
	err := r.db.Model(&model.GroupRole{}).Where("role_id = ?", roleID).Pluck("group_id", &groupIDs).Error
	return groupIDs, err
}

// GetGroupRoleIDsRecursive 递归获取分组及其父分组的角色 ID 列表
func (r *GroupRepo) GetGroupRoleIDsRecursive(groupID uint) ([]uint, error) {
	var allRoleIDs []uint
	visited := make(map[uint]bool)

	currentID := groupID
	for currentID != 0 {
		if visited[currentID] {
			break // 防止循环
		}
		visited[currentID] = true

		roleIDs, err := r.GetGroupRoleIDs(currentID)
		if err != nil {
			return nil, err
		}
		allRoleIDs = append(allRoleIDs, roleIDs...)

		// 获取父分组 ID
		group, err := r.GetByID(currentID)
		if err != nil {
			break
		}
		currentID = group.PID
	}

	return allRoleIDs, nil
}
