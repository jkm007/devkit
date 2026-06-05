package service

import (
	"errors"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
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
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status"`
	UserID   string `form:"userId"`
}

// Create 创建角色申请
func (s *RoleApplicationService) Create(userID uint, req *RoleApplicationRequest) error {
	// 检查是否已有相同角色的待审申请
	existingApps, _, err := s.repo.ListByUser(userID, 1, 100)
	if err != nil {
		return err
	}
	for _, app := range existingApps {
		if app.RoleID == req.RoleID && app.Status == 0 {
			return errors.New("You already have a pending application for this role.")
		}
	}

	application := &model.RoleApplication{
		UserID: userID,
		RoleID: req.RoleID,
		Reason: req.Reason,
		Status: 0,
	}
	return s.repo.Create(application)
}

// ListByUser 获取用户的申请列表
func (s *RoleApplicationService) ListByUser(userID uint, page, pageSize int) ([]model.RoleApplication, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.ListByUser(userID, page, pageSize)
}

// ListAll 获取所有申请（管理员）
func (s *RoleApplicationService) ListAll(req *RoleApplicationListRequest) ([]model.RoleApplication, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filters := map[string]interface{}{
		"status": req.Status,
		"userId": req.UserID,
	}

	return s.repo.ListAll(req.Page, req.PageSize, filters)
}

// Approve 审核通过
func (s *RoleApplicationService) Approve(id, reviewerID uint, note string) error {
	app, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if app.Status != 0 {
		return errors.New("Application already reviewed.")
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
	return s.userRepo.AddUserRole(app.UserID, app.RoleID)
}

// Reject 审核拒绝
func (s *RoleApplicationService) Reject(id, reviewerID uint, note string) error {
	app, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if app.Status != 0 {
		return errors.New("Application already reviewed.")
	}

	now := time.Now()
	app.Status = 2
	app.ReviewedBy = &reviewerID
	app.ReviewedAt = &now
	app.ReviewNote = note

	return s.repo.Update(app)
}
