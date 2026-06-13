package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// QuestionFeedbackHandler 题目纠错处理器
type QuestionFeedbackHandler struct {
	feedbackService *service.QuestionFeedbackService
}

// NewQuestionFeedbackHandler 创建题目纠错处理器
func NewQuestionFeedbackHandler() *QuestionFeedbackHandler {
	return &QuestionFeedbackHandler{
		feedbackService: service.NewQuestionFeedbackService(),
	}
}

// Create 提交纠错反馈（移动端）
func (h *QuestionFeedbackHandler) Create(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	var req service.CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.feedbackService.Create(userID, &req); err != nil {
		response.InternalError(c, "提交纠错失败")
		return
	}

	response.SuccessWithMessage(c, "纠错反馈已提交", nil)
}

// List 获取纠错列表（移动端）
func (h *QuestionFeedbackHandler) List(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	items, total, err := h.feedbackService.List(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, "获取纠错列表失败")
		return
	}

	response.SuccessPage(c, items, total)
}

// GetDetail 获取纠错详情（移动端）
func (h *QuestionFeedbackHandler) GetDetail(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的纠错ID")
		return
	}

	fb, err := h.feedbackService.GetByID(userID, uint(id))
	if err != nil {
		response.NotFound(c, "纠错记录不存在")
		return
	}

	response.Success(c, fb)
}

// Delete 删除纠错反馈（移动端）
func (h *QuestionFeedbackHandler) Delete(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的纠错ID")
		return
	}

	if err := h.feedbackService.Delete(userID, uint(id)); err != nil {
		response.InternalError(c, "删除纠错失败")
		return
	}

	response.SuccessWithMessage(c, "已删除", nil)
}

// AdminList 管理端列表
func (h *QuestionFeedbackHandler) AdminList(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	items, total, err := h.feedbackService.AdminList(status, page, pageSize)
	if err != nil {
		response.InternalError(c, "获取纠错列表失败")
		return
	}

	response.SuccessPage(c, items, total)
}

// AdminUpdate 管理端更新
func (h *QuestionFeedbackHandler) AdminUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的纠错ID")
		return
	}

	var req struct {
		Status     string `json:"status" binding:"required"`
		AdminReply string `json:"adminReply"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.feedbackService.AdminUpdate(uint(id), req.Status, req.AdminReply); err != nil {
		response.InternalError(c, "更新纠错状态失败")
		return
	}

	response.SuccessWithMessage(c, "已更新", nil)
}
