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

// CreateRequest 创建任务请求
type CreateRequest struct {
	Name     string        `json:"name" binding:"required"`
	TaskType string        `json:"taskType" binding:"required,oneof=recycle_cleanup"`
	CronExpr string        `json:"cronExpr" binding:"required"`
	Config   model.JSONMap `json:"config"`
}

// Create 创建任务
// @Summary      创建定时任务
// @Description  创建新的定时任务
// @Tags         定时任务
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  CreateRequest  true  "任务信息"
// @Success      200  {object}  response.Response{data=model.ScheduledTask} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/scheduled-tasks [post]
func (h *ScheduledTaskHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := h.taskService.Create(req.Name, req.TaskType, req.CronExpr, req.Config)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, task)
}

// List 获取所有任务
// @Summary      获取定时任务列表
// @Description  获取所有定时任务
// @Tags         定时任务
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]model.ScheduledTask} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/scheduled-tasks [get]
func (h *ScheduledTaskHandler) List(c *gin.Context) {
	tasks, err := h.taskService.GetAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, tasks)
}

// GetByID 获取任务详情
// @Summary      获取定时任务详情
// @Description  根据 ID 获取定时任务详情
// @Tags         定时任务
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "任务 ID"
// @Success      200  {object}  response.Response{data=model.ScheduledTask} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/scheduled-tasks/{id} [get]
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
// @Summary      更新定时任务
// @Description  更新指定定时任务
// @Tags         定时任务
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "任务 ID"
// @Param        data  body  UpdateRequest  true  "任务信息"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/scheduled-tasks/{id} [put]
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
// @Summary      更新定时任务启用状态
// @Description  启用或禁用指定定时任务
// @Tags         定时任务
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "任务 ID"
// @Param        data  body  object  true  "启用状态"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/scheduled-tasks/{id}/enabled [put]
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

// Delete 删除任务
// @Summary      删除定时任务
// @Description  删除指定定时任务
// @Tags         定时任务
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "任务 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/scheduled-tasks/{id} [delete]
func (h *ScheduledTaskHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.taskService.Delete(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Run 手动执行任务
// @Summary      执行定时任务
// @Description  手动执行指定定时任务
// @Tags         定时任务
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "任务 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/scheduled-tasks/{id}/run [post]
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
