package handler

import (
	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// PasswordHistoryHandler 密码历史处理器
type PasswordHistoryHandler struct {
	service *service.PasswordHistoryService
}

// NewPasswordHistoryHandler 创建密码历史处理器
func NewPasswordHistoryHandler() *PasswordHistoryHandler {
	return &PasswordHistoryHandler{
		service: service.NewPasswordHistoryService(),
	}
}

// Check 检查新密码是否与历史密码重复
// @Summary      检查密码是否重复
// @Description  修改密码前，校验新密码是否与最近 5 次密码重复
// @Tags         密码历史
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.CheckPasswordRequest  true  "检查密码请求"
// @Success      200  {object}  response.Response{data=map[string]bool} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/password-history/check [post]
func (h *PasswordHistoryHandler) Check(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req service.CheckPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	isRepeated, err := h.service.CheckRepeated(userID, req.NewPassword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"isRepeated": isRepeated})
}
