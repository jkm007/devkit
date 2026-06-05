package handler

import (
	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// OAuthHandler 第三方登录绑定处理器
type OAuthHandler struct {
	service *service.OAuthService
}

// NewOAuthHandler 创建第三方登录绑定处理器
func NewOAuthHandler() *OAuthHandler {
	return &OAuthHandler{
		service: service.NewOAuthService(),
	}
}

// GetBindings 获取当前用户的第三方绑定列表
// @Summary      获取第三方绑定列表
// @Description  获取当前用户已绑定的第三方账号列表
// @Tags         OAuth绑定
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]service.OAuthBindingItem} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/oauth/bindings [get]
func (h *OAuthHandler) GetBindings(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	bindings, err := h.service.ListBindings(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, bindings)
}

// GetBindURL 获取第三方授权 URL
// @Summary      获取第三方授权 URL
// @Description  获取第三方平台的授权页面 URL
// @Tags         OAuth绑定
// @Produce      json
// @Security     BearerAuth
// @Param        provider     query  string  true  "提供商：wechat/github/google"
// @Param        redirectUri  query  string  false "授权完成后的回调地址"
// @Success      200  {object}  response.Response{data=map[string]string} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/oauth/bind-url [get]
func (h *OAuthHandler) GetBindURL(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		response.BadRequest(c, "provider is required")
		return
	}

	// TODO: 根据 provider 生成真实的授权 URL
	url := "https://example.com/oauth/authorize?provider=" + provider
	response.Success(c, gin.H{"url": url})
}

// Callback 第三方授权回调
// @Summary      第三方授权回调
// @Description  第三方平台授权完成后的回调接口
// @Tags         OAuth绑定
// @Produce      json
// @Param        provider  query  string  true  "提供商"
// @Param        code      query  string  true  "授权码"
// @Param        state     query  string  true  "状态参数"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Router       /auth/oauth/callback [get]
func (h *OAuthHandler) Callback(c *gin.Context) {
	// TODO: 实现 OAuth code-for-token exchange + state 验证
	response.BadRequest(c, "OAuth callback not yet implemented")
}

// Unbind 解绑第三方账号
// @Summary      解绑第三方账号
// @Description  解绑指定的第三方账号
// @Tags         OAuth绑定
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.OAuthUnbindRequest  true  "解绑请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/oauth/unbind [post]
func (h *OAuthHandler) Unbind(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req service.OAuthUnbindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.Unbind(userID, req.Provider); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}
