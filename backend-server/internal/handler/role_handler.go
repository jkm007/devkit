package handler

import (
	"encoding/json"
	"strconv"

	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler() *RoleHandler {
	return &RoleHandler{
		roleService: service.NewRoleService(),
	}
}

// List 获取角色列表
// @Summary      获取角色列表
// @Description  获取角色列表数据，支持分页和多条件筛选
// @Tags         角色管理
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int     false "页码，默认 1"           minimum(1)
// @Param        pageSize   query  int     false "每页条数，默认 20"       minimum(1) maximum(100)
// @Param        name       query  string  false "角色名称（模糊搜索）"
// @Param        id         query  string  false "角色 ID（精确匹配）"
// @Param        status     query  string  false "状态：0=禁用, 1=启用"
// @Param        startTime  query  string  false "创建时间起始"
// @Param        endTime    query  string  false "创建时间结束"
// @Param        remark     query  string  false "备注（模糊搜索）"
// @Success      200  {object}  response.Response{data=response.PageData{items=[]service.RoleResponse}} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/role/list [get]
func (h *RoleHandler) List(c *gin.Context) {
	var req service.ListRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	roles, total, err := h.roleService.List(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, roles, total)
}

// GetDetail 获取角色详情
// @Summary      获取角色详情
// @Description  根据 ID 获取角色详情，包含完整的权限码列表
// @Tags         角色管理
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "角色 ID"
// @Success      200  {object}  response.Response{data=service.RoleResponse} "角色详情"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Failure      404  {object}  response.Response "角色不存在"
// @Router       /system/role/{id} [get]
func (h *RoleHandler) GetDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	role, err := h.roleService.GetByID(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 转换为响应结构
	var permissions []string
	if role.Permissions != "" {
		json.Unmarshal([]byte(role.Permissions), &permissions)
	}
	if permissions == nil {
		permissions = []string{}
	}

	response.Success(c, service.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Status:      role.Status,
		Permissions: permissions,
		Remark:      role.Remark,
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	})
}

// Create 创建角色
// @Summary      创建角色
// @Description  创建新角色，可同时设置权限码列表
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.CreateRoleRequest  true  "创建角色请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/role [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.roleService.Create(&req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Update 更新角色
// @Summary      更新角色
// @Description  更新指定角色信息，可同时更新权限码列表
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                       true  "角色 ID"
// @Param        request  body  service.UpdateRoleRequest  true  "更新角色请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Failure      404  {object}  response.Response "角色不存在"
// @Router       /system/role/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.roleService.Update(uint(id), &req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除角色
// @Summary      删除角色
// @Description  删除指定角色（软删除）
// @Tags         角色管理
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "角色 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Failure      404  {object}  response.Response "角色不存在"
// @Router       /system/role/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	if err := h.roleService.Delete(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
