package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	notifService *service.NotificationService
}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{
		notifService: service.NewNotificationService(),
	}
}

// List 获取当前用户的通知列表
func (h *NotificationHandler) List(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	items, total, err := h.notifService.List(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, items, total)
}

// GetUnreadCount 获取未读通知数量
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	count, err := h.notifService.GetUnreadCount(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, map[string]int64{"count": count})
}

// MarkRead 标记单条已读
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的通知ID")
		return
	}

	if err := h.notifService.MarkRead(uint(id), userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// MarkAllRead 全部标记已读
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	if err := h.notifService.MarkAllRead(userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除通知
func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的通知ID")
		return
	}

	if err := h.notifService.Delete(uint(id), userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// AdminList 管理员查看所有通知
func (h *NotificationHandler) AdminList(c *gin.Context) {
	notifType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.notifService.AdminList(notifType, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, items, total)
}

// PublishAnnouncement 发布公告
type publishAnnouncementRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Link    string `json:"link"`
}

func (h *NotificationHandler) PublishAnnouncement(c *gin.Context) {
	var req publishAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	senderID := middleware.GetCurrentUserID(c)

	if err := h.notifService.CreateBroadcast(senderID, req.Title, req.Content, req.Link); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "公告发布成功", nil)
}
