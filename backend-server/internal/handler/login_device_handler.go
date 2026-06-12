package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// LoginDeviceHandler 登录设备处理器
type LoginDeviceHandler struct {
	service *service.LoginDeviceService
}

// NewLoginDeviceHandler 创建登录设备处理器
func NewLoginDeviceHandler() *LoginDeviceHandler {
	return &LoginDeviceHandler{
		service: service.NewLoginDeviceService(),
	}
}

// List 获取当前用户的登录设备列表
// @Summary      获取登录设备列表
// @Description  获取当前用户所有已登录设备，可按设备类型过滤
// @Tags         登录设备
// @Produce      json
// @Security     BearerAuth
// @Param        deviceType  query  string  false  "设备类型: web/h5/app/miniapp"
// @Success      200  {object}  response.Response{data=[]model.LoginDevice} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/devices [get]
func (h *LoginDeviceHandler) List(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	devices, err := h.service.ListByType(userID, c.Query("deviceType"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, devices)
}

// Kick 踢出指定设备
// @Summary      踢出指定设备
// @Description  将指定设备踢下线
// @Tags         登录设备
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "设备 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/devices/{id} [delete]
func (h *LoginDeviceHandler) Kick(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	currentDeviceID := middleware.GetCurrentDeviceID(c)
	if err := h.service.KickDevice(userID, uint(id), currentDeviceID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// KickAllOther 踢出所有其他设备
// @Summary      踢出所有其他设备
// @Description  踢出除当前设备外的所有设备
// @Tags         登录设备
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=map[string]int64} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/devices/kick-all [delete]
func (h *LoginDeviceHandler) KickAllOther(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	currentDeviceID := middleware.GetCurrentDeviceID(c)
	count, err := h.service.KickAllOther(userID, currentDeviceID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"kickedCount": count})
}
