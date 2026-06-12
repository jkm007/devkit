package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// RoleApplicationHandler 角色申请处理器
type RoleApplicationHandler struct {
	service *service.RoleApplicationService
}

// NewRoleApplicationHandler 创建角色申请处理器
func NewRoleApplicationHandler() *RoleApplicationHandler {
	return &RoleApplicationHandler{
		service: service.NewRoleApplicationService(),
	}
}

// Create 创建角色申请
// @Summary      创建角色申请
// @Description  用户提交角色申请
// @Tags         角色申请
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.RoleApplicationRequest  true  "角色申请请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/role-applications [post]
func (h *RoleApplicationHandler) Create(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req service.RoleApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.Create(userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetAvailableRoles 获取可申请角色列表
// @Summary      获取可申请角色列表
// @Description  获取当前用户还可以申请的启用角色
// @Tags         角色申请
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]service.AvailableRoleItem} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/role-applications/available-roles [get]
func (h *RoleApplicationHandler) GetAvailableRoles(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	roles, err := h.service.ListAvailableRoles(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, roles)
}

// GetMyList 获取当前用户的申请列表
// @Summary      获取我的角色申请列表
// @Description  获取当前用户的角色申请记录
// @Tags         角色申请
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int  false "页码，默认 1"         minimum(1)
// @Param        pageSize   query  int  false "每页条数，默认 20"     minimum(1) maximum(100)
// @Success      200  {object}  response.Response{data=response.PageData{items=[]service.RoleApplicationItem}} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/role-applications [get]
func (h *RoleApplicationHandler) GetMyList(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, total, err := h.service.ListByUser(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, list, total)
}

// GetAllList 获取所有申请列表（管理员）
// @Summary      获取所有角色申请列表
// @Description  管理员查看所有角色申请
// @Tags         角色申请
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int     false "页码，默认 1"         minimum(1)
// @Param        pageSize   query  int     false "每页条数，默认 20"     minimum(1) maximum(100)
// @Param        status     query  string  false "状态：0=待审, 1=通过, 2=驳回"
// @Param        userId     query  string  false "用户 ID"
// @Param        roleId     query  string  false "角色 ID"
// @Success      200  {object}  response.Response{data=response.PageData{items=[]service.RoleApplicationItem}} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/role-applications/list [get]
func (h *RoleApplicationHandler) GetAllList(c *gin.Context) {
	var req service.RoleApplicationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	list, total, err := h.service.ListAll(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, list, total)
}

// Approve 审核通过
// @Summary      审核通过角色申请
// @Description  管理员审核通过角色申请
// @Tags         角色申请
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                                    true  "申请 ID"
// @Param        request  body  service.RoleApplicationReviewRequest   true  "审核备注"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/role-applications/{id}/approve [put]
func (h *RoleApplicationHandler) Approve(c *gin.Context) {
	reviewerID := middleware.GetCurrentUserID(c)
	if reviewerID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var req service.RoleApplicationReviewRequest
	// note 可选，但 JSON 格式必须正确
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}

	if err := h.service.Approve(uint(id), reviewerID, req.Note); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Reject 审核拒绝
// @Summary      审核拒绝角色申请
// @Description  管理员审核拒绝角色申请
// @Tags         角色申请
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                                    true  "申请 ID"
// @Param        request  body  service.RoleApplicationReviewRequest   true  "审核备注"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/role-applications/{id}/reject [put]
func (h *RoleApplicationHandler) Reject(c *gin.Context) {
	reviewerID := middleware.GetCurrentUserID(c)
	if reviewerID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var req service.RoleApplicationReviewRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}

	if err := h.service.Reject(uint(id), reviewerID, req.Note); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}
