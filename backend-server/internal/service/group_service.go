package service

import (
	"context"
	"fmt"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/cache"
	"backend-server/pkg/database"
)

// GroupService 分组服务
type GroupService struct {
	groupRepo *repository.GroupRepo
	userRepo  *repository.UserRepo
}

// NewGroupService 创建分组服务
func NewGroupService() *GroupService {
	db := database.GetMySQL()
	return &GroupService{
		groupRepo: repository.NewGroupRepo(db),
		userRepo:  repository.NewUserRepo(db),
	}
}

// CreateGroupRequest 创建分组请求
type CreateGroupRequest struct {
	Name    string `json:"name" binding:"required"`
	PID     uint   `json:"pid"`
	Status  int    `json:"status" binding:"required"`
	Remark  string `json:"remark"`
	RoleIDs []uint `json:"roleIds"` // 角色ID列表
}

// UpdateGroupRequest 更新分组请求
type UpdateGroupRequest struct {
	Name    string `json:"name"`
	Status  int    `json:"status"`
	Remark  string `json:"remark"`
	RoleIDs []uint `json:"roleIds"` // 角色ID列表
}

// GroupResponse 分组响应（包含角色ID）
type GroupResponse struct {
	model.Group
	RoleIDs []uint `json:"roleIds"`
}

// List 获取分组列表（树形结构，包含角色ID）
func (s *GroupService) List() ([]*GroupTreeNode, error) {
	groups, err := s.groupRepo.List()
	if err != nil {
		return nil, err
	}

	return s.buildGroupTree(groups, 0), nil
}

// GroupTreeNode 分组树节点（包含角色ID）
type GroupTreeNode struct {
	model.Group
	RoleIDs  []uint           `json:"roleIds"`
	Children []*GroupTreeNode `json:"children,omitempty"`
}

// buildGroupTree 构建分组树
func (s *GroupService) buildGroupTree(groups []model.Group, pid uint) []*GroupTreeNode {
	var trees []*GroupTreeNode
	for _, group := range groups {
		if group.PID == pid {
			roleIDs, _ := s.groupRepo.GetGroupRoleIDs(group.ID)
			node := &GroupTreeNode{
				Group:   group,
				RoleIDs: roleIDs,
			}
			children := s.buildGroupTree(groups, group.ID)
			if len(children) > 0 {
				node.Children = children
			}
			trees = append(trees, node)
		}
	}
	return trees
}

// GetByID 根据 ID 获取分组
func (s *GroupService) GetByID(id uint) (*model.Group, error) {
	return s.groupRepo.GetByID(id)
}

// Create 创建分组
func (s *GroupService) Create(req *CreateGroupRequest) error {
	group := &model.Group{
		PID:    req.PID,
		Name:   req.Name,
		Status: req.Status,
		Remark: req.Remark,
	}

	if err := s.groupRepo.Create(group); err != nil {
		return err
	}

	// 同步分组角色
	if len(req.RoleIDs) > 0 {
		return s.groupRepo.SyncGroupRoles(group.ID, req.RoleIDs)
	}

	return nil
}

// Update 更新分组
func (s *GroupService) Update(id uint, req *UpdateGroupRequest) error {
	group, err := s.groupRepo.GetByID(id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		group.Name = req.Name
	}
	// Status 始终更新（0=禁用, 1=启用）
	group.Status = req.Status
	if req.Remark != "" {
		group.Remark = req.Remark
	}

	if err := s.groupRepo.Update(group); err != nil {
		return err
	}

	// 同步分组角色
	if req.RoleIDs != nil {
		if err := s.groupRepo.SyncGroupRoles(group.ID, req.RoleIDs); err != nil {
			return err
		}
	}

	// 清除该分组所有用户的权限缓存
	s.invalidateCacheForGroup(id)

	return nil
}

// invalidateCacheForGroup 清除指定分组所有用户的权限缓存
func (s *GroupService) invalidateCacheForGroup(groupID uint) {
	userIDs, err := s.userRepo.GetUserIDsByGroupID(groupID)
	if err != nil || len(userIDs) == 0 {
		return
	}
	ctx := context.Background()
	for _, uid := range userIDs {
		_ = cache.Delete(ctx, fmt.Sprintf("permission_codes:%d", uid))
	}
}

// Delete 删除分组（同时清除权限缓存）
func (s *GroupService) Delete(id uint) error {
	// 先清除该分组所有用户的权限缓存
	s.invalidateCacheForGroup(id)

	return s.groupRepo.DeleteWithChildren(id)
}

// buildTree 构建分组树
func (s *GroupService) buildTree(groups []model.Group, pid uint) []model.GroupTree {
	var trees []model.GroupTree
	for _, group := range groups {
		if group.PID == pid {
			tree := model.GroupTree{
				Group: group,
			}
			children := s.buildTree(groups, group.ID)
			if len(children) > 0 {
				tree.Children = make([]*model.GroupTree, len(children))
				for i := range children {
					tree.Children[i] = &children[i]
				}
			}
			trees = append(trees, tree)
		}
	}
	return trees
}
