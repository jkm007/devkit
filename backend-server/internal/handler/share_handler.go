package handler

import (
	"strconv"
	"time"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"
	"backend-server/pkg/storage"

	"github.com/gin-gonic/gin"
)

type ShareHandler struct {
	shareService *service.ShareService
	fileService  *service.FileService
}

func NewShareHandler() *ShareHandler {
	return &ShareHandler{
		shareService: service.NewShareService(),
		fileService:  service.NewFileService(),
	}
}

// CreateFileShare 创建文件分享
// @Router /files/:id/share [post]
func (h *ShareHandler) CreateFileShare(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := middleware.GetCurrentUserID(c)

	var req struct {
		ExpireHours int `json:"expireHours"`
		MaxAccess   int `json:"maxAccess"`
	}
	c.ShouldBindJSON(&req)

	share, err := h.shareService.CreateFileShare(userID, uint(id), req.ExpireHours, req.MaxAccess)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"shareCode": share.ShareCode,
		"shareUrl":  "/share/" + share.ShareCode,
		"expireAt":  share.ExpireAt,
	})
}

// CreateFolderShare 创建文件夹分享
// @Router /folders/:id/share [post]
func (h *ShareHandler) CreateFolderShare(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := middleware.GetCurrentUserID(c)

	var req struct {
		ExpireHours int `json:"expireHours"`
		MaxAccess   int `json:"maxAccess"`
	}
	c.ShouldBindJSON(&req)

	share, err := h.shareService.CreateFolderShare(userID, uint(id), req.ExpireHours, req.MaxAccess)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"shareCode": share.ShareCode,
		"shareUrl":  "/share/" + share.ShareCode,
		"expireAt":  share.ExpireAt,
	})
}

// GetShareInfo 获取分享信息（公开）
// @Router /share/:code [get]
func (h *ShareHandler) GetShareInfo(c *gin.Context) {
	code := c.Param("code")

	info, err := h.shareService.GetShareInfo(code)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, info)
}

// GetShareFolderFiles 获取分享文件夹内的文件列表（公开）
// @Router /share/:code/files [get]
func (h *ShareHandler) GetShareFolderFiles(c *gin.Context) {
	code := c.Param("code")

	info, err := h.shareService.GetShareInfo(code)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	if info["type"] != "folder" {
		response.BadRequest(c, "不是文件夹分享")
		return
	}

	folderID := uint(info["folderId"].(uint))

	// 获取文件夹内的文件列表
	files, err := h.shareService.GetShareFolderFiles(folderID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, files)
}

// GetShareFile 获取分享的文件内容（公开）
// 支持：/share/:code/file (文件分享)
// 支持：/share/:code/file/:fileId (文件夹分享中指定文件)
// @Router /share/:code/file [get]
// @Router /share/:code/file/:fileId [get]
func (h *ShareHandler) GetShareFile(c *gin.Context) {
	code := c.Param("code")
	fileIDStr := c.Param("fileId") // 可选，文件夹分享时指定具体文件

	info, err := h.shareService.GetShareInfo(code)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	var objectKey, fileName, contentType string
	var fileSize int64

	if info["type"] == "file" {
		// 文件分享
		objectKey = info["objectKey"].(string)
		fileName = info["fileName"].(string)
		contentType = info["contentType"].(string)
		fileSize = info["fileSize"].(int64)
	} else if info["type"] == "folder" {
		// 文件夹分享 - 需要指定 fileId
		if fileIDStr == "" {
			response.BadRequest(c, "文件夹分享需要指定文件ID")
			return
		}
		fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "无效的文件ID")
			return
		}

		// 验证文件属于分享的文件夹
		folderID := uint(info["folderId"].(uint))
		fileInfo, err := h.shareService.GetFileInFolder(folderID, uint(fileID))
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}

		objectKey = fileInfo["objectKey"].(string)
		fileName = fileInfo["fileName"].(string)
		contentType = fileInfo["contentType"].(string)
		fileSize = fileInfo["fileSize"].(int64)
	} else {
		response.BadRequest(c, "无效的分享类型")
		return
	}

	// 下载文件
	reader, err := storage.GetStorage().Download(c, objectKey)
	if err != nil {
		response.InternalError(c, "获取文件失败")
		return
	}
	defer reader.Close()

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename="+fileName)
	c.DataFromReader(200, fileSize, contentType, reader, nil)
}

// GetMyShares 获取我的分享列表
// @Router /my-shares [get]
func (h *ShareHandler) GetMyShares(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	shares, err := h.shareService.GetMyShares(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, shares)
}

// DeleteShare 删除分享
// @Router /shares/:id [delete]
func (h *ShareHandler) DeleteShare(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := middleware.GetCurrentUserID(c)

	err := h.shareService.DeleteShare(userID, uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetUserShares 获取用户分享列表（带分页）
// @Router /files/shares [get]
func (h *ShareHandler) GetUserShares(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.shareService.GetUserShares(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"items": items,
		"total": total,
	})
}

// RenewShare 续签分享
// @Router /files/shares/:id/renew [put]
func (h *ShareHandler) RenewShare(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := middleware.GetCurrentUserID(c)

	var req struct {
		ExpireHours int `json:"expireHours" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的续签时长")
		return
	}

	err := h.shareService.RenewShare(userID, uint(id), req.ExpireHours)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ExpireShare 立即过期分享
// @Router /files/shares/:id/expire [put]
func (h *ShareHandler) ExpireShare(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := middleware.GetCurrentUserID(c)

	err := h.shareService.ExpireShare(userID, uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateShareExpiry 修改分享到期时间
// @Router /files/shares/:id/expiry [put]
func (h *ShareHandler) UpdateShareExpiry(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := middleware.GetCurrentUserID(c)

	var req struct {
		ExpireAt string `json:"expireAt"` // ISO 8601 格式，空字符串表示永不过期
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	var expireAt *time.Time
	if req.ExpireAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpireAt)
		if err != nil {
			response.BadRequest(c, "无效的时间格式")
			return
		}
		expireAt = &t
	}

	err := h.shareService.UpdateShareExpiry(userID, uint(id), expireAt)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DisableShare 禁用分享
// @Router /files/shares/:id/disable [put]
func (h *ShareHandler) DisableShare(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := middleware.GetCurrentUserID(c)

	err := h.shareService.DisableShare(userID, uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// EnableShare 启用分享
// @Router /files/shares/:id/enable [put]
func (h *ShareHandler) EnableShare(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := middleware.GetCurrentUserID(c)

	err := h.shareService.EnableShare(userID, uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}