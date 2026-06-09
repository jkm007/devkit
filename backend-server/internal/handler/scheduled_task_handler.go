package handler

import (
	"strconv"

	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// ScheduledTaskHandler 定时任务处理器
type ScheduledTaskHandler struct {
	taskService *service.ScheduledTaskService
}

func NewScheduledTaskHandler() *ScheduledTaskHandler {
	return &ScheduledTaskHandler{
		taskService: service.NewScheduledTaskService(),
	}
}

// List 获取所有任务
func (h *ScheduledTaskHandler) List(c *gin.Context) {
	tasks, err := h.taskService.GetAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, tasks)
}

// GetByID 获取任务详情
func (h *ScheduledTaskHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	task, err := h.taskService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "任务不存在")
		return
	}

	response.Success(c, task)
}

// UpdateRequest 更新任务请求
type UpdateRequest struct {
	Name     string        `json:"name" binding:"required"`
	CronExpr string        `json:"cronExpr" binding:"required"`
	Config   model.JSONMap `json:"config"`
	Enabled  *bool         `json:"enabled"`
}

// Update 更新任务
func (h *ScheduledTaskHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if err := h.taskService.Update(uint(id), req.Name, req.CronExpr, req.Config, enabled); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateEnabled 更新启用状态
func (h *ScheduledTaskHandler) UpdateEnabled(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.taskService.UpdateEnabled(uint(id), req.Enabled); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Run 手动执行任务
func (h *ScheduledTaskHandler) Run(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.taskService.RunTask(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "任务已执行"})
}
