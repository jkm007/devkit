package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// SecurityLogHandler 安全日志处理器
type SecurityLogHandler struct {
	service *service.SecurityLogService
}

// NewSecurityLogHandler 创建安全日志处理器
func NewSecurityLogHandler() *SecurityLogHandler {
	return &SecurityLogHandler{
		service: service.NewSecurityLogService(),
	}
}

// GetMyLogs 获取当前用户的安全日志
// @Summary      获取当前用户的安全日志
// @Description  当前用户查看自己的安全操作记录
// @Tags         安全日志
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int     false "页码，默认 1"         minimum(1)
// @Param        pageSize   query  int     false "每页条数，默认 20"     minimum(1) maximum(100)
// @Param        eventType  query  string  false "事件类型筛选"
// @Success      200  {object}  response.Response{data=response.PageData{items=[]model.SecurityLog}} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/security-logs [get]
func (h *SecurityLogHandler) GetMyLogs(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	eventType := c.Query("eventType")

	logs, total, err := h.service.ListByUser(userID, page, pageSize, eventType)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, logs, total)
}

// GetAllLogs 获取所有用户的安全日志（管理员）
// @Summary      获取所有用户的安全日志
// @Description  管理员查看所有用户的安全日志，用于安全审计
// @Tags         安全日志
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int     false "页码，默认 1"         minimum(1)
// @Param        pageSize   query  int     false "每页条数，默认 20"     minimum(1) maximum(100)
// @Param        userId     query  string  false "用户 ID"
// @Param        eventType  query  string  false "事件类型筛选"
// @Param        status     query  string  false "状态：0=失败, 1=成功"
// @Param        ip         query  string  false "IP 地址"
// @Param        startTime  query  string  false "开始时间"
// @Param        endTime    query  string  false "结束时间"
// @Success      200  {object}  response.Response{data=response.PageData{items=[]model.SecurityLog}} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/security-logs [get]
func (h *SecurityLogHandler) GetAllLogs(c *gin.Context) {
	var req service.SecurityLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	logs, total, err := h.service.ListAll(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, logs, total)
}
