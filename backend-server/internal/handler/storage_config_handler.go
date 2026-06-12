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

// maskSensitiveFields 对存储配置中的敏感字段进行脱敏
func maskSensitiveFields(config *model.StorageConfig) {
	config.AccessKey = "******"
	config.SecretKey = "******"
}

// GetAll 获取所有存储配置
// @Summary      获取存储配置列表
// @Description  获取所有存储连接配置
// @Tags         存储配置
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]model.StorageConfig} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-configs [get]
func (h *StorageConfigHandler) GetAll(c *gin.Context) {
	configs, err := h.service.GetAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	for i := range configs {
		maskSensitiveFields(&configs[i])
	}
	response.Success(c, configs)
}

// GetByID 根据ID获取
// @Summary      获取存储配置详情
// @Description  根据 ID 获取存储连接配置详情
// @Tags         存储配置
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "配置 ID"
// @Success      200  {object}  response.Response{data=model.StorageConfig} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-configs/{id} [get]
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
	maskSensitiveFields(config)
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
// @Summary      创建存储配置
// @Description  创建新的存储连接配置
// @Tags         存储配置
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  createStorageConfigRequest  true  "存储配置"
// @Success      200  {object}  response.Response{data=model.StorageConfig} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-configs [post]
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
	if presignedURLExpiry > 604800 {
		presignedURLExpiry = 604800
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
	maskSensitiveFields(config)
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
// @Summary      更新存储配置
// @Description  更新指定存储连接配置
// @Tags         存储配置
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "配置 ID"
// @Param        data  body  updateStorageConfigRequest  true  "存储配置"
// @Success      200  {object}  response.Response{data=model.StorageConfig} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-configs/{id} [put]
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

	// 获取现有配置，用于保留未传字段的原值
	existing, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(c, "存储配置不存在")
		return
	}

	// Status 为 nil 时保留原值
	status := existing.Status
	if req.Status != nil {
		status = *req.Status
	}

	// PresignedURLExpiry 为 nil 或 0 时保留原值
	presignedURLExpiry := existing.PresignedURLExpiry
	if req.PresignedURLExpiry != nil && *req.PresignedURLExpiry > 0 {
		presignedURLExpiry = *req.PresignedURLExpiry
	}
	if presignedURLExpiry > 604800 {
		presignedURLExpiry = 604800
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
	maskSensitiveFields(config)
	response.Success(c, config)
}

// Delete 删除存储配置
// @Summary      删除存储配置
// @Description  删除指定存储连接配置
// @Tags         存储配置
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "配置 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-configs/{id} [delete]
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
// @Summary      设置默认存储配置
// @Description  将指定存储配置设为默认
// @Tags         存储配置
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "配置 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-configs/{id}/default [put]
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
// @Summary      测试存储配置连接
// @Description  测试已有存储配置连接是否可用
// @Tags         存储配置
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "配置 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-configs/{id}/test [post]
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
// @Summary      按数据测试存储连接
// @Description  根据传入配置数据测试存储连接
// @Tags         存储配置
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  testConnectionByDataRequest  true  "测试配置"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-configs/test-by-data [post]
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
// @Summary      获取已启用存储驱动
// @Description  获取已启用的存储驱动列表
// @Tags         存储配置
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-configs/enabled-drivers [get]
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
