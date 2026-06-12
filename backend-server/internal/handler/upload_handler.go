package handler

import (
	"fmt"
	"io"
	"strconv"
	"strings"

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
	FolderID uint   `json:"folderId"`
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

	userID := middleware.GetCurrentUserID(c)
	result, err := h.uploadService.CheckUpload(userID, req.FileHash, req.FileSize, req.FolderID)
	if err != nil {
		if isQuotaError(err) {
			response.QuotaExceeded(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
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
	FolderID    uint   `json:"folderId"`
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

	// 文件大小上限检查：10GB
	const maxFileSize int64 = 10 * 1024 * 1024 * 1024
	if req.FileSize <= 0 {
		response.BadRequest(c, "文件大小必须大于0")
		return
	}
	if req.FileSize > maxFileSize {
		response.BadRequest(c, "文件大小超过上限（最大10GB）")
		return
	}

	userID := middleware.GetCurrentUserID(c)
	result, err := h.uploadService.InitUpload(userID, req.FileName, req.FileSize, req.FileHash, req.ContentType, req.TotalParts, req.FolderID)
	if err != nil {
		if isQuotaError(err) {
			response.QuotaExceeded(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, result)
}

// isQuotaError 判断是否为配额超出错误
func isQuotaError(err error) bool {
	return strings.Contains(err.Error(), "存储空间不足")
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

	// 所有权检查
	userID := middleware.GetCurrentUserID(c)
	task, err := h.uploadService.GetTaskByUploadID(uploadID)
	if err != nil {
		response.NotFound(c, "上传任务不存在")
		return
	}
	if task.UserID != userID {
		response.Forbidden(c, "无权操作此上传任务")
		return
	}

	partNumberStr := c.PostForm("partNumber")
	partNumber, err := strconv.Atoi(partNumberStr)
	if err != nil || partNumber < 1 {
		response.BadRequest(c, "partNumber 无效: "+partNumberStr)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "获取分片数据失败: "+err.Error())
		return
	}
	defer file.Close()

	// 获取分片大小
	size := header.Size
	if size <= 0 {
		size = c.Request.ContentLength
	}

	result, err := h.uploadService.UploadPart(uploadID, partNumber, file, size)
	if err != nil {
		response.InternalError(c, "上传分片失败: "+err.Error())
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

	// 所有权检查
	userID := middleware.GetCurrentUserID(c)
	task, err := h.uploadService.GetTaskByUploadID(req.UploadID)
	if err != nil {
		response.NotFound(c, "上传任务不存在")
		return
	}
	if task.UserID != userID {
		response.Forbidden(c, "无权操作此上传任务")
		return
	}

	result, err := h.uploadService.CompleteUpload(req.UploadID)
	if err != nil {
		response.InternalError(c, "合并分片失败")
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

	// 所有权检查
	userID := middleware.GetCurrentUserID(c)
	task, err := h.uploadService.GetTaskByUploadID(req.UploadID)
	if err != nil {
		response.NotFound(c, "上传任务不存在")
		return
	}
	if task.UserID != userID {
		response.Forbidden(c, "无权操作此上传任务")
		return
	}

	if err := h.uploadService.AbortUpload(req.UploadID); err != nil {
		response.InternalError(c, "取消上传失败")
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

	// 所有权检查
	userID := middleware.GetCurrentUserID(c)
	task, err := h.uploadService.GetTaskByUploadID(uploadID)
	if err != nil {
		response.NotFound(c, "上传任务不存在")
		return
	}
	if task.UserID != userID {
		response.Forbidden(c, "无权查看此上传任务")
		return
	}

	result, err := h.uploadService.GetUploadStatus(uploadID)
	if err != nil {
		response.InternalError(c, "获取状态失败")
		return
	}

	response.Success(c, result)
}

// GetUserUploadTasks 获取用户的上传任务列表
// @Summary      获取用户的上传任务列表
// @Description  获取当前用户的上传任务列表，用于显示上传进度
// @Tags         文件上传
// @Produce      json
// @Param        limit  query  int  false  "返回数量限制（默认20）"
// @Success      200   {object}  response.Response
// @Router       /files/upload/tasks [get]
func (h *UploadHandler) GetUserUploadTasks(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	limitStr := c.DefaultQuery("limit", "20")
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		if l > 100 {
			l = 100
		}
		limit = l
	}

	tasks, err := h.uploadService.GetUserUploadTasks(userID, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 转换为响应格式
	result := make([]*service.TaskStatusResponse, len(tasks))
	for i, task := range tasks {
		result[i] = h.uploadService.GetTaskStatusResponse(&task)
	}

	response.Success(c, result)
}

// GetUploadTaskByID 根据ID获取上传任务状态
// @Summary      获取上传任务状态
// @Description  根据任务ID获取上传任务的详细状态
// @Tags         文件上传
// @Produce      json
// @Param        id  path  int  true  "任务ID"
// @Success      200   {object}  response.Response
// @Router       /files/upload/tasks/{id} [get]
func (h *UploadHandler) GetUploadTaskByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	task, err := h.uploadService.GetUploadTaskByID(uint(id))
	if err != nil {
		response.NotFound(c, "任务不存在")
		return
	}

	// 检查权限：只能查看自己的任务
	userID := middleware.GetCurrentUserID(c)
	if task.UserID != userID {
		response.Forbidden(c, "无权查看此任务")
		return
	}

	response.Success(c, h.uploadService.GetTaskStatusResponse(task))
}

// saveToLocal 将 io.Reader 保存到本地临时文件（用于转码等场景）
// TODO: 此函数尚未实现，当前调用方应避免使用，待后续版本完成临时文件存储逻辑
func saveToLocal(reader io.Reader) (string, error) {
	return "", fmt.Errorf("saveToLocal is not implemented: local temporary file storage is not yet available")
}
