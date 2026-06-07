package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// FileHandler 文件管理处理器
type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler() *FileHandler {
	return &FileHandler{
		fileService: service.NewFileService(),
	}
}

// hasFilePermission 检查用户是否有指定的文件权限
func (h *FileHandler) hasFilePermission(userID uint, permission string) bool {
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

// createFolderRequest 创建文件夹请求
type createFolderRequest struct {
	Name     string `json:"name" binding:"required,max=255"`
	ParentID *uint  `json:"parentId"`
}

// CreateFolder 创建文件夹
// @Summary      创建文件夹
// @Description  创建新文件夹
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Param        body  body  createFolderRequest  true  "文件夹名称"
// @Success      200   {object}  response.Response
// @Router       /files/folder [post]
func (h *FileHandler) CreateFolder(c *gin.Context) {
	var req createFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetCurrentUserID(c)
	folder, err := h.fileService.CreateFolder(userID, req.Name, req.ParentID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, folder)
}

// GetFolderTree 获取目录树
// @Summary      获取目录树
// @Description  返回用户的文件夹目录树
// @Tags         文件管理
// @Produce      json
// @Success      200   {object}  response.Response
// @Router       /files/tree [get]
func (h *FileHandler) GetFolderTree(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	tree, err := h.fileService.GetFolderTree(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, tree)
}

// renameFolderRequest 重命名文件夹请求
type renameFolderRequest struct {
	Name string `json:"name" binding:"required,max=255"`
}

// RenameFolder 重命名文件夹
// @Summary      重命名文件夹
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Param        id    path   int  true  "文件夹ID"
// @Param        body  body   renameFolderRequest  true  "新名称"
// @Success      200   {object}  response.Response
// @Router       /files/folder/{id} [put]
func (h *FileHandler) RenameFolder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req renameFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetCurrentUserID(c)
	if err := h.fileService.RenameFolder(userID, uint(id), req.Name); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteFolder 删除文件夹
// @Summary      删除文件夹
// @Description  递归删除文件夹及其内容
// @Tags         文件管理
// @Produce      json
// @Param        id  path  int  true  "文件夹ID"
// @Success      200   {object}  response.Response
// @Router       /files/folder/{id} [delete]
func (h *FileHandler) DeleteFolder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	userID := middleware.GetCurrentUserID(c)
	if err := h.fileService.DeleteFolder(userID, uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ListFiles 文件列表
// @Summary      文件列表
// @Description  分页获取文件列表，支持搜索和类型过滤
// @Tags         文件管理
// @Produce      json
// @Param        folderId     query  int     false  "文件夹ID"
// @Param        page         query  int     false  "页码"
// @Param        pageSize     query  int     false  "每页数量"
// @Param        keyword      query  string  false  "搜索关键词"
// @Param        contentType  query  string  false  "MIME类型前缀"
// @Success      200   {object}  response.Response
// @Router       /files/list [get]
func (h *FileHandler) ListFiles(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	var req service.ListFilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 检查是否有查看所有文件的权限
	scope := c.DefaultQuery("scope", "own")
	if scope == "all" {
		// 验证权限
		authService := service.NewAuthService()
		permissions, err := authService.GetPermissionCodes(userID)
		if err != nil || !containsPermission(permissions, "file:view:all") {
			response.Forbidden(c, "无权查看所有文件")
			return
		}
		// 传入 0 表示查看所有文件
		userID = 0
	}

	files, total, err := h.fileService.ListFiles(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, files, total)
}

// containsPermission 检查权限列表中是否包含指定权限
func containsPermission(permissions []string, target string) bool {
	for _, p := range permissions {
		if p == target {
			return true
		}
	}
	return false
}

// moveRequest 移动文件请求
type moveRequest struct {
	FileID         uint `json:"fileId" binding:"required"`
	TargetFolderID uint `json:"targetFolderId"`
}

// MoveFile 移动文件
// @Summary      移动文件
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Param        body  body  moveRequest  true  "移动参数"
// @Success      200   {object}  response.Response
// @Router       /files/move [post]
func (h *FileHandler) MoveFile(c *gin.Context) {
	var req moveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetCurrentUserID(c)
	if err := h.fileService.MoveFile(userID, req.FileID, req.TargetFolderID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteFile 删除文件
// @Summary      删除文件
// @Tags         文件管理
// @Produce      json
// @Param        id  path  int  true  "文件ID"
// @Success      200   {object}  response.Response
// @Router       /files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	userID := middleware.GetCurrentUserID(c)
	// 检查是否有删除权限
	hasPermission := h.hasFilePermission(userID, "file:delete") || h.hasFilePermission(userID, "file:manage")

	if err := h.fileService.DeleteFile(userID, uint(id), hasPermission); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// batchDeleteRequest 批量删除请求
type batchDeleteRequest struct {
	FileIDs []uint `json:"fileIds" binding:"required,min=1"`
}

// BatchDeleteFiles 批量删除文件
// @Summary      批量删除文件
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Param        body  body  batchDeleteRequest  true  "文件ID列表"
// @Success      200   {object}  response.Response
// @Router       /files/batch-delete [post]
func (h *FileHandler) BatchDeleteFiles(c *gin.Context) {
	var req batchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetCurrentUserID(c)
	// 检查是否有删除权限
	hasPermission := h.hasFilePermission(userID, "file:delete") || h.hasFilePermission(userID, "file:manage")

	deleted, errors := h.fileService.BatchDeleteFiles(userID, req.FileIDs, hasPermission)

	response.Success(c, gin.H{
		"deleted": deleted,
		"errors":  errors,
	})
}

// batchMoveRequest 批量移动请求
type batchMoveRequest struct {
	FileIDs        []uint `json:"fileIds" binding:"required,min=1"`
	TargetFolderID uint   `json:"targetFolderId"`
}

// BatchMoveFiles 批量移动文件
// @Summary      批量移动文件
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Param        body  body  batchMoveRequest  true  "批量移动参数"
// @Success      200   {object}  response.Response
// @Router       /files/batch-move [post]
func (h *FileHandler) BatchMoveFiles(c *gin.Context) {
	var req batchMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetCurrentUserID(c)
	moved, errors := h.fileService.BatchMoveFiles(userID, req.FileIDs, req.TargetFolderID)

	response.Success(c, gin.H{
		"moved":  moved,
		"errors": errors,
	})
}
