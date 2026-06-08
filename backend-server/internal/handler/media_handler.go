package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"backend-server/internal/middleware"
	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/database"
	"backend-server/pkg/response"
	"backend-server/pkg/storage"

	"github.com/gin-gonic/gin"
)

// sanitizeContentDisposition 生成安全的 Content-Disposition 头
func sanitizeContentDisposition(disposition, fileName string) string {
	// 对文件名进行 RFC 5987 编码，防止 HTTP 响应头注入
	escaped := url.PathEscape(fileName)
	return fmt.Sprintf(`%s; filename="%s"; filename*=UTF-8''%s`, disposition, fileName, escaped)
}

// resolveExpiry 解析 expires 参数，返回实际过期时间（秒）
// 如果 expires > 0 则使用用户传入的值，否则使用存储配置的默认值
func resolveExpiry(c *gin.Context, storageType string) int64 {
	// 解析用户传入的 expires 参数
	if expiresStr := c.Query("expires"); expiresStr != "" {
		if expires, err := strconv.ParseInt(expiresStr, 10, 64); err == nil && expires > 0 {
			return expires
		}
	}

	// 使用存储配置的默认值
	if storageType != "local" && storageType != "" {
		var config model.StorageConfig
		db := database.GetMySQL()
		if db != nil {
			if err := db.Where("driver = ? AND is_default = 1 AND status = 1", storageType).First(&config).Error; err == nil && config.PresignedURLExpiry > 0 {
				return int64(config.PresignedURLExpiry)
			}
		}
	}

	return 3600 // fallback 默认1小时
}

// MediaHandler 媒体文件处理器
type MediaHandler struct {
	mediaService *service.MediaService
	fileService  *service.FileService
}

func NewMediaHandler() *MediaHandler {
	return &MediaHandler{
		mediaService: service.NewMediaService(),
		fileService:  service.NewFileService(),
	}
}

// GetMediaInfo 获取媒体元数据
// @Summary      获取媒体信息
// @Description  获取视频/音频的元数据（时长、分辨率、码率等）
// @Tags         媒体文件
// @Produce      json
// @Param        id  path  int  true  "文件ID"
// @Success      200   {object}  response.Response
// @Router       /files/{id}/metadata [get]
func (h *MediaHandler) GetMediaInfo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	asset, _, err := h.fileService.GetFileAsset(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	media, err := h.mediaService.GetMediaInfo(asset.ID)
	if err != nil {
		response.NotFound(c, "暂无媒体信息")
		return
	}

	response.Success(c, media)
}

// GetStream 获取视频流地址
// @Summary      获取视频流
// @Description  返回 HLS 播放地址或原始文件流 API
// @Tags         媒体文件
// @Produce      json
// @Param        id  path  int  true  "文件ID"
// @Success      200   {object}  response.Response
// @Router       /files/{id}/stream [get]
func (h *MediaHandler) GetStream(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	// 获取当前用户ID
	userID := middleware.GetCurrentUserID(c)

	// 验证文件归属
	entry, err := h.fileService.GetFileEntry(userID, uint(id))
	if err != nil {
		response.NotFound(c, "文件不存在或无权访问")
		return
	}

	// 获取文件资产
	asset, err := h.fileService.GetAssetByID(entry.FileAssetID)
	if err != nil {
		response.NotFound(c, "文件资产不存在")
		return
	}

	// 检查是否有 HLS 转码
	media, _ := h.mediaService.GetMediaInfo(asset.ID)
	if media != nil && media.TranscodeStatus == "completed" && media.HLSPath != "" {
		// HLS 流 - 使用文件资产的实际存储类型判断
		if asset.StorageType == "local" {
			// 本地存储：返回代理 URL
			response.Success(c, gin.H{"type": "hls", "url": "/files/" + strconv.FormatUint(id, 10) + "/hls"})
		} else {
			// 云存储：返回 presigned URL
			st := storage.GetStorageByDriver(asset.StorageType)
			expires := resolveExpiry(c, asset.StorageType)
			url, err := st.GetPresignedURL(c, media.HLSPath, expires)
			if err != nil {
				response.InternalError(c, err.Error())
				return
			}
			response.Success(c, gin.H{"type": "hls", "url": url})
		}
		return
	}

	// 原始文件 - 根据存储类型返回合适的 URL
	if asset.StorageType == "local" {
		// 本地存储：返回代理 URL
		response.Success(c, gin.H{"type": "original", "url": "/files/" + strconv.FormatUint(id, 10) + "/view"})
	} else {
		// 云存储：返回 direct-url 接口（前端获取 presigned URL）
		response.Success(c, gin.H{"type": "original", "url": "/files/" + strconv.FormatUint(id, 10) + "/direct-url"})
	}
}

// DownloadFile 文件下载（直接返回文件内容）
// @Summary      下载文件
// @Description  直接返回文件内容（需认证）
// @Tags         媒体文件
// @Produce      octet-stream
// @Param        id  path  int  true  "文件ID"
// @Success      200   {file}  binary
// @Router       /files/{id}/download [get]
func (h *MediaHandler) DownloadFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	// 获取当前用户ID
	userID := middleware.GetCurrentUserID(c)

	// 验证文件归属
	entry, err := h.fileService.GetFileEntry(userID, uint(id))
	if err != nil {
		response.NotFound(c, "文件不存在或无权访问")
		return
	}

	// 获取文件资产
	asset, err := h.fileService.GetAssetByID(entry.FileAssetID)
	if err != nil {
		response.NotFound(c, "文件资产不存在")
		return
	}

	// 根据存储类型获取对应的存储实例
	st := storage.GetStorageByDriver(asset.StorageType)

	// 下载文件内容
	reader, err := st.Download(c, asset.ObjectKey)
	if err != nil {
		response.InternalError(c, "获取文件失败")
		return
	}
	defer reader.Close()

	// 设置下载响应头
	c.Header("Content-Type", asset.ContentType)
	c.Header("Content-Disposition", sanitizeContentDisposition("attachment", entry.Name))
	c.Header("Cache-Control", "private, max-age=3600")

	// 流式返回文件内容
	c.DataFromReader(200, asset.FileSize, asset.ContentType, reader, nil)
}

// ViewFile 文件查看（带认证，用于预览，支持 Range 请求）
// 支持 ?presigned=true 参数：云存储文件 302 重定向到 presigned URL，节省服务器带宽
// @Summary      查看文件
// @Description  返回文件内容用于预览（需认证，inline 显示，支持 Range 请求用于视频流式播放）
// @Tags         媒体文件
// @Produce      octet-stream
// @Param        id         path   int     true   "文件ID"
// @Param        presigned  query  string  false  "云存储使用 presigned URL 重定向 (true)"
// @Success      200   {file}  binary
// @Router       /files/{id}/view [get]
func (h *MediaHandler) ViewFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	// 获取当前用户ID
	userID := middleware.GetCurrentUserID(c)

	// 验证文件归属
	entry, err := h.fileService.GetFileEntry(userID, uint(id))
	if err != nil {
		response.NotFound(c, "文件不存在或无权访问")
		return
	}

	// 获取文件资产
	asset, err := h.fileService.GetAssetByID(entry.FileAssetID)
	if err != nil {
		response.NotFound(c, "文件资产不存在")
		return
	}

	// 根据存储类型获取对应的存储实例
	st := storage.GetStorageByDriver(asset.StorageType)

	// presigned 模式：云存储 302 重定向，节省服务器带宽
	if c.Query("presigned") == "true" && asset.StorageType != "local" {
		expires := resolveExpiry(c, asset.StorageType)
		presignedURL, err := st.GetPresignedURL(c, asset.ObjectKey, expires)
		if err != nil {
			// presigned 失败，回退到代理模式
		} else {
			c.Redirect(http.StatusFound, presignedURL)
			return
		}
	}

	fileSize := asset.FileSize
	contentType := asset.ContentType

	// 设置通用响应头
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", sanitizeContentDisposition("inline", entry.Name))
	c.Header("Cache-Control", "private, max-age=3600")

	// 处理 Range 请求（用于视频流式播放和大文件分块下载）
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		// 解析 Range: bytes=start-end
		start, end := parseRange(rangeHeader, fileSize)
		if start < 0 || start >= fileSize {
			c.Status(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		if end < 0 || end >= fileSize {
			end = fileSize - 1
		}

		length := end - start + 1

		// 获取文件内容（带偏移）
		reader, err := st.DownloadRange(c, asset.ObjectKey, start, length)
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
	reader, err := st.Download(c, asset.ObjectKey)
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

// GetDirectURL 获取文件直链（presigned URL，仅云存储）
// 云存储返回 presigned URL（有效期1小时），本地存储返回代理 URL
// @Summary      获取文件直链
// @Description  云存储返回 presigned URL，本地存储返回代理 URL
// @Tags         媒体文件
// @Produce      json
// @Param        id  path  int  true  "文件ID"
// @Success      200  {object}  response.Response
// @Router       /files/{id}/direct-url [get]
func (h *MediaHandler) GetDirectURL(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	// 获取当前用户ID
	userID := middleware.GetCurrentUserID(c)

	// 验证文件归属
	entry, err := h.fileService.GetFileEntry(userID, uint(id))
	if err != nil {
		response.NotFound(c, "文件不存在或无权访问")
		return
	}

	// 获取文件资产
	asset, err := h.fileService.GetAssetByID(entry.FileAssetID)
	if err != nil {
		response.NotFound(c, "文件资产不存在")
		return
	}

	// 根据存储类型决定 URL 策略
	st := storage.GetStorageByDriver(asset.StorageType)
	expires := resolveExpiry(c, asset.StorageType)

	if asset.StorageType == "local" {
		// 本地存储：返回代理 URL
		response.Success(c, gin.H{
			"url":         fmt.Sprintf("/files/%d/view", id),
			"strategy":    "proxy",
			"contentType": entry.ContentType,
			"name":        entry.Name,
		})
	} else {
		// 云存储：返回 presigned URL
		presignedURL, err := st.GetPresignedURL(c, asset.ObjectKey, expires)
		if err != nil {
			// presigned 失败，回退到代理 URL
			response.Success(c, gin.H{
				"url":         fmt.Sprintf("/files/%d/view", id),
				"strategy":    "proxy",
				"contentType": entry.ContentType,
				"name":        entry.Name,
			})
		} else {
			response.Success(c, gin.H{
				"url":         presignedURL,
				"strategy":    "presigned",
				"expiresIn":   expires,
				"contentType": entry.ContentType,
				"name":        entry.Name,
			})
		}
	}
}

// GetPreviewURL 获取临时预签名 URL（用于视频流式播放）
// 云存储直接返回 presigned URL，本地存储返回 JWT token URL
// @Summary      获取临时预签名 URL
// @Description  云存储返回 presigned URL，本地存储返回 JWT token URL
// @Tags         媒体文件
// @Produce      json
// @Param        id  path  int  true  "文件ID"
// @Success      200  {object}  response.Response
// @Router       /files/{id}/preview-url [get]
func (h *MediaHandler) GetPreviewURL(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	// 获取当前用户ID
	userID := middleware.GetCurrentUserID(c)

	// 验证文件归属
	entry, err := h.fileService.GetFileEntry(userID, uint(id))
	if err != nil {
		response.NotFound(c, "文件不存在或无权访问")
		return
	}

	// 获取文件资产
	asset, err := h.fileService.GetAssetByID(entry.FileAssetID)
	if err != nil {
		response.NotFound(c, "文件资产不存在")
		return
	}

	// 云存储：直接返回 presigned URL（前端直接访问云存储，零带宽）
	if asset.StorageType != "local" {
		st := storage.GetStorageByDriver(asset.StorageType)
		expires := resolveExpiry(c, asset.StorageType)
		presignedURL, err := st.GetPresignedURL(c, asset.ObjectKey, expires)
		if err == nil {
			response.Success(c, gin.H{
				"url":         presignedURL,
				"strategy":    "presigned",
				"expiresIn":   expires,
				"contentType": entry.ContentType,
				"name":        entry.Name,
			})
			return
		}
		// presigned 失败，回退到 token 方式
	}

	// 本地存储：返回 JWT token URL（通过后端代理）
	token, err := middleware.GeneratePreviewToken(userID, uint(id))
	if err != nil {
		response.InternalError(c, "生成预览 token 失败")
		return
	}

	previewURL := fmt.Sprintf("/files/%d/preview?token=%s", id, token)
	response.Success(c, gin.H{
		"url":         previewURL,
		"strategy":    "proxy",
		"contentType": entry.ContentType,
		"name":        entry.Name,
	})
}

// PreviewFile 临时预览文件（使用 token 认证）
// @Summary      临时预览文件
// @Description  使用临时 token 预览文件（支持 Range 请求）
// @Tags         媒体文件
// @Produce      octet-stream
// @Param        id     path  int     true  "文件ID"
// @Param        token  query string  true  "临时 token"
// @Success      200   {file}  binary
// @Router       /files/{id}/preview [get]
func (h *MediaHandler) PreviewFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "缺少 token")
		return
	}

	// 验证临时 token
	userID, fileID, err := middleware.ValidatePreviewToken(token)
	if err != nil || fileID != uint(id) {
		response.Unauthorized(c, "无效或过期的 token")
		return
	}

	// 获取文件信息
	entry, err := h.fileService.GetFileEntry(userID, uint(id))
	if err != nil {
		response.NotFound(c, "文件不存在或无权访问")
		return
	}

	// 获取文件资产
	asset, err := h.fileService.GetAssetByID(entry.FileAssetID)
	if err != nil {
		response.NotFound(c, "文件资产不存在")
		return
	}

	// 根据存储类型获取对应的存储实例
	st := storage.GetStorageByDriver(asset.StorageType)

	// 云存储：302 重定向到 presigned URL，节省服务器带宽
	if asset.StorageType != "local" {
		expires := resolveExpiry(c, asset.StorageType)
		presignedURL, err := st.GetPresignedURL(c, asset.ObjectKey, expires)
		if err == nil {
			c.Redirect(http.StatusFound, presignedURL)
			return
		}
		// presigned 失败，回退到代理模式
	}

	fileSize := asset.FileSize
	contentType := asset.ContentType

	// 设置通用响应头
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", sanitizeContentDisposition("inline", entry.Name))
	c.Header("Cache-Control", "private, max-age=3600")

	// 处理 Range 请求
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		start, end := parseRange(rangeHeader, fileSize)
		if start < 0 || start >= fileSize {
			c.Status(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		if end < 0 || end >= fileSize {
			end = fileSize - 1
		}

		length := end - start + 1

		reader, err := st.DownloadRange(c, asset.ObjectKey, start, length)
		if err != nil {
			response.InternalError(c, "获取文件失败")
			return
		}
		defer reader.Close()

		c.Header("Content-Type", contentType)
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		c.Header("Content-Length", strconv.FormatInt(length, 10))
		c.Status(http.StatusPartialContent)

		buf := make([]byte, 32*1024)
		io.CopyBuffer(c.Writer, reader, buf)
		return
	}

	// 非 Range 请求
	reader, err := st.Download(c, asset.ObjectKey)
	if err != nil {
		response.InternalError(c, "获取文件失败")
		return
	}
	defer reader.Close()

	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(fileSize, 10))

	buf := make([]byte, 32*1024)
	io.CopyBuffer(c.Writer, reader, buf)
}

// parseRange 解析 Range 头，返回 start 和 end
func parseRange(rangeHeader string, fileSize int64) (start, end int64) {
	// 格式: bytes=start-end 或 bytes=start-
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return -1, -1
	}
	rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.SplitN(rangeSpec, "-", 2)
	if len(parts) != 2 {
		return -1, -1
	}

	if parts[0] == "" {
		// bytes=-N (从末尾算)
		n, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return -1, -1
		}
		return fileSize - n, fileSize - 1
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return -1, -1
	}

	if parts[1] == "" {
		// bytes=start-
		return start, fileSize - 1
	}

	end, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return -1, -1
	}

	return start, end
}