package handler

import (
	"net/http"
	"strconv"

	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// StorageBucketHandler 存储桶处理器
type StorageBucketHandler struct {
	service *service.StorageBucketService
}

// NewStorageBucketHandler 创建存储桶处理器
func NewStorageBucketHandler(service *service.StorageBucketService) *StorageBucketHandler {
	return &StorageBucketHandler{service: service}
}

// GetAll 获取所有存储桶
func (h *StorageBucketHandler) GetAll(c *gin.Context) {
	buckets, err := h.service.GetAll()
	if err != nil {
		response.InternalServerError(c, "获取存储桶列表失败")
		return
	}
	response.Success(c, buckets)
}

// GetByID 根据ID获取存储桶
func (h *StorageBucketHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的存储桶ID")
		return
	}

	bucket, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(c, "存储桶不存在")
		return
	}
	response.Success(c, bucket)
}

// Create 创建存储桶
func (h *StorageBucketHandler) Create(c *gin.Context) {
	var bucket model.StorageBucket
	if err := c.ShouldBindJSON(&bucket); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if bucket.Name == "" {
		response.BadRequest(c, "存储桶名称不能为空")
		return
	}
	if bucket.Driver == "" {
		response.BadRequest(c, "存储驱动不能为空")
		return
	}

	if err := h.service.Create(&bucket); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "创建成功", bucket)
}

// Update 更新存储桶
func (h *StorageBucketHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的存储桶ID")
		return
	}

	var bucket model.StorageBucket
	if err := c.ShouldBindJSON(&bucket); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	bucket.ID = id
	if err := h.service.Update(&bucket); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", bucket)
}

// Delete 删除存储桶
func (h *StorageBucketHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的存储桶ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// SetDefault 设置默认存储桶
func (h *StorageBucketHandler) SetDefault(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的存储桶ID")
		return
	}

	if err := h.service.SetDefault(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "设置成功", nil)
}

// GetByDriver 根据驱动获取存储桶
func (h *StorageBucketHandler) GetByDriver(c *gin.Context) {
	driver := c.Param("driver")
	if driver == "" {
		response.BadRequest(c, "驱动类型不能为空")
		return
	}

	buckets, err := h.service.GetByDriver(driver)
	if err != nil {
		response.InternalServerError(c, "获取存储桶列表失败")
		return
	}
	response.Success(c, buckets)
}

// GetByPurpose 根据用途获取存储桶
func (h *StorageBucketHandler) GetByPurpose(c *gin.Context) {
	purpose := c.Param("purpose")
	if purpose == "" {
		response.BadRequest(c, "用途不能为空")
		return
	}

	buckets, err := h.service.GetByPurpose(purpose)
	if err != nil {
		response.InternalServerError(c, "获取存储桶列表失败")
		return
	}
	response.Success(c, buckets)
}

// GetDefault 获取默认存储桶
func (h *StorageBucketHandler) GetDefault(c *gin.Context) {
	bucket, err := h.service.GetDefault()
	if err != nil {
		response.NotFound(c, "未配置默认存储桶")
		return
	}
	response.Success(c, bucket)
}
