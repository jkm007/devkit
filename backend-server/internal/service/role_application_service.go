package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/cache"
	"backend-server/pkg/database"
)

// RoleApplicationService 角色申请服务
type RoleApplicationService struct {
	repo     *repository.RoleApplicationRepo
	userRepo *repository.UserRepo
	roleRepo *repository.RoleRepo
}

// NewRoleApplicationService 创建角色申请服务
func NewRoleApplicationService() *RoleApplicationService {
	db := database.GetMySQL()
	return &RoleApplicationService{
		repo:     repository.NewRoleApplicationRepo(db),
		userRepo: repository.NewUserRepo(db),
		roleRepo: repository.NewRoleRepo(db),
	}
}

// RoleApplicationRequest 角色申请请求
type RoleApplicationRequest struct {
	RoleID uint   `json:"roleId" binding:"required"`
	Reason string `json:"reason"`
}

// RoleApplicationReviewRequest 角色申请审核请求
type RoleApplicationReviewRequest struct {
	Note string `json:"note"`
}

// RoleApplicationListRequest 角色申请列表请求
type RoleApplicationListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=500"`
	Status   string `form:"status"`
	UserID   string `form:"userId"`
	RoleID   string `form:"roleId"`
}

// AvailableRoleItem 可申请角色
type AvailableRoleItem struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Remark string `json:"remark"`
}

// RoleApplicationItem 角色申请列表项
type RoleApplicationItem struct {
	ID           uint       `json:"id"`
	UserID       uint       `json:"userId"`
	Username     string     `json:"username"`
	Nickname     string     `json:"nickname"`
	RoleID       uint       `json:"roleId"`
	RoleName     string     `json:"roleName"`
	RoleRemark   string     `json:"roleRemark"`
	Reason       string     `json:"reason"`
	Status       int        `json:"status"`
	ReviewNote   string     `json:"reviewNote"`
	ReviewedBy   *uint      `json:"reviewedBy"`
	ReviewerName string     `json:"reviewerName"`
	ReviewedAt   *time.Time `json:"reviewedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// Create 创建角色申请
func (s *RoleApplicationService) Create(userID uint, req *RoleApplicationRequest) error {
	role, err := s.roleRepo.GetByID(req.RoleID)
	if err != nil {
		return errors.New("申请角色不存在")
	}
	if role.Status != 1 {
		return errors.New("该角色已禁用，不能申请")
	}
	if role.AllowApply != 1 {
		return errors.New("该角色不允许申请")
	}

	roleIDs, err := s.userRepo.GetUserRoleIDs(userID)
	if err != nil {
		return err
	}
	if containsUint(roleIDs, req.RoleID) {
		return errors.New("您已拥有该角色")
	}

	hasPending, err := s.repo.HasPending(userID, req.RoleID)
	if err != nil {
		return err
	}
	if hasPending {
		return errors.New("您已有该角色的待审核申请")
	}

	application := &model.RoleApplication{
		UserID: userID,
		RoleID: req.RoleID,
		Reason: req.Reason,
		Status: 0,
	}
	return s.repo.Create(application)
}

// ListAvailableRoles 获取当前用户可申请角色
func (s *RoleApplicationService) ListAvailableRoles(userID uint) ([]AvailableRoleItem, error) {
	userRoleIDs, err := s.userRepo.GetUserRoleIDs(userID)
	if err != nil {
		return nil, err
	}
	pendingRoleIDs, err := s.repo.GetPendingRoleIDs(userID)
	if err != nil {
		return nil, err
	}

	excludeIDMap := make(map[uint]bool)
	for _, id := range userRoleIDs {
		excludeIDMap[id] = true
	}
	for _, id := range pendingRoleIDs {
		excludeIDMap[id] = true
	}
	excludeIDs := make([]uint, 0, len(excludeIDMap))
	for id := range excludeIDMap {
		excludeIDs = append(excludeIDs, id)
	}

	roles, err := s.roleRepo.ListAvailableForApply(excludeIDs)
	if err != nil {
		return nil, err
	}

	items := make([]AvailableRoleItem, 0, len(roles))
	for _, role := range roles {
		items = append(items, AvailableRoleItem{
			ID:     role.ID,
			Name:   role.Name,
			Remark: role.Remark,
		})
	}
	return items, nil
}

// ListByUser 获取用户的申请列表
func (s *RoleApplicationService) ListByUser(userID uint, page, pageSize int) ([]RoleApplicationItem, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	list, total, err := s.repo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return s.buildApplicationItems(list), total, nil
}

// ListAll 获取所有申请（管理员）
func (s *RoleApplicationService) ListAll(req *RoleApplicationListRequest) ([]RoleApplicationItem, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filters := map[string]interface{}{
		"status": req.Status,
		"userId": req.UserID,
		"roleId": req.RoleID,
	}

	list, total, err := s.repo.ListAll(req.Page, req.PageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	return s.buildApplicationItems(list), total, nil
}

// Approve 审核通过
func (s *RoleApplicationService) Approve(id, reviewerID uint, note string) error {
	app, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if app.Status != 0 {
		return errors.New("申请已审核，不能重复处理")
	}

	role, err := s.roleRepo.GetByID(app.RoleID)
	if err != nil {
		return errors.New("申请角色不存在")
	}
	if role.Status != 1 {
		return errors.New("申请角色已禁用，不能通过")
	}

	roleIDs, err := s.userRepo.GetUserRoleIDs(app.UserID)
	if err != nil {
		return err
	}
	if containsUint(roleIDs, app.RoleID) {
		return errors.New("申请人已拥有该角色")
	}

	now := time.Now()
	app.Status = 1
	app.ReviewedBy = &reviewerID
	app.ReviewedAt = &now
	app.ReviewNote = note

	if err := s.repo.Update(app); err != nil {
		return err
	}

	// 给用户追加角色（不影响已有角色）
	if err := s.userRepo.AddUserRole(app.UserID, app.RoleID); err != nil {
		return err
	}

	// 清除用户权限缓存，使新角色权限立即生效
	ctx := context.Background()
	_ = cache.Delete(ctx, fmt.Sprintf("permission_codes:%d", app.UserID))

	// 异步通知申请人
	go NewNotificationService().NotifyRoleApplication(app.UserID, role.Name, "approved")
	return nil
}

// Reject 审核拒绝
func (s *RoleApplicationService) Reject(id, reviewerID uint, note string) error {
	app, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if app.Status != 0 {
		return errors.New("申请已审核，不能重复处理")
	}

	now := time.Now()
	app.Status = 2
	app.ReviewedBy = &reviewerID
	app.ReviewedAt = &now
	app.ReviewNote = note

	if err := s.repo.Update(app); err != nil {
		return err
	}

	// 获取角色名称并异步通知申请人
	roleName := ""
	if role, err := s.roleRepo.GetByID(app.RoleID); err == nil {
		roleName = role.Name
	}
	go NewNotificationService().NotifyRoleApplication(app.UserID, roleName, "rejected")

	return nil
}

func (s *RoleApplicationService) buildApplicationItems(list []model.RoleApplication) []RoleApplicationItem {
	userIDSet := make(map[uint]bool)
	roleIDSet := make(map[uint]bool)
	for _, app := range list {
		userIDSet[app.UserID] = true
		roleIDSet[app.RoleID] = true
		if app.ReviewedBy != nil {
			userIDSet[*app.ReviewedBy] = true
		}
	}

	userIDs := make([]uint, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}
	roleIDs := make([]uint, 0, len(roleIDSet))
	for id := range roleIDSet {
		roleIDs = append(roleIDs, id)
	}

	userMap, _ := s.userRepo.GetByIDs(userIDs)
	roles, _ := s.roleRepo.GetByIDs(roleIDs)
	roleMap := make(map[uint]model.Role, len(roles))
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	items := make([]RoleApplicationItem, 0, len(list))
	for _, app := range list {
		item := RoleApplicationItem{
			ID:         app.ID,
			UserID:     app.UserID,
			RoleID:     app.RoleID,
			Reason:     app.Reason,
			Status:     app.Status,
			ReviewNote: app.ReviewNote,
			ReviewedBy: app.ReviewedBy,
			ReviewedAt: app.ReviewedAt,
			CreatedAt:  app.CreatedAt,
		}
		if userMap != nil {
			if user, ok := userMap[app.UserID]; ok {
				item.Username = user.Name
				item.Nickname = user.Nickname
			}
			if app.ReviewedBy != nil {
				if reviewer, ok := userMap[*app.ReviewedBy]; ok {
					item.ReviewerName = reviewer.Name
				}
			}
		}
		if role, ok := roleMap[app.RoleID]; ok {
			item.RoleName = role.Name
			item.RoleRemark = role.Remark
		}
		items = append(items, item)
	}
	return items
}

func containsUint(values []uint, target uint) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
