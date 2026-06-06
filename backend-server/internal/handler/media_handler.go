package handler

import (
	"strconv"

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
// @Description  返回 HLS 播放地址或原始文件 presigned URL
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

	asset, _, err := h.fileService.GetFileAsset(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	// 检查是否有 HLS 转码
	media, _ := h.mediaService.GetMediaInfo(asset.ID)
	if media != nil && media.TranscodeStatus == "completed" && media.HLSPath != "" {
		url, err := storage.GetStorage().GetPresignedURL(c, media.HLSPath, 3600)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		response.Success(c, gin.H{"type": "hls", "url": url})
		return
	}

	// 原始文件 presigned URL
	url, err := storage.GetStorage().GetPresignedURL(c, asset.ObjectKey, 3600)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"type": "original", "url": url})
}

// DownloadFile 文件下载
// @Summary      下载文件
// @Description  返回文件下载的 presigned URL
// @Tags         媒体文件
// @Produce      json
// @Param        id  path  int  true  "文件ID"
// @Success      200   {object}  response.Response
// @Router       /files/{id}/download [get]
func (h *MediaHandler) DownloadFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	asset, entry, err := h.fileService.GetFileAsset(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	url, err := storage.GetStorage().GetPresignedURL(c, asset.ObjectKey, 3600)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"url":         url,
		"fileName":    entry.Name,
		"contentType": asset.ContentType,
		"size":        asset.FileSize,
	})
}
