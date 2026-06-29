package handler

import (
	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseUintParam(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

type MobileConfigHandler struct {
	service *service.MobileConfigService
}

func NewMobileConfigHandler(service *service.MobileConfigService) *MobileConfigHandler {
	return &MobileConfigHandler{service: service}
}

// ===== 快捷菜单 =====

// GetQuickMenus 获取快捷菜单列表
func (h *MobileConfigHandler) GetQuickMenus(c *gin.Context) {
	menus, err := h.service.GetQuickMenus()
	if err != nil {
		response.InternalError(c, "获取失败")
		return
	}
	response.Success(c, menus)
}

// GetActiveQuickMenus 获取启用的快捷菜单（移动端用）
func (h *MobileConfigHandler) GetActiveQuickMenus(c *gin.Context) {
	menus, err := h.service.GetActiveQuickMenus()
	if err != nil {
		response.InternalError(c, "获取失败")
		return
	}
	response.Success(c, menus)
}

// CreateQuickMenu 创建快捷菜单
func (h *MobileConfigHandler) CreateQuickMenu(c *gin.Context) {
	var req service.QuickMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	menu := &model.QuickMenu{
		Title:     req.Title,
		Icon:      req.Icon,
		Link:      req.Link,
		LinkType:  req.LinkType,
		SortOrder: req.SortOrder,
		Status:    req.Status,
	}

	if err := h.service.CreateQuickMenu(menu); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", menu)
}

// UpdateQuickMenu 更新快捷菜单
func (h *MobileConfigHandler) UpdateQuickMenu(c *gin.Context) {
	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req service.QuickMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.UpdateQuickMenu(id, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteQuickMenu 删除快捷菜单
func (h *MobileConfigHandler) DeleteQuickMenu(c *gin.Context) {
	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.DeleteQuickMenu(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// ===== 我的页面菜单 =====

// GetMyPageMenus 获取我的页面菜单列表
func (h *MobileConfigHandler) GetMyPageMenus(c *gin.Context) {
	menus, err := h.service.GetMyPageMenus()
	if err != nil {
		response.InternalError(c, "获取失败")
		return
	}
	response.Success(c, menus)
}

// GetActiveMyPageMenus 获取启用的我的页面菜单（移动端用）
func (h *MobileConfigHandler) GetActiveMyPageMenus(c *gin.Context) {
	menus, err := h.service.GetActiveMyPageMenus()
	if err != nil {
		response.InternalError(c, "获取失败")
		return
	}
	response.Success(c, menus)
}

// CreateMyPageMenu 创建我的页面菜单
func (h *MobileConfigHandler) CreateMyPageMenu(c *gin.Context) {
	var req service.MyPageMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	menu := &model.MyPageMenu{
		Title:      req.Title,
		Icon:       req.Icon,
		Link:       req.Link,
		LinkType:   req.LinkType,
		ShowBadge:  req.ShowBadge,
		BadgeText:  req.BadgeText,
		SortOrder:  req.SortOrder,
		Status:     req.Status,
	}

	if err := h.service.CreateMyPageMenu(menu); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", menu)
}

// UpdateMyPageMenu 更新我的页面菜单
func (h *MobileConfigHandler) UpdateMyPageMenu(c *gin.Context) {
	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req service.MyPageMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.UpdateMyPageMenu(id, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteMyPageMenu 删除我的页面菜单
func (h *MobileConfigHandler) DeleteMyPageMenu(c *gin.Context) {
	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.DeleteMyPageMenu(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// ===== 移动端设置 =====

// GetMobileSettings 获取移动端设置
func (h *MobileConfigHandler) GetMobileSettings(c *gin.Context) {
	settings, err := h.service.GetMobileSettings()
	if err != nil {
		response.InternalError(c, "获取失败")
		return
	}
	response.Success(c, settings)
}

// UpdateMobileSettings 更新移动端设置
func (h *MobileConfigHandler) UpdateMobileSettings(c *gin.Context) {
	var req service.MobileSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.UpdateMobileSettings(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}
