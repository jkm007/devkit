package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// CategoryFavoriteHandler 分类收藏处理器
type CategoryFavoriteHandler struct {
	service *service.CategoryFavoriteService
}

// NewCategoryFavoriteHandler 创建分类收藏处理器
func NewCategoryFavoriteHandler() *CategoryFavoriteHandler {
	return &CategoryFavoriteHandler{
		service: service.NewCategoryFavoriteService(),
	}
}

// List 获取分类收藏列表
func (h *CategoryFavoriteHandler) List(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	items, total, err := h.service.List(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, "获取分类收藏失败")
		return
	}

	response.SuccessPage(c, items, total)
}

// Add 添加分类收藏
func (h *CategoryFavoriteHandler) Add(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	var req service.CategoryFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	item, err := h.service.Add(userID, &req)
	if err != nil {
		if err.Error() == "已收藏该分类" || err.Error() == "考试大类不存在" ||
			err.Error() == "考试不存在" || err.Error() == "科目不存在" ||
			err.Error() == "章节分类不存在" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, "添加分类收藏失败")
		return
	}

	response.Success(c, item)
}

// Remove 取消分类收藏
func (h *CategoryFavoriteHandler) Remove(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的收藏ID")
		return
	}

	if err := h.service.Remove(userID, uint(id)); err != nil {
		if err.Error() == "收藏记录不存在" {
			response.NotFound(c, "收藏记录不存在")
			return
		}
		response.InternalError(c, "取消收藏失败")
		return
	}

	response.SuccessWithMessage(c, "已取消收藏", nil)
}
