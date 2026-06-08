package handler

import (
	"errors"
	"strconv"

	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(),
	}
}

// List 获取用户列表
// @Summary      获取用户列表
// @Description  获取用户列表数据，支持分页和多条件筛选
// @Tags         用户管理
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int     false "页码，默认 1"           minimum(1)
// @Param        pageSize   query  int     false "每页条数，默认 20"       minimum(1) maximum(100)
// @Param        name       query  string  false "用户名称（模糊搜索）"
// @Param        id         query  string  false "用户 ID（精确匹配）"
// @Param        status     query  string  false "状态：0=禁用, 1=启用"
// @Param        groupId    query  string  false "分组 ID"
// @Param        startTime  query  string  false "创建时间起始（格式：2006-01-02 15:04:05）"
// @Param        endTime    query  string  false "创建时间结束（格式：2006-01-02 15:04:05）"
// @Param        remark     query  string  false "备注（模糊搜索）"
// @Success      200  {object}  response.Response{data=response.PageData{items=[]service.UserResponse}} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/user/list [get]
func (h *UserHandler) List(c *gin.Context) {
	var req service.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	users, total, err := h.userService.List(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, users, total)
}

// Create 创建用户
// @Summary      创建用户
// @Description  创建新用户，可同时指定角色列表。未设置密码时默认为 123456
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.CreateUserRequest  true  "创建用户请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/user [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.Create(&req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Update 更新用户
// @Summary      更新用户
// @Description  更新指定用户信息，可同时更新角色列表
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                       true  "用户 ID"
// @Param        request  body  service.UpdateUserRequest  true  "更新用户请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Failure      404  {object}  response.Response "用户不存在"
// @Router       /system/user/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.Update(uint(id), &req); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "用户不存在")
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// Delete 删除用户
// @Summary      删除用户
// @Description  删除指定用户（软删除）
// @Tags         用户管理
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "用户 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Failure      404  {object}  response.Response "用户不存在"
// @Router       /system/user/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.userService.Delete(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
