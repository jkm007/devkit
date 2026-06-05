package service

import (
	"context"
	"encoding/json"
	"fmt"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/cache"
	"backend-server/pkg/database"
)

// RoleService 角色服务
type RoleService struct {
	roleRepo *repository.RoleRepo
	userRepo *repository.UserRepo
}

// NewRoleService 创建角色服务
func NewRoleService() *RoleService {
	db := database.GetMySQL()
	return &RoleService{
		roleRepo: repository.NewRoleRepo(db),
		userRepo: repository.NewUserRepo(db),
	}
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Status      int      `json:"status"`
	Permissions []string `json:"permissions"`
	Remark      string   `json:"remark"`
	CreatedAt   string   `json:"createTime"`
}

// toRoleResponse 将模型转换为响应结构
func toRoleResponse(role *model.Role) RoleResponse {
	var permissions []string
	if role.Permissions != "" {
		json.Unmarshal([]byte(role.Permissions), &permissions)
	}
	if permissions == nil {
		permissions = []string{}
	}
	return RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Status:      role.Status,
		Permissions: permissions,
		Remark:      role.Remark,
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
}

// ListRoleRequest 角色列表请求
type ListRoleRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Name      string `form:"name"`
	ID        string `form:"id"`
	Status    string `form:"status"`
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
	Remark    string `form:"remark"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Status      int      `json:"status" binding:"required"`
	Permissions []string `json:"permissions"`
	Remark      string   `json:"remark"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string   `json:"name"`
	Status      int      `json:"status"`
	Permissions []string `json:"permissions"`
	Remark      string   `json:"remark"`
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
		Name:   req.Name,
		Status: req.Status,
		Remark: req.Remark,
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

	if req.Name != "" {
		role.Name = req.Name
	}
	// Status 始终更新（0=禁用, 1=启用）
	role.Status = req.Status
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

// invalidateCacheForRole 清除拥有指定角色的所有用户的权限缓存
func (s *RoleService) invalidateCacheForRole(roleID uint) {
	userIDs, err := s.userRepo.GetUserIDsByRoleID(roleID)
	if err != nil || len(userIDs) == 0 {
		return
	}
	ctx := context.Background()
	for _, uid := range userIDs {
		_ = cache.Delete(ctx, fmt.Sprintf("permission_codes:%d", uid))
	}
}

// Delete 删除角色（同时清理关联数据和权限缓存）
func (s *RoleService) Delete(id uint) error {
	// 先清除拥有该角色的所有用户的权限缓存
	s.invalidateCacheForRole(id)

	// 删除角色（软删除）
	if err := s.roleRepo.Delete(id); err != nil {
		return err
	}

	// 清理用户-角色关联
	db := database.GetMySQL()
	db.Where("role_id = ?", id).Delete(&model.UserRole{})
	db.Where("role_id = ?", id).Delete(&model.GroupRole{})

	return nil
}
