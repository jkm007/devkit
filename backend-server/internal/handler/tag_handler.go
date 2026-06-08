package handler

import (
	"strconv"

	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// TagHandler 标签处理器
type TagHandler struct {
	tagService *service.TagService
}

// NewTagHandler 创建标签处理器
func NewTagHandler(tagService *service.TagService) *TagHandler {
	return &TagHandler{tagService: tagService}
}

// GetAllTags 获取所有标签
func (h *TagHandler) GetAllTags(c *gin.Context) {
	tags, err := h.tagService.GetAllTags()
	if err != nil {
		response.InternalError(c, "获取标签失败")
		return
	}
	response.Success(c, tags)
}

// GetGroupedTags 获取按 key 分组的标签
func (h *TagHandler) GetGroupedTags(c *gin.Context) {
	tags, err := h.tagService.GetGroupedTags()
	if err != nil {
		response.InternalError(c, "获取标签失败")
		return
	}
	response.Success(c, tags)
}

// GetTagsByKey 获取指定键的标签值
func (h *TagHandler) GetTagsByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "标签键不能为空")
		return
	}

	tags, err := h.tagService.GetTagsByKey(key)
	if err != nil {
		response.InternalError(c, "获取标签失败")
		return
	}
	response.Success(c, tags)
}

// GetTagByID 根据ID获取标签
func (h *TagHandler) GetTagByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的标签ID")
		return
	}

	tag, err := h.tagService.GetTagByID(id)
	if err != nil {
		response.NotFound(c, "标签不存在")
		return
	}
	response.Success(c, tag)
}

// CreateTag 创建标签
func (h *TagHandler) CreateTag(c *gin.Context) {
	var tag model.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if tag.TagKey == "" || tag.TagValue == "" || tag.TagName == "" {
		response.BadRequest(c, "标签键、值和名称不能为空")
		return
	}

	if err := h.tagService.CreateTag(&tag); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "创建成功", tag)
}

// UpdateTag 更新标签
func (h *TagHandler) UpdateTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的标签ID")
		return
	}

	// 先获取现有标签，保留 created_at
	existing, err := h.tagService.GetTagByID(id)
	if err != nil {
		response.NotFound(c, "标签不存在")
		return
	}

	var tag model.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 保留原始的创建时间
	tag.ID = id
	tag.CreatedAt = existing.CreatedAt
	if err := h.tagService.UpdateTag(&tag); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", tag)
}

// DeleteTag 删除标签
func (h *TagHandler) DeleteTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的标签ID")
		return
	}

	if err := h.tagService.DeleteTag(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetUsageStats 获取标签使用统计
func (h *TagHandler) GetUsageStats(c *gin.Context) {
	stats, err := h.tagService.GetUsageStats()
	if err != nil {
		response.InternalError(c, "获取统计失败")
		return
	}
	response.Success(c, stats)
}

// GetFileTags 获取文件的标签
func (h *TagHandler) GetFileTags(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文件ID")
		return
	}

	tags, err := h.tagService.GetFileTags(uint(fileID))
	if err != nil {
		response.InternalError(c, "获取文件标签失败")
		return
	}
	response.Success(c, tags)
}

// AddFileTag 添加文件标签
func (h *TagHandler) AddFileTag(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文件ID")
		return
	}

	var req struct {
		TagID int64 `json:"tagId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.tagService.AddFileTag(uint(fileID), req.TagID, "manual"); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "添加成功", nil)
}

// RemoveFileTag 移除文件标签
func (h *TagHandler) RemoveFileTag(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文件ID")
		return
	}

	tagIDStr := c.Param("tagId")
	tagID, err := strconv.ParseInt(tagIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的标签ID")
		return
	}

	if err := h.tagService.RemoveFileTag(uint(fileID), tagID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "移除成功", nil)
}

// BatchUpdateFileTags 批量更新文件标签
func (h *TagHandler) BatchUpdateFileTags(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文件ID")
		return
	}

	var req struct {
		TagIDs []int64 `json:"tagIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.tagService.ReplaceFileTags(uint(fileID), req.TagIDs, "manual"); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}
