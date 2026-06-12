package handler

import (
	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHomeHandler 用户首页处理器
type UserHomeHandler struct {
	userHomeService *service.UserHomeService
}

func NewUserHomeHandler() *UserHomeHandler {
	return &UserHomeHandler{
		userHomeService: service.NewUserHomeService(),
	}
}

// GetHomeData 获取用户首页数据
func (h *UserHomeHandler) GetHomeData(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	data, err := h.userHomeService.GetHomeData(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, data)
}
