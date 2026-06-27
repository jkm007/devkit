package service

import (
	"context"
	"encoding/json"
	"fmt"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/cache"
	"backend-server/pkg/database"
	"backend-server/pkg/logger"

	"go.uber.org/zap"
)

// RoleService 角色服务
type RoleService struct {
	roleRepo  *repository.RoleRepo
	userRepo  *repository.UserRepo
	groupRepo *repository.GroupRepo
}

// NewRoleService 创建角色服务
func NewRoleService() *RoleService {
	db := database.GetMySQL()
	return &RoleService{
		roleRepo:  repository.NewRoleRepo(db),
		userRepo:  repository.NewUserRepo(db),
		groupRepo: repository.NewGroupRepo(db),
	}
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Status       int      `json:"status"`
	AllowApply   int      `json:"allowApply"`
	Permissions  []string `json:"permissions"`
	StorageQuota int64    `json:"storageQuota"`
	Remark       string   `json:"remark"`
	CreatedAt    string   `json:"createTime"`
}

// toRoleResponse 将模型转换为响应结构
func toRoleResponse(role *model.Role) RoleResponse {
	var permissions []string
	if role.Permissions != "" {
		// 反序列化角色权限列表
		if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
			logger.Error("角色权限反序列化失败",
				zap.Uint("role_id", role.ID),
				zap.String("permissions", role.Permissions),
				zap.Error(err),
			)
		}
	}
	if permissions == nil {
		permissions = []string{}
	}
	return RoleResponse{
		ID:           role.ID,
		Name:         role.Name,
		Status:       role.Status,
		AllowApply:   role.AllowApply,
		Permissions:  permissions,
		StorageQuota: role.StorageQuota,
		Remark:       role.Remark,
		CreatedAt:    role.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
}

// ListRoleRequest 角色列表请求
type ListRoleRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"pageSize" binding:"omitempty,min=1,max=500"`
	Name      string `form:"name"`
	ID        string `form:"id"`
	Status    string `form:"status"`
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
	Remark    string `form:"remark"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name         string   `json:"name" binding:"required"`
	Status       int      `json:"status" binding:"required"`
	AllowApply   int      `json:"allowApply"`
	Permissions  []string `json:"permissions"`
	StorageQuota int64    `json:"storageQuota"`
	Remark       string   `json:"remark"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name         string   `json:"name"`
	Status       int      `json:"status"`
	AllowApply   *int     `json:"allowApply"`
	Permissions  []string `json:"permissions"`
	StorageQuota *int64   `json:"storageQuota"`
	Remark       string   `json:"remark"`
}

// List 获取角色列表
func (s *RoleService) List(req *ListRoleRequest) ([]RoleResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filters := map[string]interface{}{
		"name":      req.Name,
		"id":        req.ID,
		"status":    req.Status,
		"startTime": req.StartTime,
		"endTime":   req.EndTime,
		"remark":    req.Remark,
	}

	roles, total, err := s.roleRepo.List(req.Page, req.PageSize, filters)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应结构
	responses := make([]RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = toRoleResponse(&role)
	}

	return responses, total, nil
}

// GetByID 根据 ID 获取角色
func (s *RoleService) GetByID(id uint) (*model.Role, error) {
	return s.roleRepo.GetByID(id)
}

// Create 创建角色
func (s *RoleService) Create(req *CreateRoleRequest) error {
	role := &model.Role{
		Name:         req.Name,
		Status:       req.Status,
		AllowApply:   req.AllowApply,
		StorageQuota: req.StorageQuota,
		Remark:       req.Remark,
	}

	// 将权限列表转换为 JSON 字符串
	if len(req.Permissions) > 0 {
		permissionsJSON, err := json.Marshal(req.Permissions)
		if err != nil {
			return err
		}
		role.Permissions = string(permissionsJSON)
	}

	return s.roleRepo.Create(role)
}

// Update 更新角色
func (s *RoleService) Update(id uint, req *UpdateRoleRequest) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 系统内置角色不允许修改名称
	if isProtectedRole(role.Name) && req.Name != "" && req.Name != role.Name {
		return fmt.Errorf("系统内置角色【%s】不允许修改名称", role.Name)
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	// Status 始终更新（0=禁用, 1=启用）
	role.Status = req.Status
	if req.AllowApply != nil {
		role.AllowApply = *req.AllowApply
	}
	if req.StorageQuota != nil {
		role.StorageQuota = *req.StorageQuota
	}
	if req.Remark != "" {
		role.Remark = req.Remark
	}

	// 更新权限
	if req.Permissions != nil {
		permissionsJSON, err := json.Marshal(req.Permissions)
		if err != nil {
			return err
		}
		role.Permissions = string(permissionsJSON)
	}

	if err := s.roleRepo.Update(role); err != nil {
		return err
	}

	// 清除拥有该角色的所有用户的权限缓存
	s.invalidateCacheForRole(id)

	return nil
}

// invalidateCacheForRole 清除拥有指定角色的所有用户的权限缓存（包括通过分组继承的用户）
func (s *RoleService) invalidateCacheForRole(roleID uint) {
	ctx := context.Background()

	// 清除直接关联该角色的用户的缓存
	userIDs, err := s.userRepo.GetUserIDsByRoleID(roleID)
	if err == nil {
		for _, uid := range userIDs {
			_ = cache.Delete(ctx, fmt.Sprintf("permission_codes:%d", uid))
		}
	}

	// 清除通过分组继承该角色的用户的缓存
	groupIDs, err := s.groupRepo.GetGroupIDsByRoleID(roleID)
	if err == nil {
		for _, groupID := range groupIDs {
			groupUserIDs, err := s.userRepo.GetUserIDsByGroupID(groupID)
			if err == nil {
				for _, uid := range groupUserIDs {
					_ = cache.Delete(ctx, fmt.Sprintf("permission_codes:%d", uid))
				}
			}
		}
	}
}

// protectedRoles 系统内置角色，不允许删除和修改名称
var protectedRoles = map[string]bool{
	"admin":       true,
	"super_admin": true,
}

// isProtectedRole 检查是否为系统保护角色
func isProtectedRole(name string) bool {
	return protectedRoles[name]
}

// Delete 删除角色（同时清理关联数据和权限缓存）
func (s *RoleService) Delete(id uint) error {
	// 获取角色信息
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 系统内置角色不允许删除
	if isProtectedRole(role.Name) {
		return fmt.Errorf("系统内置角色【%s】不允许删除", role.Name)
	}

	// 检查是否有用户直接关联该角色
	userIDs, err := s.userRepo.GetUserIDsByRoleID(id)
	if err != nil {
		return err
	}
	if len(userIDs) > 0 {
		return fmt.Errorf("该角色下还有 %d 个用户，请先移除用户的角色关联", len(userIDs))
	}

	// 检查是否有用户通过分组继承该角色
	groupIDs, err := s.groupRepo.GetGroupIDsByRoleID(id)
	if err != nil {
		return err
	}
	var inheritedUserCount int
	for _, groupID := range groupIDs {
		groupUserIDs, err := s.userRepo.GetUserIDsByGroupID(groupID)
		if err != nil {
			return err
		}
		inheritedUserCount += len(groupUserIDs)
	}
	if inheritedUserCount > 0 {
		return fmt.Errorf("该角色通过分组继承关联了 %d 个用户，请先移除分组的角色关联", inheritedUserCount)
	}

	// 先清除拥有该角色的所有用户的权限缓存
	s.invalidateCacheForRole(id)

	// 删除角色（软删除）
	if err := s.roleRepo.Delete(id); err != nil {
		return err
	}

	// 清理用户-角色关联（忽略错误，角色已删除）
	db := database.GetMySQL()
	if err := db.Where("role_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
		// 记录日志但不影响返回（角色已删除）
		logger.Error("清理 UserRole 关联失败",
			zap.Uint("role_id", id),
			zap.Error(err),
		)
	}
	if err := db.Where("role_id = ?", id).Delete(&model.GroupRole{}).Error; err != nil {
		logger.Error("清理 GroupRole 关联失败",
			zap.Uint("role_id", id),
			zap.Error(err),
		)
	}

	return nil
}
