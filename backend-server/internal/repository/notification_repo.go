package repository

import (
	"time"

	"backend-server/internal/model"

	"gorm.io/gorm"
)

// NotificationRepo 通知仓储
type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

// Create 创建通知
func (r *NotificationRepo) Create(n *model.Notification) error {
	return r.db.Create(n).Error
}

// GetByID 根据ID获取通知
func (r *NotificationRepo) GetByID(id uint) (*model.Notification, error) {
	var n model.Notification
	err := r.db.Where("id = ?", id).First(&n).Error
	return &n, err
}

// List 获取用户通知列表（个人通知 + 公告）
// 个人通知直接看 is_read 字段，公告通过 sys_notification_reads 判断
func (r *NotificationRepo) List(userID uint, offset, limit int) ([]map[string]interface{}, int64, error) {
	var results []map[string]interface{}
	var total int64

	// 查询个人通知 + 公告总数
	countSQL := `
		SELECT COUNT(*) FROM (
			SELECT id FROM sys_notifications WHERE user_id = ?
			UNION ALL
			SELECT n.id FROM sys_notifications n
			WHERE n.user_id = 0
			AND NOT EXISTS (SELECT 1 FROM sys_notification_reads r WHERE r.notification_id = n.id AND r.user_id = ?)
		) t
	`
	r.db.Raw(countSQL, userID, userID).Scan(&total)

	// 查询列表（合并个人通知和公告，按时间倒序）
	listSQL := `
		SELECT id, user_id, type, title, content, link, is_read, sender_id, created_at
		FROM sys_notifications
		WHERE user_id = ?
		UNION ALL
		SELECT n.id, n.user_id, n.type, n.title, n.content, n.link,
			EXISTS(SELECT 1 FROM sys_notification_reads r WHERE r.notification_id = n.id AND r.user_id = ?) as is_read,
			n.sender_id, n.created_at
		FROM sys_notifications n
		WHERE n.user_id = 0
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	r.db.Raw(listSQL, userID, userID, limit, offset).Scan(&results)

	return results, total, nil
}

// GetUnreadCount 获取未读通知数量
func (r *NotificationRepo) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	sql := `
		SELECT COUNT(*) FROM (
			SELECT id FROM sys_notifications WHERE user_id = ? AND is_read = false
			UNION ALL
			SELECT n.id FROM sys_notifications n
			WHERE n.user_id = 0
			AND NOT EXISTS (SELECT 1 FROM sys_notification_reads r WHERE r.notification_id = n.id AND r.user_id = ?)
		) t
	`
	err := r.db.Raw(sql, userID, userID).Scan(&count).Error
	return count, err
}

// MarkRead 标记单条已读
func (r *NotificationRepo) MarkRead(notificationID, userID uint) error {
	// 先检查是个人通知还是公告
	var n model.Notification
	if err := r.db.Select("user_id").Where("id = ?", notificationID).First(&n).Error; err != nil {
		return err
	}

	if n.UserID == 0 {
		// 公告：插入已读记录（忽略重复）
		return r.db.Where(model.NotificationRead{
			NotificationID: notificationID,
			UserID:         userID,
		}).Attrs(model.NotificationRead{ReadAt: time.Now()}).
			FirstOrCreate(&model.NotificationRead{}).Error
	}

	// 个人通知：更新 is_read
	return r.db.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

// MarkAllRead 标记所有已读
func (r *NotificationRepo) MarkAllRead(userID uint) error {
	// 个人通知全部标记已读
	r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true)

	// 公告：批量插入已读记录（跳过已有的）
	r.db.Exec(`
		INSERT IGNORE INTO sys_notification_reads (notification_id, user_id, read_at)
		SELECT n.id, ?, NOW()
		FROM sys_notifications n
		WHERE n.user_id = 0
		AND NOT EXISTS (SELECT 1 FROM sys_notification_reads r WHERE r.notification_id = n.id AND r.user_id = ?)
	`, userID, userID)

	return nil
}

// Delete 删除通知（仅个人通知可删除）
func (r *NotificationRepo) Delete(notificationID, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&model.Notification{}).Error
}

// ListAdmin 管理员查看所有通知（分页）
func (r *NotificationRepo) ListAdmin(notifType string, offset, limit int) ([]model.Notification, int64, error) {
	var notifications []model.Notification
	var total int64

	query := r.db.Model(&model.Notification{})
	if notifType != "" {
		query = query.Where("type = ?", notifType)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error

	return notifications, total, err
}
