package handler

import (
	"fmt"
	"net/http"

	"backend-server/config"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// WeChatHandler 微信登录处理器
type WeChatHandler struct {
	wechatService *service.WeChatService
	cfg           *config.Config
}

// NewWeChatHandler 创建微信登录处理器
func NewWeChatHandler() *WeChatHandler {
	return &WeChatHandler{
		wechatService: service.NewWeChatService(),
		cfg:           config.Get(),
	}
}

// cookieSecure 根据服务器模式决定 cookie 是否仅 HTTPS
func (h *WeChatHandler) cookieSecure() bool {
	return h.cfg.Server.Mode == "release"
}

// setCookie 设置 cookie（带 SameSite=Lax）
func (h *WeChatHandler) setCookie(c *gin.Context, name, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})
}

// MiniAppLoginRequest 小程序登录请求
type MiniAppLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// LoginByMiniProgram 小程序登录
// @Summary      微信小程序登录
// @Description  通过 wx.login() 获取的 code 登录，返回 JWT token
// @Tags         微信登录
// @Accept       json
// @Produce      json
// @Param        request body MiniAppLoginRequest true "登录请求"
// @Success      200  {object}  response.Response{data=service.LoginResponse} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Router       /auth/wechat/miniapp-login [post]
func (h *WeChatHandler) LoginByMiniProgram(c *gin.Context) {
	var req MiniAppLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "code is required")
		return
	}

	result, err := h.wechatService.LoginByMiniProgram(req.Code, c.ClientIP())
	if err != nil {
		h.wechatService.RecordSecurityLog(0, "login_fail", fmt.Sprintf("小程序登录失败: %s", err.Error()), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		response.BadRequest(c, err.Error())
		return
	}

	h.wechatService.RecordSecurityLog(result.ID, "login", fmt.Sprintf("小程序登录成功, 来源: %s", result.RegisterSource), c.ClientIP(), c.GetHeader("User-Agent"), 1)

	h.setCookie(c, "access_token", result.AccessToken, 30*24*3600)
	h.setCookie(c, "refresh_token", result.RefreshToken, 30*24*3600)

	response.Success(c, result)
}

// GetOfficialAuthorizeURL 获取公众号授权 URL
// @Summary      获取公众号 OAuth 授权 URL
// @Description  获取微信公众号网页授权页面 URL
// @Tags         微信登录
// @Produce      json
// @Param        scope  query  string  false  "授权范围: snsapi_base 或 snsapi_userinfo"
// @Success      200  {object}  response.Response{data=map[string]string} "成功"
// @Failure      400  {object}  response.Response "配置错误"
// @Router       /auth/wechat/official-authorize [get]
func (h *WeChatHandler) GetOfficialAuthorizeURL(c *gin.Context) {
	scope := c.Query("scope")
	if scope == "" {
		scope = "snsapi_userinfo"
	}
	// 白名单校验 scope
	if scope != "snsapi_base" && scope != "snsapi_userinfo" {
		response.BadRequest(c, "scope 只支持 snsapi_base 或 snsapi_userinfo")
		return
	}

	authURL, err := h.wechatService.GetOfficialAuthorizeURL(scope)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"url": authURL})
}

// LoginByOfficial 公众号 H5 登录回调
// @Summary      公众号授权回调
// @Description  微信公众号授权完成后的回调接口，返回 JWT token
// @Tags         微信登录
// @Produce      json
// @Param        code   query  string  true  "授权码"
// @Param        state  query  string  true  "状态参数"
// @Success      200  {object}  response.Response{data=service.LoginResponse} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Router       /auth/wechat/official-callback [get]
func (h *WeChatHandler) LoginByOfficial(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		response.BadRequest(c, "缺少必要参数: code, state")
		return
	}

	result, err := h.wechatService.LoginByOfficial(code, state, c.ClientIP())
	if err != nil {
		h.wechatService.RecordSecurityLog(0, "login_fail", fmt.Sprintf("公众号登录失败: %s", err.Error()), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		response.BadRequest(c, err.Error())
		return
	}

	h.wechatService.RecordSecurityLog(result.ID, "login", fmt.Sprintf("公众号登录成功, 来源: %s", result.RegisterSource), c.ClientIP(), c.GetHeader("User-Agent"), 1)

	h.setCookie(c, "access_token", result.AccessToken, 30*24*3600)
	h.setCookie(c, "refresh_token", result.RefreshToken, 30*24*3600)

	response.Success(c, result)
}

// GetWebAuthorizeURL 获取网站扫码授权 URL
// @Summary      获取微信扫码授权 URL
// @Description  获取微信网站应用扫码登录的授权页面 URL
// @Tags         微信登录
// @Produce      json
// @Success      200  {object}  response.Response{data=map[string]string} "成功"
// @Failure      400  {object}  response.Response "配置错误"
// @Router       /auth/wechat/web-authorize [get]
func (h *WeChatHandler) GetWebAuthorizeURL(c *gin.Context) {
	authURL, err := h.wechatService.GetWebAuthorizeURL()
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"url": authURL})
}

// LoginByWeb 网站扫码登录回调
// @Summary      微信扫码授权回调
// @Description  微信扫码授权完成后的回调接口，返回 JWT token
// @Tags         微信登录
// @Produce      json
// @Param        code   query  string  true  "授权码"
// @Param        state  query  string  true  "状态参数"
// @Success      200  {object}  response.Response{data=service.LoginResponse} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Router       /auth/wechat/web-callback [get]
func (h *WeChatHandler) LoginByWeb(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		response.BadRequest(c, "缺少必要参数: code, state")
		return
	}

	result, err := h.wechatService.LoginByWeb(code, state, c.ClientIP())
	if err != nil {
		h.wechatService.RecordSecurityLog(0, "login_fail", fmt.Sprintf("微信扫码登录失败: %s", err.Error()), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		response.BadRequest(c, err.Error())
		return
	}

	h.wechatService.RecordSecurityLog(result.ID, "login", fmt.Sprintf("微信扫码登录成功, 来源: %s", result.RegisterSource), c.ClientIP(), c.GetHeader("User-Agent"), 1)

	h.setCookie(c, "access_token", result.AccessToken, 30*24*3600)
	h.setCookie(c, "refresh_token", result.RefreshToken, 30*24*3600)

	response.Success(c, result)
}
