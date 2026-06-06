package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// SecurityLogRepo 安全日志仓库
type SecurityLogRepo struct {
	db *gorm.DB
}

// NewSecurityLogRepo 创建安全日志仓库
func NewSecurityLogRepo(db *gorm.DB) *SecurityLogRepo {
	return &SecurityLogRepo{db: db}
}

// Create 创建日志
func (r *SecurityLogRepo) Create(log *model.SecurityLog) error {
	return r.db.Create(log).Error
}

// ListByUser 按用户查询日志（分页）
func (r *SecurityLogRepo) ListByUser(userID uint, page, pageSize int, eventType string) ([]model.SecurityLog, int64, error) {
	var logs []model.SecurityLog
	var total int64

	query := r.db.Model(&model.SecurityLog{}).
		Select("sys_security_logs.*, sys_users.name as username").
		Joins("LEFT JOIN sys_users ON sys_users.id = sys_security_logs.user_id").
		Where("sys_security_logs.user_id = ?", userID)
	if eventType != "" {
		query = query.Where("sys_security_logs.event_type = ?", eventType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sys_security_logs.created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListAll 查询所有日志（管理员，分页）
func (r *SecurityLogRepo) ListAll(page, pageSize int, filters map[string]interface{}) ([]model.SecurityLog, int64, error) {
	var logs []model.SecurityLog
	var total int64

	query := r.db.Model(&model.SecurityLog{}).
		Select("sys_security_logs.*, sys_users.name as username").
		Joins("LEFT JOIN sys_users ON sys_users.id = sys_security_logs.user_id")

	if userID, ok := filters["userId"]; ok && userID != "" {
		query = query.Where("sys_security_logs.user_id = ?", userID)
	}
	if eventType, ok := filters["eventType"]; ok && eventType != "" {
		query = query.Where("sys_security_logs.event_type = ?", eventType)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("sys_security_logs.status = ?", status)
	}
	if ip, ok := filters["ip"]; ok && ip != "" {
		if ipStr, ok := ip.(string); ok {
			query = query.Where("sys_security_logs.ip LIKE ?", "%"+escapeLike(ipStr)+"%")
		}
	}
	if startTime, ok := filters["startTime"]; ok && startTime != "" {
		query = query.Where("sys_security_logs.created_at >= ?", startTime)
	}
	if endTime, ok := filters["endTime"]; ok && endTime != "" {
		query = query.Where("sys_security_logs.created_at <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sys_security_logs.created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
