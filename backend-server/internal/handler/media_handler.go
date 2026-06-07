package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"
	"backend-server/pkg/storage"

	"github.com/gin-gonic/gin"
)

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
	c.Header("Content-Disposition", "attachment; filename="+entry.Name)
	c.Header("Cache-Control", "private, max-age=3600")

	// 流式返回文件内容
	c.DataFromReader(200, asset.FileSize, asset.ContentType, reader, nil)
}

// ViewFile 文件查看（带认证，用于预览）
// @Summary      查看文件
// @Description  返回文件内容用于预览（需认证，inline 显示）
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

	// 下载文件内容
	reader, err := st.Download(c, asset.ObjectKey)
	if err != nil {
		response.InternalError(c, "获取文件失败")
		return
	}
	defer reader.Close()

	// 设置预览响应头（inline 显示而非下载）
	c.Header("Content-Type", asset.ContentType)
	c.Header("Content-Disposition", "inline; filename="+entry.Name)
	c.Header("Cache-Control", "private, max-age=3600")

	// 流式返回文件内容
	c.DataFromReader(200, asset.FileSize, asset.ContentType, reader, nil)
}