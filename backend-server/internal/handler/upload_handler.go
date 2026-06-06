package handler

import (
	"io"
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{
		uploadService: service.NewUploadService(),
	}
}

// checkRequest 秒传检查请求
type checkRequest struct {
	FileHash string `json:"fileHash" binding:"required"`
	FileSize int64  `json:"fileSize" binding:"required"`
}

// CheckUpload 秒传检查
// @Summary      秒传检查
// @Description  通过文件哈希检查是否已存在，实现秒传
// @Tags         文件上传
// @Accept       json
// @Produce      json
// @Param        body  body  checkRequest  true  "文件哈希和大小"
// @Success      200   {object}  response.Response
// @Router       /files/upload/check [post]
func (h *UploadHandler) CheckUpload(c *gin.Context) {
	var req checkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.uploadService.CheckUpload(req.FileHash, req.FileSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// initRequest 初始化上传请求
type initRequest struct {
	FileName    string `json:"fileName" binding:"required"`
	FileSize    int64  `json:"fileSize" binding:"required"`
	FileHash    string `json:"fileHash" binding:"required"`
	ContentType string `json:"contentType"`
	TotalParts  int    `json:"totalParts" binding:"required,min=1"`
}

// InitUpload 初始化分片上传
// @Summary      初始化分片上传
// @Description  创建上传任务，返回 uploadID 和已上传分片列表
// @Tags         文件上传
// @Accept       json
// @Produce      json
// @Param        body  body  initRequest  true  "上传参数"
// @Success      200   {object}  response.Response
// @Router       /files/upload/init [post]
func (h *UploadHandler) InitUpload(c *gin.Context) {
	var req initRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetCurrentUserID(c)
	result, err := h.uploadService.InitUpload(userID, req.FileName, req.FileSize, req.FileHash, req.ContentType, req.TotalParts)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// UploadPart 上传分片
// @Summary      上传分片
// @Description  上传单个文件分片
// @Tags         文件上传
// @Accept       multipart/form-data
// @Produce      json
// @Param        uploadId   formData  string  true  "上传任务ID"
// @Param        partNumber formData  int     true  "分片序号"
// @Param        file       formData  file    true  "分片数据"
// @Success      200   {object}  response.Response
// @Router       /files/upload/part [post]
func (h *UploadHandler) UploadPart(c *gin.Context) {
	uploadID := c.PostForm("uploadId")
	if uploadID == "" {
		response.BadRequest(c, "uploadId 不能为空")
		return
	}

	partNumberStr := c.PostForm("partNumber")
	partNumber, err := strconv.Atoi(partNumberStr)
	if err != nil || partNumber < 1 {
		response.BadRequest(c, "partNumber 无效")
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "获取分片数据失败: "+err.Error())
		return
	}
	defer file.Close()

	// 获取分片大小
	c.Request.ParseMultipartForm(100 << 20) // 100MB max
	size := c.Request.ContentLength

	result, err := h.uploadService.UploadPart(uploadID, partNumber, file, size)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// completeRequest 合并请求
type completeRequest struct {
	UploadID string `json:"uploadId" binding:"required"`
}

// CompleteUpload 合并分片
// @Summary      合并分片
// @Description  合并所有分片完成上传
// @Tags         文件上传
// @Accept       json
// @Produce      json
// @Param        body  body  completeRequest  true  "上传任务ID"
// @Success      200   {object}  response.Response
// @Router       /files/upload/complete [post]
func (h *UploadHandler) CompleteUpload(c *gin.Context) {
	var req completeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.uploadService.CompleteUpload(req.UploadID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// abortRequest 取消请求
type abortRequest struct {
	UploadID string `json:"uploadId" binding:"required"`
}

// AbortUpload 取消上传
// @Summary      取消上传
// @Description  取消分片上传，清理临时文件
// @Tags         文件上传
// @Accept       json
// @Produce      json
// @Param        body  body  abortRequest  true  "上传任务ID"
// @Success      200   {object}  response.Response
// @Router       /files/upload/abort [post]
func (h *UploadHandler) AbortUpload(c *gin.Context) {
	var req abortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.uploadService.AbortUpload(req.UploadID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetUploadStatus 获取上传状态（断点续传）
// @Summary      获取上传状态
// @Description  查询已上传分片列表，用于断点续传
// @Tags         文件上传
// @Produce      json
// @Param        uploadId  query  string  true  "上传任务ID"
// @Success      200   {object}  response.Response
// @Router       /files/upload/status [get]
func (h *UploadHandler) GetUploadStatus(c *gin.Context) {
	uploadID := c.Query("uploadId")
	if uploadID == "" {
		response.BadRequest(c, "uploadId 不能为空")
		return
	}

	result, err := h.uploadService.GetUploadStatus(uploadID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// saveToLocal 将 io.Reader 保存到本地临时文件（用于转码等场景）
func saveToLocal(reader io.Reader) (string, error) {
	// TODO: 实现本地临时文件保存
	return "", nil
}
