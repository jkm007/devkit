package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/cache"
	"backend-server/pkg/captcha"
	"backend-server/pkg/database"
	"backend-server/pkg/email"
	"backend-server/pkg/jwt"
	"backend-server/pkg/logger"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	userRepo         *repository.UserRepo
	roleRepo         *repository.RoleRepo
	menuRepo         *repository.MenuRepo
	groupRepo        *repository.GroupRepo
	securityLogSvc   *SecurityLogService
	loginDeviceSvc   *LoginDeviceService
}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	db := database.GetMySQL()
	return &AuthService{
		userRepo:       repository.NewUserRepo(db),
		roleRepo:       repository.NewRoleRepo(db),
		menuRepo:       repository.NewMenuRepo(db),
		groupRepo:      repository.NewGroupRepo(db),
		securityLogSvc: NewSecurityLogService(),
		loginDeviceSvc: NewLoginDeviceService(),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string           `json:"username" binding:"required"`
	Password    string           `json:"password" binding:"required"`
	CaptchaID   string           `json:"captchaId"`   // 验证码 ID
	CaptchaCode string           `json:"captchaCode"` // 验证码值
	CaptchaType string           `json:"captchaType"` // 验证码类型
	Points      []captcha.Point  `json:"points"`      // 点选验证码坐标
	StartTime   int64            `json:"startTime"`   // 验证码生成时间（毫秒，用于时间检测）
}

// LoginResponse 登录响应
type LoginResponse struct {
	ID             uint     `json:"id"`
	Username       string   `json:"username"`
	Nickname       string   `json:"nickname"`
	RealName       string   `json:"realName"`
	Email          string   `json:"email"`
	Phone          string   `json:"phone"`
	Avatar         string   `json:"avatar"`
	Gender         int      `json:"gender"`
	Birthday       string   `json:"birthday,omitempty"`
	Bio            string   `json:"bio"`
	Roles          []string `json:"roles"`
	RegisterSource string   `json:"registerSource,omitempty"`
	HomePath       string   `json:"homePath,omitempty"`
	AccessToken    string   `json:"accessToken,omitempty"`
	RefreshToken   string   `json:"refreshToken,omitempty"`
}

// loginFailCountKey 生成登录失败计数的 Redis key
func loginFailCountKey(clientIP string) string {
	return fmt.Sprintf("login_fail:%s", clientIP)
}

// getCaptchaSettings 获取验证码设置
func (s *AuthService) getCaptchaSettings() (enabled bool, trigger int) {
	db := database.GetMySQL()
	var settings []model.SystemSetting
	db.Where("group_key = ? AND deleted_at IS NULL", "captcha").Find(&settings)

	enabled = false
	trigger = 0 // 0 表示始终显示

	for _, setting := range settings {
		switch setting.Key {
		case "captcha_enabled":
			enabled = setting.Value == "true"
		case "captcha_login_trigger":
			if v, err := strconv.Atoi(setting.Value); err == nil {
				trigger = v
			}
		}
	}
	return
}

// checkCaptcha 验证码校验
func (s *AuthService) checkCaptcha(req *LoginRequest, clientIP string) error {
	enabled, trigger := s.getCaptchaSettings()
	if !enabled {
		return nil
	}

	// 判断是否需要验证码
	needCaptcha := false
	if trigger == 0 {
		needCaptcha = true
	} else {
		rdb := database.GetRedis()
		ctx := context.Background()
		key := loginFailCountKey(clientIP)
		countStr, err := rdb.Get(ctx, key).Result()
		if err == nil {
			count, _ := strconv.Atoi(countStr)
			if count >= trigger {
				needCaptcha = true
			}
		}
	}

	if !needCaptcha {
		return nil
	}

	// 需要验证码但未提供
	if req.CaptchaID == "" {
		return errors.New("Please complete the captcha")
	}

	// 服务端验证码验证（支持所有类型）
	valid, msg := captcha.Verify(req.CaptchaID, req.CaptchaCode, req.StartTime, req.Points)
	if !valid {
		return errors.New(msg)
	}

	return nil
}

// recordLoginFail 记录登录失败（IP 维度）
func (s *AuthService) recordLoginFail(clientIP string) {
	rdb := database.GetRedis()
	ctx := context.Background()
	key := loginFailCountKey(clientIP)
	rdb.Incr(ctx, key)
	rdb.Expire(ctx, key, 30*time.Minute) // 30 分钟后自动清除
}

// clearLoginFail 清除登录失败计数
func (s *AuthService) clearLoginFail(clientIP string) {
	rdb := database.GetRedis()
	ctx := context.Background()
	rdb.Del(ctx, loginFailCountKey(clientIP))
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest, clientIP string) (*LoginResponse, error) {
	// 验证码校验
	if err := s.checkCaptcha(req, clientIP); err != nil {
		return nil, err
	}

	// 查找用户
	user, err := s.userRepo.GetByName(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.recordLoginFail(clientIP)
			return nil, errors.New("Username or password is incorrect.")
		}
		logger.Error("登录时数据库查询失败", zap.Error(err))
		return nil, errors.New("Username or password is incorrect.")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.recordLoginFail(clientIP)
		return nil, errors.New("Username or password is incorrect.")
	}

	// 检查状态
	if user.Status != 1 {
		return nil, errors.New("Account is disabled")
	}

	return s.generateLoginResponse(user, clientIP)
}

// LoginByEmail 邮箱验证码登录
func (s *AuthService) LoginByEmail(email, code, clientIP string) (*LoginResponse, error) {
	// 验证邮箱验证码
	verifyCodeService := NewVerifyCodeService()
	valid, err := verifyCodeService.VerifyCode(email, code, "login")
	if err != nil {
		return nil, fmt.Errorf("验证码验证失败: %w", err)
	}
	if !valid {
		return nil, errors.New("验证码错误")
	}

	// 根据邮箱查找用户
	user, err := s.userRepo.GetByEmail(email)
	if err != nil || user == nil || user.ID == 0 {
		return nil, errors.New("该邮箱未注册")
	}

	// 检查状态
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	return s.generateLoginResponse(user, clientIP)
}

// LoginByPhone 手机号验证码登录
func (s *AuthService) LoginByPhone(phone, code, clientIP string) (*LoginResponse, error) {
	// 验证短信验证码
	verifyCodeService := NewVerifyCodeService()
	valid, err := verifyCodeService.VerifySMSCode(phone, code, "login")
	if err != nil {
		return nil, fmt.Errorf("验证码验证失败: %w", err)
	}
	if !valid {
		return nil, errors.New("验证码错误")
	}

	// 根据手机号查找用户
	user, err := s.userRepo.GetByPhone(phone)
	if err != nil || user == nil || user.ID == 0 {
		return nil, errors.New("该手机号未注册")
	}

	// 检查状态
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	return s.generateLoginResponse(user, clientIP)
}

// generateLoginResponse 为已验证用户生成登录响应（token、角色、日志等）
func (s *AuthService) generateLoginResponse(user *model.User, clientIP string) (*LoginResponse, error) {
	// 收集角色名称
	roleNames, err := s.collectRoleNames(user)
	if err != nil {
		return nil, err
	}
	if len(roleNames) == 0 {
		return nil, errors.New("User has no assigned roles")
	}

	// 生成 Token
	tokenPair, err := jwt.Generate(user.ID, user.Name, roleNames)
	if err != nil {
		return nil, err
	}

	// 存储 RefreshToken 哈希到 Redis
	rdb := database.GetRedis()
	rdb.Set(context.Background(), fmt.Sprintf("refresh_token:%d", user.ID), hashToken(tokenPair.RefreshToken), 30*24*time.Hour)

	// 更新用户最后登录信息
	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = clientIP
	_ = s.userRepo.Update(user)

	// 格式化生日
	birthday := ""
	if user.Birthday != nil {
		birthday = user.Birthday.Format("2006-01-02")
	}

	return &LoginResponse{
		ID:             user.ID,
		Username:       user.Name,
		Nickname:       user.Nickname,
		RealName:       user.Name,
		Email:          user.Email,
		Phone:          user.Phone,
		Avatar:         user.Avatar,
		Gender:         user.Gender,
		Birthday:       birthday,
		Bio:            user.Bio,
		Roles:          roleNames,
		RegisterSource: user.RegisterSource,
		AccessToken:    tokenPair.AccessToken,
		RefreshToken:   tokenPair.RefreshToken,
	}, nil
}

// RefreshToken 刷新 AccessToken（带 Token 轮换）
// 验证 RefreshToken 哈希，通过后生成新的 AccessToken + RefreshToken
func (s *AuthService) RefreshToken(userID uint, refreshToken string) (*TokenRefreshResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("refresh_token:%d", userID)

	// 1. 验证 Redis 中的 RefreshToken 哈希
	storedHash, err := database.GetRedis().Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, errors.New("refresh token expired or not found")
	}
	if storedHash != hashToken(refreshToken) {
		return nil, errors.New("refresh token mismatch")
	}

	// 2. 获取用户信息和角色
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	roleNames, err := s.collectRoleNames(user)
	if err != nil {
		return nil, err
	}
	if len(roleNames) == 0 {
		return nil, errors.New("User has no assigned roles")
	}

	// 3. 生成新的 Token 对（轮换，包含所有角色）
	tokenPair, err := jwt.Generate(user.ID, user.Name, roleNames)
	if err != nil {
		return nil, err
	}

	// 4. 用新 RefreshToken 哈希覆盖旧的（旧的自动失效）
	database.GetRedis().Set(ctx, cacheKey, hashToken(tokenPair.RefreshToken), 30*24*time.Hour)

	return &TokenRefreshResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// TokenRefreshResponse Token 刷新响应
type TokenRefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// hashToken 对 token 做 SHA256 哈希，用于安全存储
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// collectRoleNames 收集用户的所有角色名称（直接 + 分组继承）
func (s *AuthService) collectRoleNames(user *model.User) ([]string, error) {
	roleIDSet := make(map[uint]bool)

	directRoles, err := s.userRepo.GetUserRoles(user.ID)
	if err != nil {
		return nil, err
	}
	for _, role := range directRoles {
		roleIDSet[role.ID] = true
	}

	if user.GroupID > 0 {
		groupRoleIDs, err := s.groupRepo.GetGroupRoleIDsRecursive(user.GroupID)
		if err != nil {
			return nil, err
		}
		for _, roleID := range groupRoleIDs {
			roleIDSet[roleID] = true
		}
	}

	roleNames := make([]string, 0, len(roleIDSet))
	for roleID := range roleIDSet {
		role, err := s.roleRepo.GetByID(roleID)
		if err != nil {
			continue
		}
		roleNames = append(roleNames, role.Name)
	}
	return roleNames, nil
}

// permissionCacheKey 生成权限码缓存 key
func permissionCacheKey(userID uint) string {
	return fmt.Sprintf("permission_codes:%d", userID)
}

// GetPermissionCodes 获取权限码列表（带 Redis 缓存）
func (s *AuthService) GetPermissionCodes(userID uint) ([]string, error) {
	ctx := context.Background()
	cacheKey := permissionCacheKey(userID)

	// 1. 尝试从缓存获取
	var cachedCodes []string
	if err := cache.Get(ctx, cacheKey, &cachedCodes); err == nil {
		return cachedCodes, nil
	}

	// 2. 缓存未命中，查询数据库
	codes, err := s.loadPermissionCodesFromDB(userID)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存，10 分钟过期
	_ = cache.Set(ctx, cacheKey, codes, 10*time.Minute)

	return codes, nil
}

// InvalidatePermissionCache 使用户的权限码缓存失效
// 在用户角色/分组变更时调用
func (s *AuthService) InvalidatePermissionCache(userID uint) {
	ctx := context.Background()
	_ = cache.Delete(ctx, permissionCacheKey(userID))
}

// InvalidatePermissionCacheForUsers 批量清除多个用户的权限缓存
func InvalidatePermissionCacheForUsers(userIDs []uint) {
	ctx := context.Background()
	for _, uid := range userIDs {
		_ = cache.Delete(ctx, permissionCacheKey(uid))
	}
}

// loadPermissionCodesFromDB 从数据库加载用户权限码（核心逻辑，无缓存）
func (s *AuthService) loadPermissionCodesFromDB(userID uint) ([]string, error) {
	// 获取用户信息
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 收集所有角色 ID（去重）
	roleIDSet := make(map[uint]bool)

	// 1. 获取用户直接关联的角色
	userRoles, err := s.userRepo.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}
	for _, role := range userRoles {
		roleIDSet[role.ID] = true
	}

	// 2. 获取分组关联的角色（递归向上查找父分组）
	if user.GroupID > 0 {
		groupRoleIDs, err := s.groupRepo.GetGroupRoleIDsRecursive(user.GroupID)
		if err != nil {
			return nil, err
		}
		for _, roleID := range groupRoleIDs {
			roleIDSet[roleID] = true
		}
	}

	// 3. 收集所有权限码（去重），统一使用 JSON 数组 + 逗号分隔双兼容
	codeSet := make(map[string]bool)
	for roleID := range roleIDSet {
		role, err := s.roleRepo.GetByID(roleID)
		if err != nil {
			continue
		}
		if role.Permissions == "" {
			continue
		}
		var codes []string
		if err := json.Unmarshal([]byte(role.Permissions), &codes); err == nil {
			for _, code := range codes {
				codeSet[code] = true
			}
		} else {
			// 兼容逗号分隔格式
			for _, code := range strings.Split(role.Permissions, ",") {
				code = strings.TrimSpace(code)
				if code != "" {
					codeSet[code] = true
				}
			}
		}
	}

	// 转换为切片
	permissionCodes := make([]string, 0, len(codeSet))
	for code := range codeSet {
		permissionCodes = append(permissionCodes, code)
	}

	return permissionCodes, nil
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(userID uint) (*LoginResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 获取用户角色（包括直接角色和分组继承的角色）
	roleIDSet := make(map[uint]bool)

	// 1. 获取用户直接关联的角色
	directRoles, err := s.userRepo.GetUserRoles(user.ID)
	if err != nil {
		return nil, err
	}
	for _, role := range directRoles {
		roleIDSet[role.ID] = true
	}

	// 2. 获取分组关联的角色（递归向上查找父分组）
	if user.GroupID > 0 {
		groupRoleIDs, err := s.groupRepo.GetGroupRoleIDsRecursive(user.GroupID)
		if err != nil {
			return nil, err
		}
		for _, roleID := range groupRoleIDs {
			roleIDSet[roleID] = true
		}
	}

	// 3. 收集所有角色名称
	roleNames := make([]string, 0, len(roleIDSet))
	for roleID := range roleIDSet {
		role, err := s.roleRepo.GetByID(roleID)
		if err != nil {
			continue
		}
		roleNames = append(roleNames, role.Name)
	}

	// 格式化生日
	birthday := ""
	if user.Birthday != nil {
		birthday = user.Birthday.Format("2006-01-02")
	}

	return &LoginResponse{
		ID:       user.ID,
		Username: user.Name,
		Nickname: user.Nickname,
		RealName: user.Name,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Gender:   user.Gender,
		Birthday: birthday,
		Bio:      user.Bio,
		Roles:    roleNames,
	}, nil
}

// RecordLoginDevice 记录登录设备
func (s *AuthService) RecordLoginDevice(userID uint, ip, userAgent, deviceID string) {
	if s.loginDeviceSvc != nil {
		// 优先使用前端传来的 X-Device-ID，没有则用 UA+IP 生成
		if deviceID == "" {
			deviceID = generateDeviceID(userAgent, ip)
		}
		now := time.Now()
		device := &model.LoginDevice{
			UserID:       userID,
			DeviceID:     deviceID,
			DeviceType:   "web",
			DeviceName:   parseDeviceName(userAgent),
			Browser:      parseBrowser(userAgent),
			OS:           parseOS(userAgent),
			IP:           ip,
			Location:     "",
			LastActiveAt: &now,
		}
		_ = s.loginDeviceSvc.CreateOrUpdateDevice(device)
	}
}

func generateDeviceID(userAgent, ip string) string {
	data := []byte(userAgent + ip)
	if len(data) > 16 {
		data = data[:16]
	}
	return fmt.Sprintf("web-%x", data)
}

func parseDeviceName(userAgent string) string {
	browser := parseBrowser(userAgent)
	os := parseOS(userAgent)
	if browser != "Unknown" && os != "Unknown" {
		return browser + " on " + os
	}
	if browser != "Unknown" {
		return browser
	}
	if os != "Unknown" {
		return os
	}
	if len(userAgent) > 50 {
		return userAgent[:50]
	}
	if userAgent == "" {
		return "Unknown Device"
	}
	return userAgent
}

func parseBrowser(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "edg/") || strings.Contains(ua, "edge/"):
		return "Microsoft Edge"
	case strings.Contains(ua, "opr/") || strings.Contains(ua, "opera"):
		return "Opera"
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg"):
		return "Chrome"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		return "Safari"
	default:
		return "Unknown"
	}
}

func parseOS(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "windows nt 10"):
		return "Windows 10/11"
	case strings.Contains(ua, "windows nt 6.3"):
		return "Windows 8.1"
	case strings.Contains(ua, "windows nt 6.1"):
		return "Windows 7"
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac os x") || strings.Contains(ua, "macintosh"):
		return "macOS"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		return "iOS"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "linux"):
		return "Linux"
	default:
		return "Unknown"
	}
}

// RecordSecurityLog 记录安全日志
func (s *AuthService) RecordSecurityLog(userID uint, eventType, detail, ip, userAgent string, status int) {
	if s.securityLogSvc != nil {
		_ = s.securityLogSvc.Record(userID, eventType, detail, ip, userAgent, status)
	}
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Gender   int    `json:"gender"`
	Birthday string `json:"birthday"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
}

// UpdateProfile 更新当前用户个人资料
func (s *AuthService) UpdateProfile(userID uint, req *UpdateProfileRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Gender != 0 {
		user.Gender = req.Gender
	}
	if req.Birthday != "" {
		if t, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			user.Birthday = &t
		}
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	return s.userRepo.Update(user)
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword     string `json:"oldPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
	CaptchaID       string `json:"captchaId"`
	CaptchaCode     string `json:"captchaCode"`
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uint, req *ChangePasswordRequest, ip, userAgent string) error {
	// 验证码校验
	if req.CaptchaID == "" || req.CaptchaCode == "" {
		return fmt.Errorf("请先完成验证码验证")
	}
	valid, _ := captcha.Verify(req.CaptchaID, req.CaptchaCode, 0, nil)
	if !valid {
		return fmt.Errorf("验证码错误或已过期")
	}

	// 校验新密码和确认密码
	if req.NewPassword != req.ConfirmPassword {
		return fmt.Errorf("两次输入的密码不一致")
	}

	// 密码强度校验
	if len(req.NewPassword) < 6 {
		return fmt.Errorf("密码长度不能少于 6 位")
	}
	if len(req.NewPassword) > 128 {
		return fmt.Errorf("密码长度不能超过 128 位")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		s.RecordSecurityLog(userID, "password_change", "旧密码错误", ip, userAgent, 0)
		return fmt.Errorf("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	now := time.Now()
	user.PasswordChangedAt = &now
	user.LoginFailCount = 0

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// 使现有 Token 失效，强制所有设备重新登录
	rdb := database.GetRedis()
	ctx := context.Background()
	rdb.Del(ctx, fmt.Sprintf("refresh_token:%d", userID))

	// 记录安全日志
	s.RecordSecurityLog(userID, "password_change", "修改密码成功", ip, userAgent, 1)

	return nil
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username        string `json:"username" binding:"required,min=3,max=50"`
	Email           string `json:"email" binding:"required,email"`
	EmailCode       string `json:"emailCode" binding:"required,len=6"`
	Password        string `json:"password" binding:"required,min=6,max=128"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// validUsername 验证用户名格式（只允许字母、数字、下划线、连字符）
func validUsername(username string) bool {
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest, ip, userAgent string) (uint, error) {
	// 校验密码一致性
	if req.Password != req.ConfirmPassword {
		return 0, errors.New("两次输入的密码不一致")
	}

	// 校验用户名格式
	if !validUsername(req.Username) {
		return 0, errors.New("用户名只能包含字母、数字、下划线和连字符")
	}

	// 验证邮箱验证码
	verifyCodeService := NewVerifyCodeService()
	valid, err := verifyCodeService.VerifyCode(req.Email, req.EmailCode, "register")
	if err != nil {
		return 0, fmt.Errorf("验证码验证失败: %w", err)
	}
	if !valid {
		return 0, errors.New("验证码错误")
	}

	// 检查用户名是否已存在
	existingUser, _ := s.userRepo.GetByUsername(req.Username)
	if existingUser != nil && existingUser.ID > 0 {
		return 0, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在（一个邮箱只能注册一个账号）
	existingByEmail, _ := s.userRepo.GetByEmail(req.Email)
	if existingByEmail != nil && existingByEmail.ID > 0 {
		return 0, errors.New("该邮箱已被注册，一个邮箱只能绑定一个账号")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := &model.User{
		Name:     req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Status:   1,
		Nickname: req.Username,
	}
	if err := s.userRepo.Create(user); err != nil {
		return 0, fmt.Errorf("创建用户失败: %w", err)
	}

	// 分配默认角色（必须存在 "user" 角色）
	defaultRole, err := s.roleRepo.GetByName("user")
	if err != nil || defaultRole == nil {
		logger.Warn("默认角色 'user' 不存在，新用户将无角色，无法登录")
	} else {
		s.userRepo.SyncUserRoles(user.ID, []uint{defaultRole.ID})
	}


		// 为用户创建头像文件夹
		fileService := NewFileService()
		_, err = fileService.CreateAvatarFolder(user.ID)
		if err != nil {
			logger.Warn(fmt.Sprintf("创建头像文件夹失败: %v", err))
		}
	// 记录安全日志
	s.RecordSecurityLog(user.ID, "register", fmt.Sprintf("新用户注册成功: %s", req.Username), ip, userAgent, 1)

	logger.Info(fmt.Sprintf("新用户注册: %s (%s)", req.Username, req.Email))
	return user.ID, nil
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email           string `json:"email" binding:"required,email"`
	EmailCode       string `json:"emailCode" binding:"required,len=6"`
	NewPassword     string `json:"newPassword" binding:"required,min=6,max=128"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

// ResetPassword 重置密码（通过邮箱验证码）
func (s *AuthService) ResetPassword(req *ResetPasswordRequest, ip, userAgent string) error {
	// 校验密码一致性
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("两次输入的密码不一致")
	}

	// 验证邮箱验证码
	verifyCodeService := NewVerifyCodeService()
	valid, err := verifyCodeService.VerifyCode(req.Email, req.EmailCode, "reset_password")
	if err != nil {
		return fmt.Errorf("验证码验证失败: %w", err)
	}
	if !valid {
		return errors.New("验证码错误")
	}

	// 根据邮箱查找用户
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil || user == nil || user.ID == 0 {
		return errors.New("该邮箱未注册")
	}

	// 检查密码历史（防止重复使用旧密码）
	phs := NewPasswordHistoryService()
	isRepeated, err := phs.CheckRepeated(user.ID, req.NewPassword)
	if err != nil {
		logger.Error("密码历史检查失败", zap.Error(err))
	} else if isRepeated {
		return errors.New("新密码不能与最近5次使用的密码相同")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	now := time.Now()
	user.Password = string(hashedPassword)
	user.PasswordChangedAt = &now
	user.LoginFailCount = 0
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	// === 安全措施：踢出所有登录状态 ===
	rdb := database.GetRedis()
	ctx := context.Background()

	// 1. 清除 refresh token
	rdb.Del(ctx, fmt.Sprintf("refresh_token:%d", user.ID))

	// 2. 清除权限缓存（强制重新加载权限）
	cache.Delete(ctx, permissionCacheKey(user.ID))

	// 3. 清除所有登录设备记录
	s.loginDeviceSvc.DeleteAllDevices(user.ID)

	// 4. 发送安全通知邮件
	siteName := getSiteName()
	notifySubject := fmt.Sprintf("【%s】密码重置通知", siteName)
	notifyBody := fmt.Sprintf(`
		<p>尊敬的 %s：</p>
		<p>您的账号密码已于 %s 成功重置。</p>
		<p>如果不是您本人操作，请立即联系管理员或修改密码。</p>
		<p>此邮件为系统自动发送，请勿回复。</p>
	`, user.Name, now.Format("2006-01-02 15:04:05"))
	email.SendHTMLEmail(req.Email, notifySubject, notifyBody)

	// 记录安全日志
	s.RecordSecurityLog(user.ID, "password_reset", "通过邮箱验证码重置密码，已踢出所有设备", ip, userAgent, 1)

	logger.Info(fmt.Sprintf("用户重置密码: %s (%s)，已踢出所有设备", user.Name, req.Email))
	return nil
}

// getSiteName 从数据库获取站点名称
func getSiteName() string {
	db := database.GetMySQL()
	var value string
	err := db.Raw("SELECT value FROM sys_system_settings WHERE group_key = 'basic' AND `key` = 'site_name' AND deleted_at IS NULL").Scan(&value).Error
	if err != nil || value == "" {
		return "管理系统"
	}
	// 去除 JSON 引号
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	return value
}

// Logout 用户登出
func (s *AuthService) Logout(userID uint) error {
	// 从 Redis 删除 RefreshToken
	rdb := database.GetRedis()
	return rdb.Del(context.Background(), fmt.Sprintf("refresh_token:%d", userID)).Err()
}

// InitDefaultUsers 初始化默认用户
func (s *AuthService) InitDefaultUsers() error {
	// 检查是否已有用户
	var count int64
	database.GetMySQL().Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 创建默认菜单
	if err := s.initDefaultMenus(); err != nil {
		return err
	}

	// 创建默认角色（不指定 ID，让 auto-increment 工作）
	roles := []model.Role{
		{Name: "super", Status: 1, Remark: "超级管理员"},
		{Name: "admin", Status: 1, Remark: "管理员"},
		{Name: "user", Status: 1, Remark: "普通用户"},
	}
	for i := range roles {
		if err := s.roleRepo.Create(&roles[i]); err != nil {
			return err
		}
	}

	// 创建默认用户
	defaultUsers := []struct {
		Name     string
		Password string
		RoleName string
	}{
		{"vben", "123456", "super"},
		{"admin", "123456", "admin"},
		{"jack", "123456", "user"},
	}

	for _, u := range defaultUsers {
		// 加密密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		user := &model.User{
			Name:     u.Name,
			Password: string(hashedPassword),
			Status:   1,
		}
		if err := s.userRepo.Create(user); err != nil {
			return err
		}

		// 关联角色
		for _, role := range roles {
			if role.Name == u.RoleName {
				if err := s.userRepo.SyncUserRoles(user.ID, []uint{role.ID}); err != nil {
					return err
				}
				break
			}
		}
	}

	return nil
}

// initDefaultMenus 初始化默认菜单
func (s *AuthService) initDefaultMenus() error {
	// 创建单个菜单并返回其 ID
	create := func(m *model.Menu) error {
		return s.menuRepo.Create(m)
	}

	// --- 概览 ---
	dashboard := &model.Menu{PID: 0, Name: "Dashboard", Path: "/dashboard", Type: "catalog", Status: 1, Icon: "lucide:layout-dashboard", Meta: `{"order":1,"title":"概览"}`}
	if err := create(dashboard); err != nil {
		return err
	}
	for _, m := range []model.Menu{
		{PID: dashboard.ID, Name: "Analytics", Path: "/analytics", Component: "/dashboard/analytics/index", Type: "menu", Status: 1, Icon: "lucide:area-chart", Meta: `{"order":1,"title":"分析页","affixTab":true}`},
		{PID: dashboard.ID, Name: "Workspace", Path: "/workspace", Component: "/dashboard/workspace/index", Type: "menu", Status: 1, Icon: "lucide:monitor", Meta: `{"order":2,"title":"工作台"}`},
	} {
		if err := create(&m); err != nil {
			return err
		}
	}

	// --- 系统管理 ---
	system := &model.Menu{PID: 0, Name: "System", Path: "/system", Type: "catalog", Status: 1, Icon: "lucide:settings", Meta: `{"order":2,"title":"系统管理"}`}
	if err := create(system); err != nil {
		return err
	}

	systemMenus := []model.Menu{
		{PID: system.ID, Name: "SystemUser", Path: "/system/user", Component: "/system/user/list", Type: "menu", Status: 1, Icon: "lucide:users", AuthCode: "system:user:list", Meta: `{"order":1,"title":"用户管理"}`},
		{PID: system.ID, Name: "SystemRole", Path: "/system/role", Component: "/system/role/list", Type: "menu", Status: 1, Icon: "lucide:shield", AuthCode: "system:role:list", Meta: `{"order":2,"title":"角色管理"}`},
		{PID: system.ID, Name: "SystemMenu", Path: "/system/menu", Component: "/system/menu/list", Type: "menu", Status: 1, Icon: "lucide:list", AuthCode: "system:menu:list", Meta: `{"order":3,"title":"菜单管理"}`},
		{PID: system.ID, Name: "SystemGroup", Path: "/system/group", Component: "/system/group/list", Type: "menu", Status: 1, Icon: "lucide:boxes", AuthCode: "system:group:list", Meta: `{"order":4,"title":"分组管理"}`},
		{PID: system.ID, Name: "SystemSetting", Path: "/system/settings", Component: "/system/settings/index", Type: "menu", Status: 1, Icon: "lucide:sliders-horizontal", AuthCode: "system:setting:list", Meta: `{"order":5,"title":"系统设置"}`},
	}
	for i := range systemMenus {
		if err := create(&systemMenus[i]); err != nil {
			return err
		}
	}

	// 按钮权限
	buttonGroups := []struct {
		parent *model.Menu
		items  []model.Menu
	}{
		{&systemMenus[0], []model.Menu{
			{Name: "SystemUserView", AuthCode: "system:user:view", Meta: `{"title":"查看用户"}`},
			{Name: "SystemUserAdd", AuthCode: "system:user:add", Meta: `{"title":"添加用户"}`},
			{Name: "SystemUserEdit", AuthCode: "system:user:edit", Meta: `{"title":"编辑用户"}`},
			{Name: "SystemUserDelete", AuthCode: "system:user:delete", Meta: `{"title":"删除用户"}`},
		}},
		{&systemMenus[1], []model.Menu{
			{Name: "SystemRoleView", AuthCode: "system:role:view", Meta: `{"title":"查看角色"}`},
			{Name: "SystemRoleAdd", AuthCode: "system:role:add", Meta: `{"title":"添加角色"}`},
			{Name: "SystemRoleEdit", AuthCode: "system:role:edit", Meta: `{"title":"编辑角色"}`},
			{Name: "SystemRoleDelete", AuthCode: "system:role:delete", Meta: `{"title":"删除角色"}`},
		}},
		{&systemMenus[2], []model.Menu{
			{Name: "SystemMenuView", AuthCode: "system:menu:view", Meta: `{"title":"查看菜单"}`},
			{Name: "SystemMenuAdd", AuthCode: "system:menu:add", Meta: `{"title":"添加菜单"}`},
			{Name: "SystemMenuEdit", AuthCode: "system:menu:edit", Meta: `{"title":"编辑菜单"}`},
			{Name: "SystemMenuDelete", AuthCode: "system:menu:delete", Meta: `{"title":"删除菜单"}`},
		}},
		{&systemMenus[3], []model.Menu{
			{Name: "SystemGroupView", AuthCode: "system:group:view", Meta: `{"title":"查看分组"}`},
			{Name: "SystemGroupAdd", AuthCode: "system:group:add", Meta: `{"title":"添加分组"}`},
			{Name: "SystemGroupEdit", AuthCode: "system:group:edit", Meta: `{"title":"编辑分组"}`},
			{Name: "SystemGroupDelete", AuthCode: "system:group:delete", Meta: `{"title":"删除分组"}`},
		}},
		{&systemMenus[4], []model.Menu{
			{Name: "SystemSettingView", AuthCode: "system:setting:list", Meta: `{"title":"查看设置"}`},
			{Name: "SystemSettingEdit", AuthCode: "system:setting:edit", Meta: `{"title":"编辑设置"}`},
		}},
	}
	for _, bg := range buttonGroups {
		for i := range bg.items {
			bg.items[i].PID = bg.parent.ID
			bg.items[i].Type = "button"
			bg.items[i].Status = 1
			if err := create(&bg.items[i]); err != nil {
				return err
			}
		}
	}

	// --- 用户认证 ---
	userAuth := &model.Menu{PID: 0, Name: "UserAuth", Path: "/user-auth", Type: "catalog", Status: 1, Icon: "lucide:shield-check", Meta: `{"order":3,"title":"用户认证"}`}
	if err := create(userAuth); err != nil {
		return err
	}

	authMenus := []model.Menu{
		{PID: userAuth.ID, Name: "SecurityLog", Path: "/user-auth/security-log", Component: "/user-auth/security-log/list", Type: "menu", Status: 1, Icon: "lucide:file-text", AuthCode: "system:security:view", Meta: `{"order":1,"title":"安全日志"}`},
		{PID: userAuth.ID, Name: "LoginDevice", Path: "/user-auth/device", Component: "/user-auth/device/list", Type: "menu", Status: 1, Icon: "lucide:smartphone", AuthCode: "system:device:view", Meta: `{"order":2,"title":"登录设备"}`},
		{PID: userAuth.ID, Name: "OAuthBinding", Path: "/user-auth/oauth", Component: "/user-auth/oauth/list", Type: "menu", Status: 1, Icon: "lucide:link", AuthCode: "system:oauth:view", Meta: `{"order":3,"title":"OAuth绑定"}`},
		{PID: userAuth.ID, Name: "RealName", Path: "/user-auth/real-name", Component: "/user-auth/real-name/list", Type: "menu", Status: 1, Icon: "lucide:user-check", AuthCode: "system:realname:view", Meta: `{"order":4,"title":"实名认证"}`},
		{PID: userAuth.ID, Name: "Privacy", Path: "/user-auth/privacy", Component: "/user-auth/privacy/index", Type: "menu", Status: 1, Icon: "lucide:eye-off", AuthCode: "system:privacy:view", Meta: `{"order":5,"title":"隐私设置"}`},
		{PID: userAuth.ID, Name: "RoleApplication", Path: "/user-auth/role-app", Component: "/user-auth/role-app/list", Type: "menu", Status: 1, Icon: "lucide:clipboard-check", AuthCode: "system:roleapp:view", Meta: `{"order":6,"title":"角色申请"}`},
	}
	for i := range authMenus {
		if err := create(&authMenus[i]); err != nil {
			return err
		}
	}

	// 用户认证按钮权限
	authButtonGroups := []struct {
		parent *model.Menu
		items  []model.Menu
	}{
		{&authMenus[0], []model.Menu{
			{Name: "SecurityLogView", AuthCode: "system:security:view", Meta: `{"title":"查看安全日志"}`},
		}},
		{&authMenus[1], []model.Menu{
			{Name: "LoginDeviceView", AuthCode: "system:device:view", Meta: `{"title":"查看设备"}`},
			{Name: "LoginDeviceDelete", AuthCode: "system:device:delete", Meta: `{"title":"踢出设备"}`},
		}},
		{&authMenus[3], []model.Menu{
			{Name: "RealNameReview", AuthCode: "system:realname:review", Meta: `{"title":"审核实名"}`},
		}},
		{&authMenus[5], []model.Menu{
			{Name: "RoleAppReview", AuthCode: "system:roleapp:review", Meta: `{"title":"审核角色申请"}`},
		}},
	}
	for _, bg := range authButtonGroups {
		for i := range bg.items {
			bg.items[i].PID = bg.parent.ID
			bg.items[i].Type = "button"
			bg.items[i].Status = 1
			if err := create(&bg.items[i]); err != nil {
				return err
			}
		}
	}

	return nil
}
