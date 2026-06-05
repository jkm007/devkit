package handler

import (
	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// SystemSettingHandler 系统配置处理器
type SystemSettingHandler struct {
	service *service.SystemSettingService
}

// NewSystemSettingHandler 创建系统配置处理器
func NewSystemSettingHandler() *SystemSettingHandler {
	return &SystemSettingHandler{
		service: service.NewSystemSettingService(),
	}
}

// GetPublic 获取公开配置
// @Summary      获取公开配置
// @Description  获取前端需要的公开配置（无需登录），如站点名称、验证码开关等
// @Tags         系统设置
// @Produce      json
// @Success      200  {object}  response.Response{data=map[string]map[string]interface{}} "成功"
// @Router       /system/settings/public [get]
func (h *SystemSettingHandler) GetPublic(c *gin.Context) {
	data, err := h.service.GetPublic()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, data)
}

// GetAll 获取所有配置
// @Summary      获取所有配置
// @Description  获取所有分组的系统配置，用于系统设置页面
// @Tags         系统设置
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=map[string][]service.SettingItem} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/settings [get]
func (h *SystemSettingHandler) GetAll(c *gin.Context) {
	data, err := h.service.GetAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, data)
}

// GetByGroup 获取指定分组配置
// @Summary      获取指定分组配置
// @Description  获取单个分组的系统配置
// @Tags         系统设置
// @Produce      json
// @Security     BearerAuth
// @Param        group  path  string  true  "分组标识"  Enums(basic, email, sms, captcha, storage, wechat, security)
// @Success      200  {object}  response.Response{data=[]service.SettingItem} "成功"
// @Failure      400  {object}  response.Response "无效的分组"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/settings/{group} [get]
func (h *SystemSettingHandler) GetByGroup(c *gin.Context) {
	group := c.Param("group")
	data, err := h.service.GetByGroup(group)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, data)
}

// Update 批量更新配置
// @Summary      批量更新配置
// @Description  一次性更新多个分组的配置，敏感字段值为 "******" 则跳过
// @Tags         系统设置
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.UpdateSettingsRequest  true  "配置更新请求"
// @Success      200  {object}  response.Response{data=service.UpdateResult} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/settings [put]
func (h *SystemSettingHandler) Update(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req service.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.Update(&req, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// UpdateByGroup 更新指定分组配置
// @Summary      更新指定分组配置
// @Description  更新单个分组的系统配置
// @Tags         系统设置
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        group    path  string                      true  "分组标识"  Enums(basic, email, sms, captcha, storage, wechat, security)
// @Param        request  body  service.SettingGroupUpdateRequest  true  "配置更新请求"
// @Success      200  {object}  response.Response{data=service.UpdateResult} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/settings/{group} [put]
func (h *SystemSettingHandler) UpdateByGroup(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	group := c.Param("group")

	var req service.SettingGroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.UpdateByGroup(group, &req, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// TestEmail 测试邮件发送
// @Summary      测试邮件发送
// @Description  测试当前邮件配置是否正确，发送一封测试邮件
// @Tags         系统设置
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.TestEmailRequest  true  "测试邮件请求"
// @Success      200  {object}  response.Response{data=map[string]bool} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/settings/test-email [post]
func (h *SystemSettingHandler) TestEmail(c *gin.Context) {
	var req service.TestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.TestEmail(req.To); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"sent": true})
}

// TestSMS 测试短信发送
// @Summary      测试短信发送
// @Description  测试当前短信配置是否正确，发送一条测试短信
// @Tags         系统设置
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.TestSMSRequest  true  "测试短信请求"
// @Success      200  {object}  response.Response{data=map[string]bool} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      403  {object}  response.Response "权限不足"
// @Router       /system/settings/test-sms [post]
func (h *SystemSettingHandler) TestSMS(c *gin.Context) {
	var req service.TestSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.TestSMS(req.Phone); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"sent": true})
}
