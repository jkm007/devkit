package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
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
		// HLS 流 - 对于云存储返回 presigned URL，本地存储返回 API URL
		driver := storage.GetStorageDriver()
		if driver == "local" {
			// 本地存储：返回流 API URL（需要在路由中添加 HLS 流接口）
			response.Success(c, gin.H{"type": "hls", "url": "/files/" + strconv.FormatUint(id, 10) + "/hls"})
		} else {
			url, err := storage.GetStorage().GetPresignedURL(c, media.HLSPath, 3600)
			if err != nil {
				response.InternalError(c, err.Error())
				return
			}
			response.Success(c, gin.H{"type": "hls", "url": url})
		}
		return
	}

	// 原始文件 - 返回流 API URL（通过 /files/:id/view 访问）
	response.Success(c, gin.H{"type": "original", "url": "/files/" + strconv.FormatUint(id, 10) + "/view"})
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
// @Summary      查看文件
// @Description  返回文件内容用于预览（需认证，inline 显示，支持 Range 请求用于视频流式播放）
// @Tags         媒体文件
// @Produce      octet-stream
// @Param        id  path  int  true  "文件ID"
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

// GetPreviewURL 获取临时预签名 URL（用于视频流式播放）
// @Summary      获取临时预签名 URL
// @Description  返回带临时 token 的 URL，用于视频流式播放
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

	// 生成临时 token（使用 JWT 的短期 token）
	token, err := middleware.GeneratePreviewToken(userID, uint(id))
	if err != nil {
		response.InternalError(c, "生成预览 token 失败")
		return
	}

	// 返回带 token 的 URL
	previewURL := fmt.Sprintf("/files/%d/preview?token=%s", id, token)
	response.Success(c, gin.H{
		"url":         previewURL,
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