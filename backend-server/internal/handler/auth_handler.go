package handler

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"backend-server/config"
	"backend-server/internal/middleware"
	"backend-server/internal/service"
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

// cookieSecure 始终启用 Secure 标志，确保 Cookie 仅通过 HTTPS 传输
func (h *AuthHandler) cookieSecure() bool {
	return true
}

// setCookie 设置 cookie（带 SameSite=Lax）
func (h *AuthHandler) setCookie(c *gin.Context, name, value string, maxAge int) {
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

// getClientType 获取客户端类型，默认 web；允许 web/h5/app/miniapp。
func getClientType(c *gin.Context) string {
	clientType := strings.ToLower(strings.TrimSpace(c.GetHeader("X-Client-Type")))
	switch clientType {
	case "h5", "app", "miniapp", "web":
		return clientType
	default:
		return "web"
	}
}

func getDeviceMeta(c *gin.Context) service.DeviceMeta {
	return service.DeviceMeta{
		AppVersion:    strings.TrimSpace(c.GetHeader("X-App-Version")),
		SystemVersion: strings.TrimSpace(c.GetHeader("X-System-Version")),
		DeviceModel:   strings.TrimSpace(c.GetHeader("X-Device-Model")),
		Platform:      strings.TrimSpace(c.GetHeader("X-Platform")),
		Channel:       strings.TrimSpace(c.GetHeader("X-Channel")),
	}
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
		// 记录登录失败日志（保留完整错误信息用于内部排查）
		h.authService.RecordSecurityLog(0, "login_fail", fmt.Sprintf("用户[%s]登录失败: %s", req.Username, err.Error()), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		// 区分验证码错误和账号错误
		errMsg := err.Error()
		if strings.Contains(errMsg, "验证码") || strings.Contains(errMsg, "captcha") || strings.Contains(errMsg, "验证失败") {
			// 返回 403001 让前端弹出验证码（不关闭弹窗，刷新验证码重试）
			response.CaptchaRequired(c, "验证码错误，请重新验证")
		} else {
			response.Forbidden(c, "用户名或密码错误")
		}
		return
	}

	// 记录登录成功日志
	h.authService.RecordSecurityLog(result.ID, "login", fmt.Sprintf("登录成功, 来源: %s", result.RegisterSource), c.ClientIP(), c.GetHeader("User-Agent"), 1)
	// 使用前端传来的 X-Device-ID，没有则用 User-Agent+IP 生成
	deviceID := c.GetHeader("X-Device-ID")
	h.authService.RecordLoginDevice(result.ID, c.ClientIP(), c.GetHeader("User-Agent"), deviceID, getClientType(c), getDeviceMeta(c))

	// 双重返回 Token：Cookie 用于浏览器自动携带认证，响应体用于前端显式管理（如存入内存或 localStorage）
	// Cookie MaxAge 与对应 Token TTL 保持一致，避免 Cookie 存活超过 Token 有效期
	accessTokenMaxAge := int(h.cfg.JWT.AccessTokenTTL.Seconds())
	refreshTokenMaxAge := int(h.cfg.JWT.RefreshTokenTTL.Seconds())
	h.setCookie(c, "access_token", result.AccessToken, accessTokenMaxAge)
	h.setCookie(c, "refresh_token", result.RefreshToken, refreshTokenMaxAge)

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
	var tokenUserID uint
	if tokenStr != "" {
		// 将 Token 加入 Redis 黑名单，TTL 与 access token 一致
		claims, err := jwt.Parse(tokenStr)
		ttl := h.cfg.JWT.AccessTokenTTL
		if err == nil {
			tokenUserID = claims.UserID
			if claims.ExpiresAt != nil {
				remaining := time.Until(claims.ExpiresAt.Time)
				if remaining > 0 {
					ttl = remaining
				}
			}
		}
		// 使用 SHA-256 哈希存储，减少内存占用并降低 Token 泄露风险
		blacklistKey := fmt.Sprintf("token_blacklist:%s", sha256HashString(tokenStr))
		database.GetRedis().Set(c, blacklistKey, "1", ttl)
	}

	// 同时清除 refresh token。Logout 是公开路由，未经过 JWTAuth 时从 token claims 回退获取用户 ID。
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		userID = tokenUserID
	}
	if userID > 0 {
		_ = h.authService.Logout(userID)
	}

	// 清除 Cookie
	h.setCookie(c, "access_token", "", -1)
	h.setCookie(c, "refresh_token", "", -1)

	response.Success(c, "")
}

// RefreshTokenRequest 刷新 Token 请求
// H5/App 等 Cookie 不稳定的客户端可通过 Body 传递 refreshToken。
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshToken 刷新 AccessToken（带 Token 轮换）
// @Summary      刷新 AccessToken
// @Description  使用 Cookie/Header/Body 中的 RefreshToken 获取新的 AccessToken + RefreshToken（Token 轮换，旧 Token 自动失效）
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body handler.RefreshTokenRequest false "刷新 Token 请求（H5/App 可选）"
// @Success      200  {object}  response.Response{data=service.TokenRefreshResponse} "成功"
// @Failure      403  {object}  response.Response "RefreshToken 无效或已过期"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 优先从 Cookie 获取 RefreshToken，兼容 H5/App 从 Header 或 Body 传递
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		refreshToken = strings.TrimSpace(c.GetHeader("X-Refresh-Token"))
	}
	if refreshToken == "" {
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			refreshToken = strings.TrimSpace(req.RefreshToken)
		}
	}
	if refreshToken == "" {
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

	// 用新的 Token 更新 Cookie，MaxAge 与对应 Token TTL 保持一致
	accessTokenMaxAge := int(h.cfg.JWT.AccessTokenTTL.Seconds())
	refreshTokenMaxAge := int(h.cfg.JWT.RefreshTokenTTL.Seconds())
	h.setCookie(c, "access_token", result.AccessToken, accessTokenMaxAge)
	h.setCookie(c, "refresh_token", result.RefreshToken, refreshTokenMaxAge)

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
		response.InternalError(c, "获取权限信息失败，请稍后重试")
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
		response.InternalError(c, "获取用户信息失败，请稍后重试")
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
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.authService.UpdateProfile(userID, &req); err != nil {
		response.InternalError(c, "更新个人资料失败，请稍后重试")
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
		response.BadRequest(c, "请求参数错误")
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
		// 用户输入校验类错误允许展示，其余返回通用消息避免暴露内部实现细节
		if isUserInputError(err.Error()) {
			response.BadRequest(c, err.Error())
		} else {
			response.BadRequest(c, "密码修改失败，请稍后重试")
		}
		return
	}

	response.Success(c, nil)
}

// LoginByEmailRequest 邮箱验证码登录请求
type LoginByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
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
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.authService.LoginByEmail(req.Email, req.Code, c.ClientIP())
	if err != nil {
		// 记录登录失败日志（保留完整错误信息用于内部排查）
		h.authService.RecordSecurityLog(0, "login_fail", fmt.Sprintf("邮箱[%s]登录失败: %s", req.Email, err.Error()), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		// 区分验证码错误和账号错误
		errMsg := err.Error()
		if strings.Contains(errMsg, "验证码") {
			response.BadRequest(c, errMsg)
		} else {
			// 不泄露账号是否存在
			response.BadRequest(c, "登录失败，请稍后重试")
		}
		return
	}

	// 记录登录成功日志
	h.authService.RecordSecurityLog(result.ID, "login", fmt.Sprintf("邮箱验证码登录成功, 来源: %s", result.RegisterSource), c.ClientIP(), c.GetHeader("User-Agent"), 1)
	deviceID := c.GetHeader("X-Device-ID")
	h.authService.RecordLoginDevice(result.ID, c.ClientIP(), c.GetHeader("User-Agent"), deviceID, getClientType(c), getDeviceMeta(c))

	// 双重返回 Token：Cookie 用于浏览器自动携带认证，响应体用于前端显式管理
	accessTokenMaxAge := int(h.cfg.JWT.AccessTokenTTL.Seconds())
	refreshTokenMaxAge := int(h.cfg.JWT.RefreshTokenTTL.Seconds())
	h.setCookie(c, "access_token", result.AccessToken, accessTokenMaxAge)
	h.setCookie(c, "refresh_token", result.RefreshToken, refreshTokenMaxAge)

	response.Success(c, result)
}

// LoginByPhoneRequest 手机号验证码登录请求
type LoginByPhoneRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Code  string `json:"code" binding:"required,len=6"`
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
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.authService.LoginByPhone(req.Phone, req.Code, c.ClientIP())
	if err != nil {
		// 记录登录失败日志（保留完整错误信息用于内部排查）
		h.authService.RecordSecurityLog(0, "login_fail", fmt.Sprintf("手机号[%s]登录失败: %s", req.Phone, err.Error()), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		// 统一返回通用错误消息，避免泄露账号是否存在等内部信息
		response.BadRequest(c, "登录失败，请稍后重试")
		return
	}

	// 记录登录成功日志
	h.authService.RecordSecurityLog(result.ID, "login", fmt.Sprintf("手机号验证码登录成功, 来源: %s", result.RegisterSource), c.ClientIP(), c.GetHeader("User-Agent"), 1)
	deviceID := c.GetHeader("X-Device-ID")
	h.authService.RecordLoginDevice(result.ID, c.ClientIP(), c.GetHeader("User-Agent"), deviceID, getClientType(c), getDeviceMeta(c))

	// 双重返回 Token：Cookie 用于浏览器自动携带认证，响应体用于前端显式管理
	accessTokenMaxAge := int(h.cfg.JWT.AccessTokenTTL.Seconds())
	refreshTokenMaxAge := int(h.cfg.JWT.RefreshTokenTTL.Seconds())
	h.setCookie(c, "access_token", result.AccessToken, accessTokenMaxAge)
	h.setCookie(c, "refresh_token", result.RefreshToken, refreshTokenMaxAge)

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
		response.BadRequest(c, "请求参数错误")
		return
	}

	userID, err := h.authService.Register(&req, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		// 记录注册失败日志（保留完整错误信息用于内部排查）
		h.authService.RecordSecurityLog(0, "register_fail", err.Error(), c.ClientIP(), c.GetHeader("User-Agent"), 0)
		// 统一返回通用错误消息，避免泄露用户名/邮箱是否已存在等内部信息
		response.BadRequest(c, "注册失败，请稍后重试")
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
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.authService.ResetPassword(&req, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		// 统一返回通用错误消息，避免泄露邮箱是否已注册等内部信息
		response.BadRequest(c, "密码重置失败，请稍后重试")
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
		response.InternalError(c, "获取权限版本失败，请稍后重试")
		return
	}

	// 排序后计算 hash，确保一致性（使用 SHA-256 替代 MD5）
	sort.Strings(codes)
	hash := sha256HashString(strings.Join(codes, ","))
	response.Success(c, hash)
}

// isUserInputError 判断错误是否属于用户输入校验类错误（允许展示给用户）
// 仅匹配已知的用户输入校验错误关键词，其余一律视为内部错误不予暴露
func isUserInputError(errMsg string) bool {
	userInputKeywords := []string{
		"验证码",
		"密码不一致",
		"密码长度",
		"旧密码错误",
	}
	for _, keyword := range userInputKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	return false
}

// sha256HashString 计算 SHA-256 哈希
func sha256HashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}
