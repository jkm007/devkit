package handler

import (
	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserPrivacyHandler 用户隐私设置处理器
type UserPrivacyHandler struct {
	service *service.UserPrivacyService
}

// NewUserPrivacyHandler 创建用户隐私设置处理器
func NewUserPrivacyHandler() *UserPrivacyHandler {
	return &UserPrivacyHandler{
		service: service.NewUserPrivacyService(),
	}
}

// Get 获取当前用户的隐私设置
// @Summary      获取隐私设置
// @Description  获取当前用户的隐私设置
// @Tags         隐私设置
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=model.UserPrivacy} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /user/privacy [get]
func (h *UserPrivacyHandler) Get(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	privacy, err := h.service.Get(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, privacy)
}

// Update 更新当前用户的隐私设置
// @Summary      更新隐私设置
// @Description  更新当前用户的隐私设置
// @Tags         隐私设置
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.PrivacyRequest  true  "隐私设置请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /user/privacy [put]
func (h *UserPrivacyHandler) Update(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req service.PrivacyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.Update(userID, &req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
