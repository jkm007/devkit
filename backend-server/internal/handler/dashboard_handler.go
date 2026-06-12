package handler

import (
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// DashboardHandler 仪表盘处理器
type DashboardHandler struct {
	service *service.DashboardService
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{
		service: service.NewDashboardService(),
	}
}

// GetStats 获取仪表盘统计
// @Summary      获取仪表盘统计数据
// @Description  获取系统概览、事件趋势、设备分布等统计信息
// @Tags         仪表盘
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=service.DashboardStats} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /system/dashboard/stats [get]
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}
