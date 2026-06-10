package handler

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	"backend-server/config"
	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/captcha"
	"backend-server/pkg/database"
	"backend-server/pkg/jwt"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
	cfg         *config.Config
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
		cfg:         config.Get(),
	}
}

// cookieSecure 根据服务器模式决定 cookie 是否仅 HTTPS
func (h *AuthHandler) cookieSecure() bool {
	return h.cfg.Server.Mode == "release"
}

// checkCaptchaForAltLogin 验证码校验（邮箱/短信登录共用）
// 读取验证码系统设置，根据配置判断是否需要图形验证码
func (h *AuthHandler) checkCaptchaForAltLogin(captchaID, captchaCode string, startTime int64, points []captcha.Point) error {
	// 获取验证码设置
	settings := h.getCaptchaSettings()
	enabled, _ := settings["enabled"].(bool)
	if !enabled {
		return nil
	}

	// 需要验证码但未提供
	if captchaID == "" {
		return fmt.Errorf("请完成图形验证码验证")
	}

	// 服务端验证码验证
	valid, msg := captcha.Verify(captchaID, captchaCode, startTime, points)
	if !valid {
		return fmt.Errorf("%s", msg)
	}

	return nil
}

// getCaptchaSettings 获取验证码设置
func (h *AuthHandler) getCaptchaSettings() map[string]interface{} {
	db := database.GetMySQL()
	var settings []struct {
		Key   string
		Value string
	}
	db.Raw("SELECT `key`, `value` FROM system_settings WHERE group_key = ? AND deleted_at IS NULL", "captcha").Scan(&settings)

	result := map[string]interface{}{
		"enabled": false,
	}
	for _, s := range settings {
		if s.Key == "captcha_enabled" {
			result["enabled"] = s.Value == "true"
		}
	}
	return result
}

// Login 用户登录
// @Summary      用户登录
// @Description  用户名密码登录，返回 AccessToken 和用户信息
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body service.LoginRequest true "登录请求"
// @Success      200  {object}  response.Response{data=service.LoginResponse} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      403  {object}  response.Response "账号密码错误或账号已禁用"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Username and password are required")
		return
	}

	result, err := h.authService.Login(&req, c.ClientIP())
	if err != nil {
		// 记录登录失败日志
		h.authService.RecordSecurityLog(0, "login_fail", err.Error(), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		response.Forbidden(c, err.Error())
		return
	}

	// 记录登录成功日志
	h.authService.RecordSecurityLog(result.ID, "login", fmt.Sprintf("登录成功, 来源: %s", result.RegisterSource), c.ClientIP(), c.GetHeader("User-Agent"), 1)
	// 使用前端传来的 X-Device-ID，没有则用 User-Agent+IP 生成
	deviceID := c.GetHeader("X-Device-ID")
	h.authService.RecordLoginDevice(result.ID, c.ClientIP(), c.GetHeader("User-Agent"), deviceID)

	// 设置 AccessToken 和 RefreshToken 到独立的 Cookie
	c.SetCookie("access_token", result.AccessToken, 30*24*3600, "/", "", h.cookieSecure(), true)
	c.SetCookie("refresh_token", result.RefreshToken, 30*24*3600, "/", "", h.cookieSecure(), true)

	response.Success(c, result)
}

// Logout 用户登出
// @Summary      用户登出
// @Description  清除登录状态，将 Token 加入黑名单
// @Tags         认证
// @Produce      json
// @Success      200  {object}  response.Response "成功"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 获取当前 Token 并加入黑名单
	// 优先从 Authorization header 读取，如果为空则从 Cookie 中读取
	tokenStr := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenStr = parts[1]
		}
	}
	if tokenStr == "" {
		tokenStr, _ = c.Cookie("access_token")
	}
	if tokenStr != "" {
		// 将 Token 加入 Redis 黑名单，TTL 与 access token 一致
		claims, err := jwt.Parse(tokenStr)
		ttl := h.cfg.JWT.AccessTokenTTL
		if err == nil && claims.ExpiresAt != nil {
			remaining := time.Until(claims.ExpiresAt.Time)
			if remaining > 0 {
				ttl = remaining
			}
		}
		blacklistKey := fmt.Sprintf("token_blacklist:%s", tokenStr)
		database.GetRedis().Set(c, blacklistKey, "1", ttl)
	}

	// 同时清除 refresh token
	userID := middleware.GetCurrentUserID(c)
	if userID > 0 {
		_ = h.authService.Logout(userID)
	}

	// 清除 Cookie
	c.SetCookie("access_token", "", -1, "/", "", h.cookieSecure(), true)
	c.SetCookie("refresh_token", "", -1, "/", "", h.cookieSecure(), true)

	response.Success(c, "")
}

// RefreshToken 刷新 AccessToken（带 Token 轮换）
// @Summary      刷新 AccessToken
// @Description  使用 Cookie 中的 RefreshToken 获取新的 AccessToken + RefreshToken（Token 轮换，旧 Token 自动失效）
// @Tags         认证
// @Produce      json
// @Success      200  {object}  response.Response{data=service.TokenRefreshResponse} "成功"
// @Failure      403  {object}  response.Response "RefreshToken 无效或已过期"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从 Cookie 获取 RefreshToken
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		response.Forbidden(c, "Forbidden Exception")
		return
	}

	// 解析 Token（仅验证签名和过期，不做业务校验）
	claims, err := jwt.Parse(refreshToken)
	if err != nil {
		response.Forbidden(c, "Forbidden Exception")
		return
	}

	// 刷新 Token（验证哈希 + 轮换）
	result, err := h.authService.RefreshToken(claims.UserID, refreshToken)
	if err != nil {
		response.Forbidden(c, "Forbidden Exception")
		return
	}

	// 用新的 RefreshToken 更新 Cookie
	c.SetCookie("refresh_token", result.RefreshToken, 30*24*3600, "/", "", h.cookieSecure(), true)

	response.Success(c, result)
}

// GetPermissionCodes 获取权限码
// @Summary      获取权限码
// @Description  获取当前用户的权限码列表（带 Redis 缓存，10 分钟 TTL）
// @Tags         认证
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]string} "权限码列表"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/codes [get]
func (h *AuthHandler) GetPermissionCodes(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	codes, err := h.authService.GetPermissionCodes(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, codes)
}

// GetUserInfo 获取当前用户信息
// @Summary      获取当前用户信息
// @Description  获取当前登录用户的详细信息，包括 ID、用户名、角色列表
// @Tags         认证
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=service.LoginResponse} "用户信息"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /user/info [get]
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	userInfo, err := h.authService.GetUserInfo(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, userInfo)
}

// UpdateProfile 更新当前用户个人资料
// @Summary      更新个人资料
// @Description  更新当前登录用户的昵称、邮箱、手机、性别、生日、简介
// @Tags         认证
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.UpdateProfileRequest  true  "更新资料请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /user/info [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authService.UpdateProfile(userID, &req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ChangePassword 修改密码
// @Summary      修改密码
// @Description  当前用户修改自己的密码
// @Tags         认证
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  service.ChangePasswordRequest  true  "修改密码请求"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/change-password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 获取当前 access token，用于修改密码后加入黑名单
	tokenStr := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenStr = parts[1]
		}
	}

	if err := h.authService.ChangePassword(userID, &req, c.ClientIP(), c.GetHeader("User-Agent"), tokenStr); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// LoginByEmailRequest 邮箱验证码登录请求
type LoginByEmailRequest struct {
	Email       string         `json:"email" binding:"required,email"`
	Code        string         `json:"code" binding:"required,len=6"`
	CaptchaID   string         `json:"captchaId"`   // 图形验证码 ID
	CaptchaCode string         `json:"captchaCode"` // 图形验证码值
	Points      []captcha.Point `json:"points"`      // 点选验证码坐标
	StartTime   int64          `json:"startTime"`   // 验证码生成时间（毫秒）
}

// LoginByEmail 邮箱验证码登录
// @Summary      邮箱验证码登录
// @Description  使用邮箱+验证码登录，返回 AccessToken 和用户信息
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  handler.LoginByEmailRequest  true  "邮箱登录请求"
// @Success      200   {object}  response.Response{data=service.LoginResponse} "成功"
// @Failure      400   {object}  response.Response "参数错误或验证码错误"
// @Router       /auth/login-by-email [post]
func (h *AuthHandler) LoginByEmail(c *gin.Context) {
	var req LoginByEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 图形验证码校验（与密码登录一致）
	if err := h.checkCaptchaForAltLogin(req.CaptchaID, req.CaptchaCode, req.StartTime, req.Points); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.authService.LoginByEmail(req.Email, req.Code, c.ClientIP())
	if err != nil {
		h.authService.RecordSecurityLog(0, "login_fail", "邮箱登录失败: "+err.Error(), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		response.BadRequest(c, err.Error())
		return
	}

	// 记录登录成功日志
	h.authService.RecordSecurityLog(result.ID, "login", fmt.Sprintf("邮箱验证码登录成功, 来源: %s", result.RegisterSource), c.ClientIP(), c.GetHeader("User-Agent"), 1)
	deviceID := c.GetHeader("X-Device-ID")
	h.authService.RecordLoginDevice(result.ID, c.ClientIP(), c.GetHeader("User-Agent"), deviceID)

	// 设置 Cookie
	c.SetCookie("access_token", result.AccessToken, 30*24*3600, "/", "", h.cookieSecure(), true)
	c.SetCookie("refresh_token", result.RefreshToken, 30*24*3600, "/", "", h.cookieSecure(), true)

	response.Success(c, result)
}

// LoginByPhoneRequest 手机号验证码登录请求
type LoginByPhoneRequest struct {
	Phone       string         `json:"phone" binding:"required,len=11"`
	Code        string         `json:"code" binding:"required,len=6"`
	CaptchaID   string         `json:"captchaId"`   // 图形验证码 ID
	CaptchaCode string         `json:"captchaCode"` // 图形验证码值
	Points      []captcha.Point `json:"points"`      // 点选验证码坐标
	StartTime   int64          `json:"startTime"`   // 验证码生成时间（毫秒）
}

// LoginByPhone 手机号验证码登录
// @Summary      手机号验证码登录
// @Description  使用手机号+短信验证码登录，返回 AccessToken 和用户信息
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  handler.LoginByPhoneRequest  true  "手机号登录请求"
// @Success      200   {object}  response.Response{data=service.LoginResponse} "成功"
// @Failure      400   {object}  response.Response "参数错误或验证码错误"
// @Router       /auth/login-by-phone [post]
func (h *AuthHandler) LoginByPhone(c *gin.Context) {
	var req LoginByPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 图形验证码校验（与密码登录一致）
	if err := h.checkCaptchaForAltLogin(req.CaptchaID, req.CaptchaCode, req.StartTime, req.Points); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.authService.LoginByPhone(req.Phone, req.Code, c.ClientIP())
	if err != nil {
		h.authService.RecordSecurityLog(0, "login_fail", "手机号登录失败: "+err.Error(), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		response.BadRequest(c, err.Error())
		return
	}

	// 记录登录成功日志
	h.authService.RecordSecurityLog(result.ID, "login", fmt.Sprintf("手机号验证码登录成功, 来源: %s", result.RegisterSource), c.ClientIP(), c.GetHeader("User-Agent"), 1)
	deviceID := c.GetHeader("X-Device-ID")
	h.authService.RecordLoginDevice(result.ID, c.ClientIP(), c.GetHeader("User-Agent"), deviceID)

	// 设置 Cookie
	c.SetCookie("access_token", result.AccessToken, 30*24*3600, "/", "", h.cookieSecure(), true)
	c.SetCookie("refresh_token", result.RefreshToken, 30*24*3600, "/", "", h.cookieSecure(), true)

	response.Success(c, result)
}

// Register 用户注册
// @Summary      用户注册
// @Description  新用户注册（需要邮箱验证码）
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  service.RegisterRequest  true  "注册请求"
// @Success      200   {object}  response.Response
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, err := h.authService.Register(&req, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		h.authService.RecordSecurityLog(0, "register_fail", err.Error(), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		response.BadRequest(c, err.Error())
		return
	}

	h.authService.RecordSecurityLog(userID, "register", fmt.Sprintf("新用户注册: %s", req.Username), c.ClientIP(), c.GetHeader("User-Agent"), 1)
	response.SuccessWithMessage(c, "注册成功", nil)
}

// ResetPassword 重置密码
// @Summary      重置密码
// @Description  通过邮箱验证码重置密码
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  service.ResetPasswordRequest  true  "重置密码请求"
// @Success      200   {object}  response.Response
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authService.ResetPassword(&req, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "密码重置成功", nil)
}

// GetPermissionVersion 获取权限版本
// @Summary      获取权限版本
// @Description  返回用户当前权限码的 SHA-256 hash，用于前端检测权限是否变更（轮询比对）
// @Tags         认证
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=string} "权限版本 hash"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /auth/permission-version [get]
func (h *AuthHandler) GetPermissionVersion(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	codes, err := h.authService.GetPermissionCodes(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 排序后计算 hash，确保一致性（使用 SHA-256 替代 MD5）
	sort.Strings(codes)
	hash := sha256HashString(strings.Join(codes, ","))
	response.Success(c, hash)
}

// sha256HashString 计算 SHA-256 哈希
func sha256HashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}
