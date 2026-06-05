package handler

import (
	"strconv"

	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// GroupHandler 分组处理器
type GroupHandler struct {
	groupService *service.GroupService
}

// NewGroupHandler 创建分组处理器
func NewGroupHandler() *GroupHandler {
	return &GroupHandler{
		groupService: service.NewGroupService(),
	}
}

// List 获取分组列表（树形结构）
// @Summary      获取分组列表
// @Description  获取分组树形结构数据，包含每个分组的角色 ID 列表
// @Tags         分组管理
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]service.GroupTreeNode} "分组树"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/group/list [get]
func (h *GroupHandler) List(c *gin.Context) {
	groups, err := h.groupService.List()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, groups)
}

// Create 创建分组
// @Summary      创建分组
// @Description  创建新分组，可同时指定关联的角色列表
// @Tags         分组管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.CreateGroupRequest  true  "创建分组请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/group [post]
func (h *GroupHandler) Create(c *gin.Context) {
	var req service.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.groupService.Create(&req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Update 更新分组
// @Summary      更新分组
// @Description  更新指定分组信息，可同时更新关联的角色列表
// @Tags         分组管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                       true  "分组 ID"
// @Param        request  body  service.UpdateGroupRequest  true  "更新分组请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Failure      404  {object}  response.Response "分组不存在"
// @Router       /system/group/{id} [put]
func (h *GroupHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	var req service.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.groupService.Update(uint(id), &req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除分组
// @Summary      删除分组
// @Description  删除指定分组及其所有子分组（级联软删除），同时清除分组-角色关联
// @Tags         分组管理
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "分组 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Failure      404  {object}  response.Response "分组不存在"
// @Router       /system/group/{id} [delete]
func (h *GroupHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	if err := h.groupService.Delete(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
