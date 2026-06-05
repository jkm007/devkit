package service

import (
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

// SecurityLogService 安全日志服务
type SecurityLogService struct {
	repo *repository.SecurityLogRepo
}

// NewSecurityLogService 创建安全日志服务
func NewSecurityLogService() *SecurityLogService {
	return &SecurityLogService{
		repo: repository.NewSecurityLogRepo(database.GetMySQL()),
	}
}

// SecurityLogListRequest 安全日志列表请求
type SecurityLogListRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	UserID    string `form:"userId"`
	EventType string `form:"eventType"`
	Status    string `form:"status"`
	IP        string `form:"ip"`
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
}

// Record 记录安全日志
func (s *SecurityLogService) Record(userID uint, eventType, detail, ip, userAgent string, status int) error {
	log := &model.SecurityLog{
		UserID:      userID,
		EventType:   eventType,
		EventDetail: detail,
		IP:          ip,
		UserAgent:   userAgent,
		Status:      status,
	}
	return s.repo.Create(log)
}

// ListByUser 获取当前用户的日志
func (s *SecurityLogService) ListByUser(userID uint, page, pageSize int, eventType string) ([]model.SecurityLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.ListByUser(userID, page, pageSize, eventType)
}

// ListAll 获取所有日志（管理员）
func (s *SecurityLogService) ListAll(req *SecurityLogListRequest) ([]model.SecurityLog, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filters := map[string]interface{}{
		"userId":    req.UserID,
		"eventType": req.EventType,
		"status":    req.Status,
		"ip":        req.IP,
		"startTime": req.StartTime,
		"endTime":   req.EndTime,
	}

	return s.repo.ListAll(req.Page, req.PageSize, filters)
}
