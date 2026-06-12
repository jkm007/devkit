package handler

import (
	"strconv"

	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// maskBucketSensitiveFields 对存储桶中的敏感字段进行脱敏
func maskBucketSensitiveFields(bucket *model.StorageBucket) {
	bucket.AccessKey = "******"
	bucket.SecretKey = "******"
}

// StorageBucketHandler 存储桶处理器
type StorageBucketHandler struct {
	service *service.StorageBucketService
}

// NewStorageBucketHandler 创建存储桶处理器
func NewStorageBucketHandler(service *service.StorageBucketService) *StorageBucketHandler {
	return &StorageBucketHandler{service: service}
}

// GetAll 获取所有存储桶
// @Summary      获取存储桶列表
// @Description  获取所有存储桶配置
// @Tags         存储桶
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]model.StorageBucket} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets [get]
func (h *StorageBucketHandler) GetAll(c *gin.Context) {
	buckets, err := h.service.GetAll()
	if err != nil {
		response.InternalError(c, "获取存储桶列表失败")
		return
	}
	for i := range buckets {
		maskBucketSensitiveFields(&buckets[i])
	}
	response.Success(c, buckets)
}

// GetByID 根据ID获取存储桶
// @Summary      获取存储桶详情
// @Description  根据 ID 获取存储桶配置详情
// @Tags         存储桶
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "存储桶 ID"
// @Success      200  {object}  response.Response{data=model.StorageBucket} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets/{id} [get]
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
	maskBucketSensitiveFields(bucket)
	response.Success(c, bucket)
}

// Create 创建存储桶
// @Summary      创建存储桶
// @Description  创建新的存储桶配置
// @Tags         存储桶
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  model.StorageBucket  true  "存储桶配置"
// @Success      200  {object}  response.Response{data=model.StorageBucket} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets [post]
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

	maskBucketSensitiveFields(&bucket)
	response.SuccessWithMessage(c, "创建成功", bucket)
}

// Update 更新存储桶
// @Summary      更新存储桶
// @Description  更新指定存储桶配置
// @Tags         存储桶
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "存储桶 ID"
// @Param        data  body  model.StorageBucket  true  "存储桶配置"
// @Success      200  {object}  response.Response{data=model.StorageBucket} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets/{id} [put]
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

	maskBucketSensitiveFields(&bucket)
	response.SuccessWithMessage(c, "更新成功", bucket)
}

// Delete 删除存储桶
// @Summary      删除存储桶
// @Description  删除指定存储桶配置
// @Tags         存储桶
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "存储桶 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets/{id} [delete]
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
// @Summary      设置默认存储桶
// @Description  将指定存储桶设为默认
// @Tags         存储桶
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "存储桶 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets/{id}/default [put]
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
// @Summary      按驱动获取存储桶
// @Description  根据存储驱动获取存储桶列表
// @Tags         存储桶
// @Produce      json
// @Security     BearerAuth
// @Param        driver  path  string  true  "存储驱动"
// @Success      200  {object}  response.Response{data=[]model.StorageBucket} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets/driver/{driver} [get]
func (h *StorageBucketHandler) GetByDriver(c *gin.Context) {
	driver := c.Param("driver")
	if driver == "" {
		response.BadRequest(c, "驱动类型不能为空")
		return
	}

	buckets, err := h.service.GetByDriver(driver)
	if err != nil {
		response.InternalError(c, "获取存储桶列表失败")
		return
	}
	for i := range buckets {
		maskBucketSensitiveFields(&buckets[i])
	}
	response.Success(c, buckets)
}

// GetByPurpose 根据用途获取存储桶
// @Summary      按用途获取存储桶
// @Description  根据用途获取存储桶列表
// @Tags         存储桶
// @Produce      json
// @Security     BearerAuth
// @Param        purpose  path  string  true  "用途"
// @Success      200  {object}  response.Response{data=[]model.StorageBucket} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets/purpose/{purpose} [get]
func (h *StorageBucketHandler) GetByPurpose(c *gin.Context) {
	purpose := c.Param("purpose")
	if purpose == "" {
		response.BadRequest(c, "用途不能为空")
		return
	}

	buckets, err := h.service.GetByPurpose(purpose)
	if err != nil {
		response.InternalError(c, "获取存储桶列表失败")
		return
	}
	for i := range buckets {
		maskBucketSensitiveFields(&buckets[i])
	}
	response.Success(c, buckets)
}

// GetDefault 获取默认存储桶
// @Summary      获取默认存储桶
// @Description  获取当前默认存储桶配置
// @Tags         存储桶
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=model.StorageBucket} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets/default [get]
func (h *StorageBucketHandler) GetDefault(c *gin.Context) {
	bucket, err := h.service.GetDefault()
	if err != nil {
		response.NotFound(c, "未配置默认存储桶")
		return
	}
	maskBucketSensitiveFields(bucket)
	response.Success(c, bucket)
}

// TestConnection 测试存储桶连接
// @Summary      测试存储桶连接
// @Description  测试指定存储桶连接是否可用
// @Tags         存储桶
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "存储桶 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets/{id}/test [post]
func (h *StorageBucketHandler) TestConnection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的存储桶ID")
		return
	}

	msg, err := service.TestConnection(id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": msg})
}

// GetEnabledDrivers 获取已启用的存储驱动列表
// @Summary      获取已启用存储驱动
// @Description  获取当前已启用的存储驱动列表
// @Tags         存储桶
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets/enabled-drivers [get]
func (h *StorageBucketHandler) GetEnabledDrivers(c *gin.Context) {
	drivers := service.GetEnabledDrivers()
	response.Success(c, drivers)
}

// TestConnectionByDriver 根据驱动和桶名测试连接（无需先保存）
// @Summary      按驱动测试连接
// @Description  根据驱动和桶名测试连接，无需保存配置
// @Tags         存储桶
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  object  true  "测试参数"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/storage-buckets/test-by-driver [post]
func (h *StorageBucketHandler) TestConnectionByDriver(c *gin.Context) {
	var req struct {
		Driver     string `json:"driver" binding:"required"`
		BucketName string `json:"bucketName" binding:"required"`
		Region     string `json:"region"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	msg, err := service.TestConnectionByDriver(req.Driver, req.BucketName, req.Region)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": msg})
}
