package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserRealNameHandler 实名认证处理器
type UserRealNameHandler struct {
	service *service.UserRealNameService
}

// NewUserRealNameHandler 创建实名认证处理器
func NewUserRealNameHandler() *UserRealNameHandler {
	return &UserRealNameHandler{
		service: service.NewUserRealNameService(),
	}
}

// GetStatus 获取当前用户的实名认证状态
// @Summary      获取实名认证状态
// @Description  获取当前用户的实名认证状态
// @Tags         实名认证
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=service.RealNameStatusResponse} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /user/real-name [get]
func (h *UserRealNameHandler) GetStatus(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	status, err := h.service.GetStatus(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, status)
}

// Submit 提交实名认证
// @Summary      提交实名认证
// @Description  提交实名认证申请
// @Tags         实名认证
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.RealNameSubmitRequest  true  "实名认证请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /user/real-name [post]
func (h *UserRealNameHandler) Submit(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req service.RealNameSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.Submit(userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// List 获取实名认证列表（管理员）
// @Summary      获取实名认证列表
// @Description  管理员查看实名认证申请列表
// @Tags         实名认证
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int     false "页码，默认 1"         minimum(1)
// @Param        pageSize   query  int     false "每页条数，默认 20"     minimum(1) maximum(100)
// @Param        status     query  string  false "状态：0=待审, 1=已认证, 2=认证失败"
// @Param        userId     query  string  false "用户 ID"
// @Param        realName   query  string  false "真实姓名（模糊搜索）"
// @Success      200  {object}  response.Response{data=response.PageData{items=[]model.UserRealName}} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/real-name/list [get]
func (h *UserRealNameHandler) List(c *gin.Context) {
	var req service.RealNameListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	list, total, err := h.service.List(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, list, total)
}

// Approve 审核通过
// @Summary      审核通过
// @Description  审核通过用户的实名认证
// @Tags         实名认证
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "实名认证记录 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/real-name/{id}/approve [put]
func (h *UserRealNameHandler) Approve(c *gin.Context) {
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

	if err := h.service.Approve(uint(id), reviewerID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Reject 审核拒绝
// @Summary      审核拒绝
// @Description  拒绝用户的实名认证申请
// @Tags         实名认证
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                               true  "实名认证记录 ID"
// @Param        request  body  service.RealNameRejectRequest     true  "拒绝原因"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/real-name/{id}/reject [put]
func (h *UserRealNameHandler) Reject(c *gin.Context) {
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

	var req service.RealNameRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.Reject(uint(id), reviewerID, req.Reason); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}
