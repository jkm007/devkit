package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	menuService *service.MenuService
}

// NewMenuHandler 创建菜单处理器
func NewMenuHandler() *MenuHandler {
	return &MenuHandler{
		menuService: service.NewMenuService(),
	}
}

// GetAll 获取用户菜单（侧边栏渲染）
// @Summary      获取用户菜单
// @Description  登录后获取当前用户可访问的菜单树（过滤按钮类型），用于渲染侧边栏导航
// @Tags         菜单管理
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]service.MenuItem} "菜单树"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /menu/all [get]
func (h *MenuHandler) GetAll(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	// 获取用户权限码（通过 Service 层，自带 Redis 缓存）
	authService := service.NewAuthService()
	permCodes, err := authService.GetPermissionCodes(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	menus, err := h.menuService.GetUserMenus(userID, permCodes)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, menus)
}

// List 获取所有菜单列表（CRUD 管理）
// @Summary      获取菜单列表
// @Description  获取所有菜单的树形列表（含按钮类型），用于菜单管理页面的 CRUD 操作
// @Tags         菜单管理
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]service.MenuItem} "菜单树"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/menu/list [get]
func (h *MenuHandler) List(c *gin.Context) {
	menus, err := h.menuService.GetAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, menus)
}

// NameExists 检查菜单名称是否存在
// @Summary      检查菜单名称是否存在
// @Description  创建/编辑菜单时校验名称唯一性
// @Tags         菜单管理
// @Produce      json
// @Security     BearerAuth
// @Param        name  query  string  true   "菜单名称"
// @Param        id    query  int     false  "编辑时传入当前菜单 ID，排除自身"
// @Success      200  {object}  response.Response{data=bool} "true=已存在, false=不存在"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/menu/name-exists [get]
func (h *MenuHandler) NameExists(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		response.BadRequest(c, "Name is required")
		return
	}

	var excludeID uint
	if idStr := c.Query("id"); idStr != "" {
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			excludeID = uint(id)
		}
	}

	exists, err := h.menuService.CheckNameExists(name, excludeID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, exists)
}

// PathExists 检查菜单路径是否存在
// @Summary      检查菜单路径是否存在
// @Description  创建/编辑菜单时校验路径唯一性
// @Tags         菜单管理
// @Produce      json
// @Security     BearerAuth
// @Param        path  query  string  true   "菜单路径"
// @Param        id    query  int     false  "编辑时传入当前菜单 ID，排除自身"
// @Success      200  {object}  response.Response{data=bool} "true=已存在, false=不存在"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/menu/path-exists [get]
func (h *MenuHandler) PathExists(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.BadRequest(c, "Path is required")
		return
	}

	var excludeID uint
	if idStr := c.Query("id"); idStr != "" {
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			excludeID = uint(id)
		}
	}

	exists, err := h.menuService.CheckPathExists(path, excludeID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, exists)
}

// Create 创建菜单
// @Summary      创建菜单
// @Description  创建新菜单（支持 catalog/menu/embedded/link/button 类型）
// @Tags         菜单管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.CreateMenuRequest  true  "创建菜单请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/menu [post]
func (h *MenuHandler) Create(c *gin.Context) {
	var req service.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.menuService.Create(&req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Update 更新菜单
// @Summary      更新菜单
// @Description  更新指定菜单信息
// @Tags         菜单管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                       true  "菜单 ID"
// @Param        request  body  service.UpdateMenuRequest  true  "更新菜单请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Failure      404  {object}  response.Response "菜单不存在"
// @Router       /system/menu/{id} [put]
func (h *MenuHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid menu ID")
		return
	}

	var req service.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.menuService.Update(uint(id), &req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除菜单
// @Summary      删除菜单
// @Description  删除指定菜单（软删除）
// @Tags         菜单管理
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "菜单 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Failure      404  {object}  response.Response "菜单不存在"
// @Router       /system/menu/{id} [delete]
func (h *MenuHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid menu ID")
		return
	}

	if err := h.menuService.Delete(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
