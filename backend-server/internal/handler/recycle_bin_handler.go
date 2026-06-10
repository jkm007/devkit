package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// RecycleBinHandler 回收站处理器
type RecycleBinHandler struct {
	recycleService *service.RecycleBinService
}

func NewRecycleBinHandler() *RecycleBinHandler {
	return &RecycleBinHandler{
		recycleService: service.NewRecycleBinService(),
	}
}

// List 回收站列表
func (h *RecycleBinHandler) List(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if pageSize > 100 {
		pageSize = 100
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if page < 1 {
		page = 1
	}
	scope := c.DefaultQuery("scope", "own")

	// 检查是否有查看所有文件的权限
	var listUserID uint
	if scope == "all" {
		authService := service.NewAuthService()
		permissions, err := authService.GetPermissionCodes(userID)
		if err != nil || !containsPermission(permissions, "file:view:all") {
			listUserID = userID // 无权限则只看自己的
		} else {
			listUserID = 0 // 0 表示查看所有
		}
	} else {
		listUserID = userID
	}

	items, total, err := h.recycleService.GetRecycleBinList(listUserID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, items, total)
}

// GetCount 获取回收站文件数量
func (h *RecycleBinHandler) GetCount(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	count, err := h.recycleService.GetRecycleBinCount(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"count": count})
}

// Restore 恢复文件
func (h *RecycleBinHandler) Restore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	userID := middleware.GetCurrentUserID(c)
	hasPermission := hasFileDeletePermission(userID)

	if err := h.recycleService.RestoreFile(userID, uint(id), hasPermission); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// BatchRestore 批量恢复
func (h *RecycleBinHandler) BatchRestore(c *gin.Context) {
	var req struct {
		FileIDs []uint `json:"fileIds" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetCurrentUserID(c)
	hasPermission := hasFileDeletePermission(userID)

	restored, errors := h.recycleService.BatchRestoreFiles(userID, req.FileIDs, hasPermission)

	response.Success(c, gin.H{
		"restored": restored,
		"errors":   errors,
	})
}

// PermanentDelete 永久删除
func (h *RecycleBinHandler) PermanentDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	userID := middleware.GetCurrentUserID(c)
	hasPermission := hasFileDeletePermission(userID)

	if err := h.recycleService.PermanentDeleteFile(userID, uint(id), hasPermission); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// BatchPermanentDelete 批量永久删除
func (h *RecycleBinHandler) BatchPermanentDelete(c *gin.Context) {
	var req struct {
		FileIDs []uint `json:"fileIds" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetCurrentUserID(c)
	hasPermission := hasFileDeletePermission(userID)

	deleted, errors := h.recycleService.BatchPermanentDelete(userID, req.FileIDs, hasPermission)

	response.Success(c, gin.H{
		"deleted": deleted,
		"errors":  errors,
	})
}

// Empty 清空回收站
func (h *RecycleBinHandler) Empty(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	if err := h.recycleService.EmptyRecycleBin(userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// hasFileDeletePermission 检查文件删除权限
func hasFileDeletePermission(userID uint) bool {
	authService := service.NewAuthService()
	codes, err := authService.GetPermissionCodes(userID)
	if err != nil {
		return false
	}
	for _, code := range codes {
		if code == "file:delete" || code == "file:manage" || code == "file:recycle:delete" {
			return true
		}
	}
	return false
}
