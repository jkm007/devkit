package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"backend-server/internal/middleware"
	"backend-server/internal/repository"
	"backend-server/internal/service"
	"backend-server/pkg/response"
	"backend-server/pkg/storage"

	"github.com/gin-gonic/gin"
)

type ShareHandler struct {
	shareService *service.ShareService
	fileService  *service.FileService
}

// toUint 安全地将 interface{} 转换为 uint（JSON 反序列化数字为 float64）
func toUint(v interface{}) (uint, error) {
	switch val := v.(type) {
	case uint:
		return val, nil
	case float64:
		return uint(val), nil
	case int:
		return uint(val), nil
	case int64:
		return uint(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to uint", v)
	}
}

func NewShareHandler() *ShareHandler {
	return &ShareHandler{
		shareService: service.NewShareService(),
		fileService:  service.NewFileService(),
	}
}

// hasSharePermission 检查用户是否有指定的分享权限
func (h *ShareHandler) hasSharePermission(userID uint, permission string) bool {
	authService := service.NewAuthService()
	codes, err := authService.GetPermissionCodes(userID)
	if err != nil {
		return false
	}
	for _, code := range codes {
		if code == permission {
			return true
		}
	}
	return false
}

// CreateFileShare 创建文件分享
// @Router /files/{id}/share [post]
func (h *ShareHandler) CreateFileShare(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userID := middleware.GetCurrentUserID(c)

	var req struct {
		ExpireHours int    `json:"expireHours"`
		MaxAccess   int    `json:"maxAccess"`
		Password    string `json:"password"` // 可选密码
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	share, err := h.shareService.CreateFileShare(userID, uint(id), req.ExpireHours, req.MaxAccess, req.Password)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"shareCode":   share.ShareCode,
		"shareUrl":    "/share/" + share.ShareCode,
		"expireAt":    share.ExpireAt,
		"hasPassword": share.HasPassword,
	})
}

// CreateFolderShare 创建文件夹分享
// @Router /folders/{id}/share [post]
func (h *ShareHandler) CreateFolderShare(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userID := middleware.GetCurrentUserID(c)

	var req struct {
		ExpireHours int    `json:"expireHours"`
		MaxAccess   int    `json:"maxAccess"`
		Password    string `json:"password"` // 可选密码
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	share, err := h.shareService.CreateFolderShare(userID, uint(id), req.ExpireHours, req.MaxAccess, req.Password)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"shareCode":   share.ShareCode,
		"shareUrl":    "/share/" + share.ShareCode,
		"expireAt":    share.ExpireAt,
		"hasPassword": share.HasPassword,
	})
}

// GetShareInfo 获取分享信息（公开）
// @Router /share/{code} [get]
func (h *ShareHandler) GetShareInfo(c *gin.Context) {
	code := c.Param("code")

	info, err := h.shareService.GetShareInfo(code)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, info)
}

// VerifySharePassword 验证分享密码（公开）
// @Router /share/{code}/verify [post]
func (h *ShareHandler) VerifySharePassword(c *gin.Context) {
	code := c.Param("code")

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入密码")
		return
	}

	ok, err := h.shareService.VerifySharePassword(code, req.Password)
	if err != nil || !ok {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"verified": true})
}

// GetShareFolderFiles 获取分享文件夹内的文件列表（公开，支持分页和搜索）
// @Router /share/{code}/files [get]
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

	folderIDVal, ok := info["folderId"]
	if !ok {
		response.InternalError(c, "分享数据异常")
		return
	}
	folderID, err := toUint(folderIDVal)
	if err != nil {
		response.InternalError(c, "分享数据异常")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取文件夹内的文件列表（分页）
	files, total, err := h.shareService.GetShareFolderFiles(folderID, page, pageSize, keyword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"items": files,
		"total": total,
	})
}

// GetShareFile 获取分享的文件内容（公开，支持 Range 请求）
// 支持：/share/{code}/file (文件分享)
// 支持：/share/{code}/file/{fileId} (文件夹分享中指定文件)
// @Router /share/{code}/file [get]
// @Router /share/{code}/file/{fileId} [get]
func (h *ShareHandler) GetShareFile(c *gin.Context) {
	code := c.Param("code")
	fileIDStr := c.Param("fileId") // 可选，文件夹分享时指定具体文件

	info, err := h.shareService.GetShareInfo(code)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	var objectKey, fileName, contentType, storageType string
	var fileSize int64

	if info["type"] == "file" {
		// 文件分享
		objectKey, _ = info["objectKey"].(string)
		fileName, _ = info["fileName"].(string)
		contentType, _ = info["contentType"].(string)
		fileSize, _ = info["fileSize"].(int64)
		storageType, _ = info["storageType"].(string)
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
		folderIDVal, ok := info["folderId"]
		if !ok {
			response.InternalError(c, "分享数据异常")
			return
		}
		folderID, err := toUint(folderIDVal)
		if err != nil {
			response.InternalError(c, "分享数据异常")
			return
		}
		fileInfo, err := h.shareService.GetFileInFolder(folderID, uint(fileID))
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}

		objectKey, _ = fileInfo["objectKey"].(string)
		fileName, _ = fileInfo["fileName"].(string)
		contentType, _ = fileInfo["contentType"].(string)
		fileSize, _ = fileInfo["fileSize"].(int64)
		storageType, _ = fileInfo["storageType"].(string)
	} else {
		response.BadRequest(c, "无效的分享类型")
		return
	}

	// 根据存储类型获取对应的存储实例
	st := storage.GetStorageByDriver(storageType)

	// 增加访问次数（仅在实际下载文件时计数）
	h.shareService.IncrementAccessCount(code)

	// 设置通用响应头
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", sanitizeContentDisposition("inline", fileName))
	c.Header("Cache-Control", "private, max-age=3600")

	// 处理 Range 请求（用于视频流式播放）
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		// 解析 Range: bytes=start-end
		start, end := parseShareRange(rangeHeader, fileSize)
		if start < 0 || start >= fileSize {
			c.Status(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		if end < 0 || end >= fileSize {
			end = fileSize - 1
		}

		length := end - start + 1

		// 获取文件内容（带偏移）
		reader, err := st.DownloadRange(c, objectKey, start, length)
		if err != nil {
			response.InternalError(c, "获取文件失败")
			return
		}
		defer reader.Close()

		c.Header("Content-Type", contentType)
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		c.Header("Content-Length", strconv.FormatInt(length, 10))
		c.Status(http.StatusPartialContent)

		// 流式返回部分内容
		buf := make([]byte, 32*1024) // 32KB buffer
		io.CopyBuffer(c.Writer, reader, buf)
		return
	}

	// 非 Range 请求，返回完整文件
	reader, err := st.Download(c, objectKey)
	if err != nil {
		response.InternalError(c, "获取文件失败")
		return
	}
	defer reader.Close()

	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(fileSize, 10))

	// 流式返回文件内容
	buf := make([]byte, 32*1024) // 32KB buffer
	io.CopyBuffer(c.Writer, reader, buf)
}

// parseShareRange 解析 Range 头
func parseShareRange(rangeHeader string, fileSize int64) (start, end int64) {
	// Range: bytes=start-end
	if len(rangeHeader) < 7 || rangeHeader[:6] != "bytes=" {
		return -1, -1
	}
	rangeHeader = rangeHeader[6:]

	// 处理 bytes=-suffix (最后 N 个字节)
	if len(rangeHeader) > 0 && rangeHeader[0] == '-' {
		if len(rangeHeader) == 1 {
			// bytes=- 表示整个文件
			return 0, fileSize - 1
		}
		length, err := strconv.ParseInt(rangeHeader[1:], 10, 64)
		if err != nil {
			return -1, -1
		}
		return fileSize - length, fileSize - 1
	}

	// 处理 bytes=start-end 或 bytes=start-
	dashIndex := -1
	for i := 0; i < len(rangeHeader); i++ {
		if rangeHeader[i] == '-' {
			dashIndex = i
			break
		}
	}

	if dashIndex < 0 {
		return -1, -1
	}

	// 解析 start
	startStr := rangeHeader[:dashIndex]
	if startStr == "" {
		return -1, -1
	}
	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		return -1, -1
	}

	// 解析 end（可选）
	endStr := rangeHeader[dashIndex+1:]
	if endStr == "" {
		// bytes=start- 表示从 start 到文件末尾
		return start, fileSize - 1
	}

	end, err = strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		return -1, -1
	}

	return start, end
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
// @Router /shares/{id} [delete]
func (h *ShareHandler) DeleteShare(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userID := middleware.GetCurrentUserID(c)

	// 检查是否有删除权限
	hasPermission := h.hasSharePermission(userID, "share:delete") || h.hasSharePermission(userID, "share:manage")

	err = h.shareService.DeleteShare(userID, uint(id), hasPermission)
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
	scope := c.DefaultQuery("scope", "own")
	keyword := c.Query("keyword")

	// 解析状态筛选参数
	var statusFilter *int
	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil && s >= 1 && s <= 3 {
			statusFilter = &s
		}
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 检查权限：如果有 share:view:all 或 file:view:all 权限且 scope=all，则查看所有分享
	viewAll := false
	if scope == "all" {
		// 通过 Service 层获取用户权限码
		authService := service.NewAuthService()
		codes, err := authService.GetPermissionCodes(userID)
		if err == nil {
			for _, code := range codes {
				if code == "share:view:all" || code == "file:view:all" {
					viewAll = true
					break
				}
			}
		}
	}

	// 构建筛选参数
	filter := &repository.ShareFilterOptions{
		Status:  statusFilter,
		Keyword: keyword,
	}

	items, total, err := h.shareService.GetUserShares(userID, page, pageSize, viewAll, filter)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 获取各状态的数量统计（用于前端状态卡片展示）
	statusCounts, _ := h.shareService.GetShareStatusCounts(userID, viewAll, keyword)

	response.Success(c, gin.H{
		"items":        items,
		"total":        total,
		"statusCounts": statusCounts,
	})
}

// RenewShare 续签分享
// @Router /files/shares/{id}/renew [put]
func (h *ShareHandler) RenewShare(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userID := middleware.GetCurrentUserID(c)

	var req struct {
		ExpireHours int `json:"expireHours" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的续签时长")
		return
	}

	// 检查是否有管理权限
	hasPermission := h.hasSharePermission(userID, "share:manage")

	err = h.shareService.RenewShare(userID, uint(id), req.ExpireHours, hasPermission)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ExpireShare 立即过期分享
// @Router /files/shares/{id}/expire [put]
func (h *ShareHandler) ExpireShare(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userID := middleware.GetCurrentUserID(c)

	// 检查是否有管理权限
	hasPermission := h.hasSharePermission(userID, "share:manage")

	err = h.shareService.ExpireShare(userID, uint(id), hasPermission)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateShareExpiry 修改分享到期时间
// @Router /files/shares/{id}/expiry [put]
func (h *ShareHandler) UpdateShareExpiry(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
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

	// 检查是否有管理权限
	hasPermission := h.hasSharePermission(userID, "share:manage")

	err = h.shareService.UpdateShareExpiry(userID, uint(id), expireAt, hasPermission)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DisableShare 禁用分享
// @Router /files/shares/{id}/disable [put]
func (h *ShareHandler) DisableShare(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userID := middleware.GetCurrentUserID(c)

	// 检查是否有管理权限
	hasPermission := h.hasSharePermission(userID, "share:manage")

	err = h.shareService.DisableShare(userID, uint(id), hasPermission)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// EnableShare 启用分享
// @Router /files/shares/{id}/enable [put]
func (h *ShareHandler) EnableShare(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userID := middleware.GetCurrentUserID(c)

	// 检查是否有管理权限
	hasPermission := h.hasSharePermission(userID, "share:manage")

	err = h.shareService.EnableShare(userID, uint(id), hasPermission)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}
