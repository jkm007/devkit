package handler

import (
	"backend-server/internal/model"
	"backend-server/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": menus})
}

// GetActiveQuickMenus 获取启用的快捷菜单（移动端用）
func (h *MobileConfigHandler) GetActiveQuickMenus(c *gin.Context) {
	menus, err := h.service.GetActiveQuickMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": menus})
}

// CreateQuickMenu 创建快捷菜单
func (h *MobileConfigHandler) CreateQuickMenu(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	// 转换为 model
	menu := &model.QuickMenu{}
	if v, ok := data["title"].(string); ok {
		menu.Title = v
	}
	if v, ok := data["icon"].(string); ok {
		menu.Icon = v
	}
	if v, ok := data["link"].(string); ok {
		menu.Link = v
	}
	if v, ok := data["linkType"].(string); ok {
		menu.LinkType = v
	}
	if v, ok := data["sortOrder"].(float64); ok {
		menu.SortOrder = int(v)
	}
	if v, ok := data["status"].(string); ok {
		menu.Status = v
	}

	if err := h.service.CreateQuickMenu(menu); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": menu, "message": "创建成功"})
}

// UpdateQuickMenu 更新快捷菜单
func (h *MobileConfigHandler) UpdateQuickMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "无效的ID"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	if err := h.service.UpdateQuickMenu(uint(id), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

// DeleteQuickMenu 删除快捷菜单
func (h *MobileConfigHandler) DeleteQuickMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "无效的ID"})
		return
	}

	if err := h.service.DeleteQuickMenu(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ===== 我的页面菜单 =====

// GetMyPageMenus 获取我的页面菜单列表
func (h *MobileConfigHandler) GetMyPageMenus(c *gin.Context) {
	menus, err := h.service.GetMyPageMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": menus})
}

// GetActiveMyPageMenus 获取启用的我的页面菜单（移动端用）
func (h *MobileConfigHandler) GetActiveMyPageMenus(c *gin.Context) {
	menus, err := h.service.GetActiveMyPageMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": menus})
}

// CreateMyPageMenu 创建我的页面菜单
func (h *MobileConfigHandler) CreateMyPageMenu(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	// 转换为 model
	menu := &model.MyPageMenu{}
	if v, ok := data["title"].(string); ok {
		menu.Title = v
	}
	if v, ok := data["icon"].(string); ok {
		menu.Icon = v
	}
	if v, ok := data["link"].(string); ok {
		menu.Link = v
	}
	if v, ok := data["showBadge"].(bool); ok {
		menu.ShowBadge = v
	}
	if v, ok := data["badgeText"].(string); ok {
		menu.BadgeText = v
	}
	if v, ok := data["sortOrder"].(float64); ok {
		menu.SortOrder = int(v)
	}
	if v, ok := data["status"].(string); ok {
		menu.Status = v
	}

	if err := h.service.CreateMyPageMenu(menu); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": menu, "message": "创建成功"})
}

// UpdateMyPageMenu 更新我的页面菜单
func (h *MobileConfigHandler) UpdateMyPageMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "无效的ID"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	if err := h.service.UpdateMyPageMenu(uint(id), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

// DeleteMyPageMenu 删除我的页面菜单
func (h *MobileConfigHandler) DeleteMyPageMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "无效的ID"})
		return
	}

	if err := h.service.DeleteMyPageMenu(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ===== 移动端设置 =====

// GetMobileSettings 获取移动端设置
func (h *MobileConfigHandler) GetMobileSettings(c *gin.Context) {
	settings, err := h.service.GetMobileSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": settings})
}

// UpdateMobileSettings 更新移动端设置
func (h *MobileConfigHandler) UpdateMobileSettings(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	if err := h.service.UpdateMobileSettings(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}
