package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// ClassHandler 班级处理器
type ClassHandler struct {
	classService *service.ClassService
	studyService *service.StudyService
}

// NewClassHandler 创建班级处理器
func NewClassHandler() *ClassHandler {
	return &ClassHandler{
		classService: service.NewClassService(),
		studyService: service.NewStudyService(),
	}
}

// ==================== 管理端接口 ====================

// List 班级列表（管理端）
func (h *ClassHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")

	items, total, err := h.classService.ListClasses(page, pageSize, keyword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, items, total)
}

// GetDetail 班级详情（管理端/用户端共用）
func (h *ClassHandler) GetDetail(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	item, err := h.classService.GetClassDetail(userID, uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, item)
}

// Create 创建班级
func (h *ClassHandler) Create(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	var req service.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	class, err := h.classService.CreateClass(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, class)
}

// Update 更新班级
func (h *ClassHandler) Update(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	var req service.UpdateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.classService.UpdateClass(userID, uint(id), &req); err != nil {
		if err.Error() == "无权操作" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// Delete 删除班级
func (h *ClassHandler) Delete(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	if err := h.classService.DeleteClass(userID, uint(id)); err != nil {
		if err.Error() == "无权操作" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// ListMembers 班级成员列表
func (h *ClassHandler) ListMembers(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	items, total, err := h.classService.ListMembers(userID, uint(id), page, pageSize)
	if err != nil {
		if err.Error() == "无权操作" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, items, total)
}

// AddMember 添加班级成员
func (h *ClassHandler) AddMember(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	classID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	var req struct {
		UserID uint   `json:"userId" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.classService.AddMember(userID, uint(classID), req.UserID, model.ClassMemberRole(req.Role)); err != nil {
		if err.Error() == "无权操作" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "添加成功", nil)
}

// UpdateMemberRole 更新成员角色
func (h *ClassHandler) UpdateMemberRole(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	memberIDStr := c.Param("memberId")
	memberID, err := strconv.ParseUint(memberIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的成员ID")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.classService.UpdateMemberRole(userID, uint(classID), uint(memberID), model.ClassMemberRole(req.Role)); err != nil {
		if err.Error() == "无权操作" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// RemoveMember 移除成员
func (h *ClassHandler) RemoveMember(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	memberIDStr := c.Param("memberId")
	memberID, err := strconv.ParseUint(memberIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的成员ID")
		return
	}

	if err := h.classService.RemoveMember(userID, uint(classID), uint(memberID)); err != nil {
		if err.Error() == "无权操作" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "移除成功", nil)
}

// CreateInvitation 创建邀请码
func (h *ClassHandler) CreateInvitation(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	var req service.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	invitation, err := h.classService.CreateInvitation(userID, uint(classID), &req)
	if err != nil {
		if err.Error() == "无权操作" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, invitation)
}

// DisableInvitation 禁用邀请码
func (h *ClassHandler) DisableInvitation(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的邀请ID")
		return
	}

	if err := h.classService.DisableInvitation(userID, uint(id)); err != nil {
		if err.Error() == "无权操作" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "已禁用", nil)
}

// ListInvitations 邀请码列表
func (h *ClassHandler) ListInvitations(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	items, err := h.classService.ListInvitations(userID, uint(classID))
	if err != nil {
		if err.Error() == "无权操作" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, items)
}

// ==================== 用户端接口 ====================

// ListMyClasses 我加入的班级列表
func (h *ClassHandler) ListMyClasses(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	items, err := h.classService.ListMyClasses(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, items)
}

// JoinByCode 通过邀请码加入班级
func (h *ClassHandler) JoinByCode(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	var req service.JoinByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	class, err := h.classService.JoinByCode(userID, req.Code)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, class)
}

// ListMyClassMembers 班级成员（班级内用户可见）
func (h *ClassHandler) ListMyClassMembers(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	items, total, err := h.classService.ListMembers(userID, uint(id), page, pageSize)
	if err != nil {
		if err.Error() == "无权操作" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, items, total)
}

// ListClassQuestions 班级题目（班级内用户可见）
func (h *ClassHandler) ListClassQuestions(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	userGroupID := middleware.GetCurrentUserGroupID(c)
	userClassIDs := middleware.GetCurrentUserClassIDs(c)

	idStr := c.Param("id")
	classID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	// 校验用户是否在该班级中
	if err := h.classService.CheckClassPermission(userID, uint(classID), model.ClassRoleStudent); err != nil {
		response.Forbidden(c, err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	filters := map[string]interface{}{
		"classId": uint(classID),
	}

	items, total, err := h.studyService.ListQuestions(userID, userGroupID, userClassIDs, page, pageSize, filters)
	if err != nil {
		response.InternalError(c, "获取班级题目失败")
		return
	}

	response.SuccessPage(c, items, total)
}

// Leave 用户主动退出班级
func (h *ClassHandler) Leave(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	classID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的班级ID")
		return
	}

	if err := h.classService.LeaveClass(userID, uint(classID)); err != nil {
		if err.Error() == "无权操作" || err.Error() == "不是班级成员" {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "已退出班级", nil)
}
