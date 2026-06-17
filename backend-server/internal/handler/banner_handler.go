package handler

import (
	"strconv"

	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/logger"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BannerHandler 轮播图处理器
type BannerHandler struct {
	bannerService *service.BannerService
}

// NewBannerHandler 创建轮播图处理器
func NewBannerHandler() *BannerHandler {
	return &BannerHandler{
		bannerService: service.NewBannerService(),
	}
}

// GetBanners 获取启用的轮播图列表（移动端公开接口）
func (h *BannerHandler) GetBanners(c *gin.Context) {
	banners, err := h.bannerService.ListEnabled()
	if err != nil {
		logger.Error("获取轮播图失败", zap.Error(err), zap.String("path", c.Request.URL.Path))
		response.InternalError(c, "获取轮播图失败")
		return
	}

	logger.Info("获取轮播图成功", zap.Int("count", len(banners)))
	response.Success(c, banners)
}

// AdminList 获取所有轮播图（管理端）
func (h *BannerHandler) AdminList(c *gin.Context) {
	banners, err := h.bannerService.ListAll()
	if err != nil {
		response.InternalError(c, "获取轮播图列表失败")
		return
	}

	response.Success(c, banners)
}

// AdminCreate 创建轮播图（管理端）
func (h *BannerHandler) AdminCreate(c *gin.Context) {
	var req struct {
		Title     string `json:"title" binding:"required"`
		Image     string `json:"image" binding:"required"`
		FileID    *uint  `json:"fileId"`
		Link      string `json:"link"`
		LinkType  string `json:"linkType"`
		SortOrder int    `json:"sortOrder"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	banner := &model.Banner{
		Title:     req.Title,
		Image:     req.Image,
		FileID:    req.FileID,
		Link:      req.Link,
		LinkType:  req.LinkType,
		SortOrder: req.SortOrder,
		Status:    model.BannerEnabled,
	}

	if err := h.bannerService.Create(banner); err != nil {
		response.InternalError(c, "创建轮播图失败")
		return
	}

	response.Success(c, banner)
}

// AdminUpdate 更新轮播图（管理端）
func (h *BannerHandler) AdminUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的轮播图ID")
		return
	}

	var req struct {
		Title     string `json:"title"`
		Image     string `json:"image"`
		FileID    *uint  `json:"fileId"`
		Link      string `json:"link"`
		LinkType  string `json:"linkType"`
		SortOrder int    `json:"sortOrder"`
		Status    string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	banner := &model.Banner{ID: uint(id)}
	if req.Title != "" {
		banner.Title = req.Title
	}
	if req.Image != "" {
		banner.Image = req.Image
	}
	banner.FileID = req.FileID
	banner.Link = req.Link
	banner.LinkType = req.LinkType
	if req.SortOrder > 0 {
		banner.SortOrder = req.SortOrder
	}
	if req.Status != "" {
		banner.Status = req.Status
	}

	if err := h.bannerService.Update(banner); err != nil {
		response.InternalError(c, "更新轮播图失败")
		return
	}

	response.Success(c, banner)
}

// AdminDelete 删除轮播图（管理端）
func (h *BannerHandler) AdminDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的轮播图ID")
		return
	}

	if err := h.bannerService.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除轮播图失败")
		return
	}

	response.SuccessWithMessage(c, "轮播图已删除", nil)
}

// AdminUpdateStatus 更新轮播图状态（管理端）
func (h *BannerHandler) AdminUpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的轮播图ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.bannerService.UpdateStatus(uint(id), req.Status); err != nil {
		response.InternalError(c, "更新状态失败")
		return
	}

	response.SuccessWithMessage(c, "状态已更新", nil)
}

// AdminUpdateSort 更新轮播图排序（管理端）
func (h *BannerHandler) AdminUpdateSort(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的轮播图ID")
		return
	}

	var req struct {
		SortOrder int `json:"sortOrder" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.bannerService.UpdateSortOrder(uint(id), req.SortOrder); err != nil {
		response.InternalError(c, "更新排序失败")
		return
	}

	response.SuccessWithMessage(c, "排序已更新", nil)
}
