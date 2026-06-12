package service

import (
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/internal/ws"
	"backend-server/pkg/database"
)

// NotificationService 通知服务
type NotificationService struct {
	notifRepo *repository.NotificationRepo
	hub       *ws.Hub
}

var globalHub *ws.Hub

// SetGlobalHub 设置全局 WebSocket Hub（由 main.go 调用）
func SetGlobalHub(h *ws.Hub) {
	globalHub = h
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		notifRepo: repository.NewNotificationRepo(database.GetMySQL()),
		hub:       globalHub,
	}
}

// NotificationResponse 通知响应
type NotificationResponse struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Link      string    `json:"link"`
	IsRead    bool      `json:"isRead"`
	SenderID  uint      `json:"senderId"`
	CreatedAt time.Time `json:"createdAt"`
}

// 通知类型常量
const (
	NotifTypeLoginAlert    = "login_alert"    // 新设备登录
	NotifTypeUploadDone    = "upload_done"    // 上传完成
	NotifTypeRoleChange    = "role_change"    // 角色变更
	NotifTypeRoleApproved  = "role_approved"  // 角色申请通过
	NotifTypeRoleRejected  = "role_rejected"  // 角色申请拒绝
	NotifTypeAnnouncement  = "announcement"   // 系统公告
	NotifTypeStorageWarn   = "storage_warn"   // 存储空间警告
	NotifTypeSecurityWarn  = "security_warn"  // 安全警告
)

// List 获取用户通知列表
func (s *NotificationService) List(userID uint, page, pageSize int) ([]NotificationResponse, int64, error) {
	offset := (page - 1) * pageSize
	items, total, err := s.notifRepo.List(userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var result []NotificationResponse
	for _, item := range items {
		nr := NotificationResponse{}
		if v, ok := item["id"]; ok {
			nr.ID, _ = v.(uint)
		}
		if v, ok := item["type"]; ok {
			nr.Type, _ = v.(string)
		}
		if v, ok := item["title"]; ok {
			nr.Title, _ = v.(string)
		}
		if v, ok := item["content"]; ok {
			nr.Content, _ = v.(string)
		}
		if v, ok := item["link"]; ok {
			nr.Link, _ = v.(string)
		}
		if v, ok := item["is_read"]; ok {
			switch val := v.(type) {
			case bool:
				nr.IsRead = val
			case int64:
				nr.IsRead = val != 0
			case []uint8:
				nr.IsRead = len(val) > 0 && val[0] != 0
			}
		}
		if v, ok := item["sender_id"]; ok {
			nr.SenderID, _ = v.(uint)
		}
		if v, ok := item["created_at"]; ok {
			if t, ok := v.(time.Time); ok {
				nr.CreatedAt = t
			}
		}
		result = append(result, nr)
	}

	return result, total, nil
}

// GetUnreadCount 获取未读数量
func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	return s.notifRepo.GetUnreadCount(userID)
}

// MarkRead 标记已读
func (s *NotificationService) MarkRead(notificationID, userID uint) error {
	return s.notifRepo.MarkRead(notificationID, userID)
}

// MarkAllRead 全部标记已读
func (s *NotificationService) MarkAllRead(userID uint) error {
	return s.notifRepo.MarkAllRead(userID)
}

// Delete 删除通知
func (s *NotificationService) Delete(notificationID, userID uint) error {
	return s.notifRepo.Delete(notificationID, userID)
}

// CreateAndPush 创建通知并推送（个人通知）
func (s *NotificationService) CreateAndPush(userID uint, notifType, title, content, link string) error {
	n := &model.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Content: content,
		Link:    link,
	}
	if err := s.notifRepo.Create(n); err != nil {
		return err
	}

	// WebSocket 推送
	if s.hub != nil {
		s.hub.SendToUser(userID, &ws.Message{
			Type: "notification",
			Payload: map[string]interface{}{
				"id":        n.ID,
				"type":      notifType,
				"title":     title,
				"content":   content,
				"link":      link,
				"createdAt": n.CreatedAt,
			},
		})
	}

	return nil
}

// CreateBroadcast 创建公告（全员通知）
func (s *NotificationService) CreateBroadcast(senderID uint, title, content, link string) error {
	n := &model.Notification{
		UserID:   0, // 0 = 公告
		Type:     NotifTypeAnnouncement,
		Title:    title,
		Content:  content,
		Link:     link,
		SenderID: senderID,
	}
	if err := s.notifRepo.Create(n); err != nil {
		return err
	}

	// 广播推送
	if s.hub != nil {
		s.hub.BroadcastMessage(&ws.Message{
			Type: "notification",
			Payload: map[string]interface{}{
				"id":        n.ID,
				"type":      NotifTypeAnnouncement,
				"title":     title,
				"content":   content,
				"link":      link,
				"isRead":    false,
				"createdAt": n.CreatedAt,
			},
		})
	}

	return nil
}

// NotifyLoginAlert 新设备登录通知
func (s *NotificationService) NotifyLoginAlert(userID uint, deviceName, ip, location string) {
	title := "新设备登录提醒"
	content := fmt.Sprintf("您的账号在 %s 通过新设备登录（IP: %s，位置: %s）", time.Now().Format("2006-01-02 15:04"), ip, location)
	s.CreateAndPush(userID, NotifTypeLoginAlert, title, content, "/account/index?tab=security")
}

// NotifyUploadDone 上传完成通知
func (s *NotificationService) NotifyUploadDone(userID uint, fileName string, fileID uint) {
	title := "文件上传完成"
	content := fmt.Sprintf("文件 %s 上传成功", fileName)
	link := fmt.Sprintf("/file/list")
	s.CreateAndPush(userID, NotifTypeUploadDone, title, content, link)
}

// NotifyRoleChange 角色变更通知
func (s *NotificationService) NotifyRoleChange(userID uint, roleName, action string) {
	title := "角色变更通知"
	content := fmt.Sprintf("您的角色 %s 已被%s", roleName, action)
	s.CreateAndPush(userID, NotifTypeRoleChange, title, content, "/account/index?tab=profile")
}

// NotifyRoleApplication 角色申请审批结果
func (s *NotificationService) NotifyRoleApplication(userID uint, roleName, status string) {
	var notifType, title, content string
	if status == "approved" {
		notifType = NotifTypeRoleApproved
		title = "角色申请已通过"
		content = fmt.Sprintf("您申请的角色 %s 已通过审批", roleName)
	} else {
		notifType = NotifTypeRoleRejected
		title = "角色申请未通过"
		content = fmt.Sprintf("您申请的角色 %s 未通过审批", roleName)
	}
	s.CreateAndPush(userID, notifType, title, content, "/user-auth/role-app")
}

// NotifyStorageWarning 存储空间警告
func (s *NotificationService) NotifyStorageWarning(userID uint, usedPercent float64) {
	title := "存储空间不足"
	content := fmt.Sprintf("您的存储空间已使用 %.0f%%，请及时清理文件", usedPercent)
	s.CreateAndPush(userID, NotifTypeStorageWarn, title, content, "/file/list")
}

// AdminList 管理员查看所有通知
func (s *NotificationService) AdminList(notifType string, page, pageSize int) ([]model.Notification, int64, error) {
	offset := (page - 1) * pageSize
	return s.notifRepo.ListAdmin(notifType, offset, pageSize)
}
