package handler

import (
	"strconv"
	"strings"

	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// StorageConfigHandler 存储连接配置处理器
type StorageConfigHandler struct {
	service *service.StorageConfigService
}

func NewStorageConfigHandler() *StorageConfigHandler {
	return &StorageConfigHandler{
		service: service.NewStorageConfigService(),
	}
}

// GetAll 获取所有存储配置
func (h *StorageConfigHandler) GetAll(c *gin.Context) {
	configs, err := h.service.GetAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, configs)
}

// GetByID 根据ID获取
func (h *StorageConfigHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	config, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(c, "存储配置不存在")
		return
	}
	response.Success(c, config)
}

// createStorageConfigRequest 创建请求
type createStorageConfigRequest struct {
	Name               string `json:"name" binding:"required,max=100"`
	Driver             string `json:"driver" binding:"required,oneof=local minio oss cos"`
	Endpoint           string `json:"endpoint"`
	AccessKey          string `json:"accessKey"`
	SecretKey          string `json:"secretKey"`
	Bucket             string `json:"bucket"`
	Region             string `json:"region"`
	UseSSL             bool   `json:"useSsl"`
	CDNDomain          string `json:"cdnDomain"`
	IsDefault          bool   `json:"isDefault"`
	PresignedURLExpiry *int   `json:"presignedUrlExpiry"`
	Status             *int8  `json:"status"`
	Description        string `json:"description"`
}

// Create 创建存储配置
func (h *StorageConfigHandler) Create(c *gin.Context) {
	var req createStorageConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}

	presignedURLExpiry := 3600
	if req.PresignedURLExpiry != nil && *req.PresignedURLExpiry > 0 {
		presignedURLExpiry = *req.PresignedURLExpiry
	}

	config := &model.StorageConfig{
		Name:               req.Name,
		Driver:             req.Driver,
		Endpoint:           req.Endpoint,
		AccessKey:          req.AccessKey,
		SecretKey:          req.SecretKey,
		Bucket:             req.Bucket,
		Region:             req.Region,
		UseSSL:             req.UseSSL,
		CDNDomain:          req.CDNDomain,
		IsDefault:          req.IsDefault,
		PresignedURLExpiry: presignedURLExpiry,
		Status:             status,
		Description:        req.Description,
	}

	if err := h.service.Create(config); err != nil {
		// 区分业务错误和系统错误
		if isBusinessError(err) {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}
	response.Success(c, config)
}

// updateStorageConfigRequest 更新请求
type updateStorageConfigRequest struct {
	Name               string `json:"name" binding:"required,max=100"`
	Driver             string `json:"driver" binding:"required,oneof=local minio oss cos"`
	Endpoint           string `json:"endpoint"`
	AccessKey          string `json:"accessKey"`
	SecretKey          string `json:"secretKey"`
	Bucket             string `json:"bucket"`
	Region             string `json:"region"`
	UseSSL             bool   `json:"useSsl"`
	CDNDomain          string `json:"cdnDomain"`
	IsDefault          bool   `json:"isDefault"`
	PresignedURLExpiry *int   `json:"presignedUrlExpiry"`
	Status             *int8  `json:"status"`
	Description        string `json:"description"`
}

// Update 更新存储配置
func (h *StorageConfigHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req updateStorageConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Status 为 nil 时保留原值（不默认为启用）
	var status int8
	if req.Status != nil {
		status = *req.Status
	} else {
		// 获取现有配置的 status
		existing, err := h.service.GetByID(id)
		if err != nil {
			response.NotFound(c, "存储配置不存在")
			return
		}
		status = existing.Status
	}

	presignedURLExpiry := 3600
	if req.PresignedURLExpiry != nil && *req.PresignedURLExpiry > 0 {
		presignedURLExpiry = *req.PresignedURLExpiry
	}

	config := &model.StorageConfig{
		ID:                 id,
		Name:               req.Name,
		Driver:             req.Driver,
		Endpoint:           req.Endpoint,
		AccessKey:          req.AccessKey,
		SecretKey:          req.SecretKey,
		Bucket:             req.Bucket,
		Region:             req.Region,
		UseSSL:             req.UseSSL,
		CDNDomain:          req.CDNDomain,
		IsDefault:          req.IsDefault,
		PresignedURLExpiry: presignedURLExpiry,
		Status:             status,
		Description:        req.Description,
	}

	if err := h.service.Update(config); err != nil {
		if isBusinessError(err) {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}
	response.Success(c, config)
}

// Delete 删除存储配置
func (h *StorageConfigHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if isBusinessError(err) {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}
	response.Success(c, nil)
}

// SetDefault 设置默认
func (h *StorageConfigHandler) SetDefault(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.SetDefault(id); err != nil {
		if isBusinessError(err) {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}
	response.Success(c, nil)
}

// TestConnection 测试已有配置的连接
func (h *StorageConfigHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.service.TestConnection(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"connected": true})
}

// testConnectionByDataRequest 根据数据测试连接
type testConnectionByDataRequest struct {
	Driver    string `json:"driver" binding:"required"`
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	UseSSL    bool   `json:"useSsl"`
}

// TestConnectionByData 根据传入数据测试连接
func (h *StorageConfigHandler) TestConnectionByData(c *gin.Context) {
	var req testConnectionByDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.TestConnectionByData(req.Driver, req.Endpoint, req.AccessKey, req.SecretKey, req.Bucket, req.Region, req.UseSSL); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"connected": true})
}

// GetEnabledDrivers 获取已启用的驱动
func (h *StorageConfigHandler) GetEnabledDrivers(c *gin.Context) {
	drivers, err := h.service.GetEnabledDrivers()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, drivers)
}

// isBusinessError 判断是否为业务逻辑错误（应返回 400 而非 500）
func isBusinessError(err error) bool {
	msg := err.Error()
	businessKeywords := []string{
		"已存在", "不存在", "不允许", "不能", "禁用",
		"缺少", "不能为空", "无效",
	}
	for _, kw := range businessKeywords {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}
