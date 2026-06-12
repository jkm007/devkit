package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/cache"
	"backend-server/pkg/captcha"
	"backend-server/pkg/database"
	"backend-server/pkg/email"
	"backend-server/pkg/jwt"
	"backend-server/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	userRepo       *repository.UserRepo
	roleRepo       *repository.RoleRepo
	menuRepo       *repository.MenuRepo
	groupRepo      *repository.GroupRepo
	securityLogSvc *SecurityLogService
	loginDeviceSvc *LoginDeviceService
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
	Username    string          `json:"username" binding:"required"`
	Password    string          `json:"password" binding:"required"`
	CaptchaID   string          `json:"captchaId"`   // 验证码 ID
	CaptchaCode string          `json:"captchaCode"` // 验证码值
	CaptchaType string          `json:"captchaType"` // 验证码类型
	Points      []captcha.Point `json:"points"`      // 点选验证码坐标
	StartTime   int64           `json:"startTime"`   // 验证码生成时间（毫秒，用于时间检测）
}

// LoginResponse 登录响应
type LoginResponse struct {
	ID                 uint     `json:"id"`
	Username           string   `json:"username"`
	Nickname           string   `json:"nickname"`
	RealName           string   `json:"realName"`
	Email              string   `json:"email"`
	Phone              string   `json:"phone"`
	Avatar             string   `json:"avatar"`
	Gender             int      `json:"gender"`
	Birthday           string   `json:"birthday,omitempty"`
	Bio                string   `json:"bio"`
	Roles              []string `json:"roles"`
	RegisterSource     string   `json:"registerSource,omitempty"`
	HomePath           string   `json:"homePath,omitempty"`
	AccessToken        string   `json:"accessToken,omitempty"`
	RefreshToken       string   `json:"refreshToken,omitempty"`
	MustChangePassword bool     `json:"mustChangePassword,omitempty"` // 首次登录或管理员重置密码后需强制修改
}

// loginFailCountKeyByIP 生成基于 IP 维度的登录失败计数 Redis key
func loginFailCountKeyByIP(clientIP string) string {
	return fmt.Sprintf("login_fail:ip:%s", clientIP)
}

// loginFailCountKeyByUsername 生成基于账号维度的登录失败计数 Redis key
func loginFailCountKeyByUsername(username string) string {
	return fmt.Sprintf("login_fail:user:%s", username)
}

// captchaSettingsCacheItem 验证码配置内存缓存项
type captchaSettingsCacheItem struct {
	enabled   bool
	trigger   int
	expiresAt time.Time
}

var (
	// 验证码配置内存缓存（避免每次登录都查数据库）
	captchaSettingsCache    *captchaSettingsCacheItem
	captchaSettingsCacheMu  sync.RWMutex
	captchaSettingsCacheTTL = 5 * time.Minute
)

// getCaptchaSettings 获取验证码设置（带内存缓存，TTL 5 分钟）
func (s *AuthService) getCaptchaSettings() (enabled bool, trigger int) {
	// 先从内存缓存读取
	captchaSettingsCacheMu.RLock()
	if captchaSettingsCache != nil && time.Now().Before(captchaSettingsCache.expiresAt) {
		enabled = captchaSettingsCache.enabled
		trigger = captchaSettingsCache.trigger
		captchaSettingsCacheMu.RUnlock()
		return
	}
	captchaSettingsCacheMu.RUnlock()

	// 缓存未命中或已过期，从数据库加载
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

	// 写入内存缓存
	captchaSettingsCacheMu.Lock()
	captchaSettingsCache = &captchaSettingsCacheItem{
		enabled:   enabled,
		trigger:   trigger,
		expiresAt: time.Now().Add(captchaSettingsCacheTTL),
	}
	captchaSettingsCacheMu.Unlock()

	return
}

// InvalidateCaptchaSettingsCache 使验证码配置内存缓存失效
// 在管理员修改验证码配置后调用，确保下次登录读取最新配置
func InvalidateCaptchaSettingsCache() {
	captchaSettingsCacheMu.Lock()
	captchaSettingsCache = nil
	captchaSettingsCacheMu.Unlock()
}

// checkCaptcha 验证码校验
func (s *AuthService) checkCaptcha(req *LoginRequest, clientIP string) error {
	enabled, trigger := s.getCaptchaSettings()
	if !enabled {
		return nil
	}

	// 判断是否需要验证码（IP 和账号任一维度达到阈值即触发）
	needCaptcha := false
	if trigger == 0 {
		needCaptcha = true
	} else {
		rdb := database.GetRedis()
		ctx := context.Background()

		// IP 维度检查
		ipKey := loginFailCountKeyByIP(clientIP)
		countStr, err := rdb.Get(ctx, ipKey).Result()
		if err == nil {
			count, _ := strconv.Atoi(countStr)
			if count >= trigger {
				needCaptcha = true
			}
		}

		// 账号维度检查（如果 IP 维度未触发）
		if !needCaptcha && req.Username != "" {
			userKey := loginFailCountKeyByUsername(req.Username)
			countStr, err = rdb.Get(ctx, userKey).Result()
			if err == nil {
				count, _ := strconv.Atoi(countStr)
				if count >= trigger {
					needCaptcha = true
				}
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

// recordLoginFail 记录登录失败（IP + 账号双维度）
func (s *AuthService) recordLoginFail(clientIP, username string) {
	rdb := database.GetRedis()
	ctx := context.Background()

	// IP 维度
	ipKey := loginFailCountKeyByIP(clientIP)
	rdb.Incr(ctx, ipKey)
	rdb.Expire(ctx, ipKey, 30*time.Minute)

	// 账号维度
	if username != "" {
		userKey := loginFailCountKeyByUsername(username)
		rdb.Incr(ctx, userKey)
		rdb.Expire(ctx, userKey, 30*time.Minute)
	}
}

// clearLoginFail 清除登录失败计数（IP + 账号双维度）
func (s *AuthService) clearLoginFail(clientIP, username string) {
	rdb := database.GetRedis()
	ctx := context.Background()
	rdb.Del(ctx, loginFailCountKeyByIP(clientIP))
	if username != "" {
		rdb.Del(ctx, loginFailCountKeyByUsername(username))
	}
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
			s.recordLoginFail(clientIP, req.Username)
			return nil, errors.New("Username or password is incorrect.")
		}
		logger.Error("登录时数据库查询失败", zap.Error(err))
		return nil, errors.New("Username or password is incorrect.")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.recordLoginFail(clientIP, req.Username)
		return nil, errors.New("Username or password is incorrect.")
	}

	// 检查状态
	if user.Status != 1 {
		return nil, errors.New("Account is disabled")
	}

	// 登录成功，清除 IP + 账号双维度的失败计数
	s.clearLoginFail(clientIP, req.Username)

	return s.generateLoginResponse(user, clientIP)
}

// LoginByEmail 邮箱验证码登录
func (s *AuthService) LoginByEmail(email, code, clientIP string) (*LoginResponse, error) {
	// 验证邮箱验证码
	verifyCodeService := NewVerifyCodeService()
	valid, err := verifyCodeService.VerifyCode(email, code, "login")
	if err != nil {
		s.recordLoginFail(clientIP, email)
		return nil, fmt.Errorf("验证码验证失败: %w", err)
	}
	if !valid {
		s.recordLoginFail(clientIP, email)
		return nil, errors.New("验证码错误")
	}

	// 根据邮箱查找用户
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.recordLoginFail(clientIP, email)
			return nil, errors.New("该邮箱未注册")
		}
		logger.Error("邮箱登录时数据库查询失败", zap.Error(err))
		return nil, errors.New("系统错误，请稍后重试")
	}
	if user == nil || user.ID == 0 {
		s.recordLoginFail(clientIP, email)
		return nil, errors.New("该邮箱未注册")
	}

	// 检查状态
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	s.clearLoginFail(clientIP, email)
	return s.generateLoginResponse(user, clientIP)
}

// LoginByPhone 手机号验证码登录
func (s *AuthService) LoginByPhone(phone, code, clientIP string) (*LoginResponse, error) {
	// 验证短信验证码
	verifyCodeService := NewVerifyCodeService()
	valid, err := verifyCodeService.VerifySMSCode(phone, code, "login")
	if err != nil {
		s.recordLoginFail(clientIP, phone)
		return nil, fmt.Errorf("验证码验证失败: %w", err)
	}
	if !valid {
		s.recordLoginFail(clientIP, phone)
		return nil, errors.New("验证码错误")
	}

	// 根据手机号查找用户
	user, err := s.userRepo.GetByPhone(phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.recordLoginFail(clientIP, phone)
			return nil, errors.New("该手机号未注册")
		}
		logger.Error("手机登录时数据库查询失败", zap.Error(err))
		return nil, errors.New("系统错误，请稍后重试")
	}
	if user == nil || user.ID == 0 {
		s.recordLoginFail(clientIP, phone)
		return nil, errors.New("该手机号未注册")
	}

	// 检查状态
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	s.clearLoginFail(clientIP, phone)
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
		// 自动分配默认 user 角色（与 OAuth 登录保持一致）
		defaultRole, err := s.roleRepo.GetByName("user")
		if err == nil && defaultRole != nil {
			_ = s.userRepo.SyncUserRoles(user.ID, []uint{defaultRole.ID})
			roleNames = []string{"user"}
		}
		if len(roleNames) == 0 {
			return nil, errors.New("User has no assigned roles")
		}
	}

	// 根据角色设置默认首页
	homePath := "/user-home"
	for _, name := range roleNames {
		if name == "admin" || name == "super" {
			homePath = "/analytics"
			break
		}
	}

	// 生成 Token
	tokenPair, err := jwt.Generate(user.ID, user.Name, roleNames)
	if err != nil {
		return nil, err
	}

	// 存储 RefreshToken 哈希到 Redis（使用 Lua 脚本确保存储原子性，避免与并发 Refresh 操作产生竞态）
	ctx := context.Background()
	refreshCacheKey := fmt.Sprintf("refresh_token:%d", user.ID)
	refreshTTLSeconds := int((30 * 24 * time.Hour).Seconds())
	storeResult, err := storeRefreshTokenLuaScript.Run(ctx, database.GetRedis(), []string{refreshCacheKey}, hashToken(tokenPair.RefreshToken), refreshTTLSeconds).Int()
	if err != nil {
		logger.Error("存储 RefreshToken 失败", zap.Error(err))
		return nil, errors.New("系统错误，请稍后重试")
	}
	if storeResult == 0 {
		// 覆盖了已有 RefreshToken，可能是因为并发登录或 Token 轮换冲突
		logger.Warn("登录时覆盖已有 RefreshToken", zap.Uint("userID", user.ID))
	}

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
		ID:                 user.ID,
		Username:           user.Name,
		Nickname:           user.Nickname,
		RealName:           user.Name,
		Email:              user.Email,
		Phone:              user.Phone,
		Avatar:             user.Avatar,
		Gender:             user.Gender,
		Birthday:           birthday,
		Bio:                user.Bio,
		Roles:              roleNames,
		RegisterSource:     user.RegisterSource,
		HomePath:           homePath,
		AccessToken:        tokenPair.AccessToken,
		RefreshToken:       tokenPair.RefreshToken,
		MustChangePassword: user.MustChangePassword,
	}, nil
}

// RefreshToken 刷新 AccessToken（带 Token 轮换）
// 使用 Lua 脚本原子性完成：验证旧哈希 → 替换为新哈希
// 并发刷新请求只有一个能成功，其余自动失败（防止 Token 重放攻击）
func (s *AuthService) RefreshToken(userID uint, refreshToken string) (*TokenRefreshResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("refresh_token:%d", userID)

	// 1. 获取用户信息和角色（在原子操作之前，减少 Lua 脚本执行时间）
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	roleNames, err := s.collectRoleNames(user)
	if err != nil {
		return nil, err
	}
	if len(roleNames) == 0 {
		// 自动分配默认 user 角色
		defaultRole, err := s.roleRepo.GetByName("user")
		if err == nil && defaultRole != nil {
			_ = s.userRepo.SyncUserRoles(user.ID, []uint{defaultRole.ID})
			roleNames = []string{"user"}
		}
		if len(roleNames) == 0 {
			return nil, errors.New("User has no assigned roles")
		}
	}

	// 2. 生成新的 Token 对（轮换，包含所有角色）
	tokenPair, err := jwt.Generate(user.ID, user.Name, roleNames)
	if err != nil {
		return nil, err
	}

	// 3. 使用 Lua 脚本原子性地验证旧哈希并替换为新哈希
	//    这确保了 READ-VERIFY-WRITE 的原子性，避免 TOCTOU 竞态条件
	oldHash := hashToken(refreshToken)
	newHash := hashToken(tokenPair.RefreshToken)
	ttlSeconds := int((30 * 24 * time.Hour).Seconds())

	result, err := refreshTokenLuaScript.Run(ctx, database.GetRedis(), []string{cacheKey}, oldHash, newHash, ttlSeconds).Int()
	if err != nil {
		logger.Error("原子性更新 RefreshToken 失败", zap.Error(err))
		return nil, errors.New("系统错误，请稍后重试")
	}

	switch result {
	case 1:
		// 成功：旧哈希匹配，已原子替换为新哈希
	case 0:
		// 哈希不匹配：可能已被并发请求使用（Token 重放），或被篡改
		return nil, errors.New("refresh token mismatch")
	case -1:
		// Key 不存在：RefreshToken 已过期或已登出
		return nil, errors.New("refresh token expired or not found")
	default:
		logger.Error("Lua 脚本返回未知结果", zap.Int("result", result))
		return nil, errors.New("系统错误，请稍后重试")
	}

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

// refreshTokenLuaScript 原子性刷新 RefreshToken 的 Lua 脚本
// 功能：读取存储的哈希 → 与旧哈希比较 → 匹配则原子替换为新哈希 → 返回结果
// 避免 TOCTOU 竞态：并发刷新请求只有一个能成功，其余自动失效
// KEYS[1] = refresh_token:{userID}
// ARGV[1] = 旧 refresh token 哈希
// ARGV[2] = 新 refresh token 哈希
// ARGV[3] = 过期时间（秒）
// 返回: 1 = 成功, 0 = 哈希不匹配, -1 = key 不存在
var refreshTokenLuaScript = redis.NewScript(`
local stored = redis.call('GET', KEYS[1])
if not stored then
	return -1
end
if stored ~= ARGV[1] then
	return 0
end
redis.call('SETEX', KEYS[1], tonumber(ARGV[3]), ARGV[2])
return 1
`)

// storeRefreshTokenLuaScript 原子性存储 RefreshToken 的 Lua 脚本
// 功能：检查 key 是否已存在 → 原子性地写入新哈希并设置过期时间
// 登录场景下新会话覆盖旧会话，通过 EXISTS 检测是否覆盖了已有值
// KEYS[1] = refresh_token:{userID}
// ARGV[1] = refresh token 哈希
// ARGV[2] = 过期时间（秒）
// 返回: 1 = 新存储, 0 = 覆盖已有值
var storeRefreshTokenLuaScript = redis.NewScript(`
local existed = redis.call('EXISTS', KEYS[1])
redis.call('SETEX', KEYS[1], tonumber(ARGV[2]), ARGV[1])
if existed == 1 then
	return 0
end
return 1
`)

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

	// 批量查询角色（消除 N+1 查询）
	roleIDs := make([]uint, 0, len(roleIDSet))
	for roleID := range roleIDSet {
		roleIDs = append(roleIDs, roleID)
	}
	roles, err := s.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
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

	// 先从缓存读取
	var codes []string
	if err := cache.Get(ctx, cacheKey, &codes); err == nil {
		return codes, nil
	}

	// 缓存未命中，从数据库加载
	codes, err := s.loadPermissionCodesFromDB(userID)
	if err != nil {
		return nil, err
	}

	// 写入缓存
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

	// 3. 批量查询角色并收集所有权限码（去重），消除 N+1 查询
	roleIDs := make([]uint, 0, len(roleIDSet))
	for roleID := range roleIDSet {
		roleIDs = append(roleIDs, roleID)
	}
	roles, err := s.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		return nil, err
	}

	codeSet := make(map[string]bool)
	for _, role := range roles {
		if role.Permissions == "" {
			continue
		}
		// 权限码统一使用 JSON 数组格式，如 ["system:user:view","system:user:add"]
		var codes []string
		if err := json.Unmarshal([]byte(role.Permissions), &codes); err != nil {
			logger.Warn("角色权限码格式错误，跳过该角色",
				zap.Uint("roleID", role.ID),
				zap.String("permissions", role.Permissions),
				zap.Error(err))
			continue
		}
		for _, code := range codes {
			codeSet[code] = true
		}
	}

	// 转换为切片
	permissionCodes := make([]string, 0, len(codeSet))
	for code := range codeSet {
		permissionCodes = append(permissionCodes, code)
	}

	fmt.Printf("[auth] 权限码加载完成: total_unique_codes=%d\n", len(permissionCodes))
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

	// 3. 批量查询角色并收集所有角色名称（消除 N+1 查询）
	roleIDs := make([]uint, 0, len(roleIDSet))
	for roleID := range roleIDSet {
		roleIDs = append(roleIDs, roleID)
	}
	roles, err := s.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	// 格式化生日
	birthday := ""
	if user.Birthday != nil {
		birthday = user.Birthday.Format("2006-01-02")
	}

	// 根据角色设置默认首页
	homePath := "/user-home"
	for _, name := range roleNames {
		if name == "admin" || name == "super" {
			homePath = "/analytics"
			break
		}
	}

	return &LoginResponse{
		ID:                 user.ID,
		Username:           user.Name,
		Nickname:           user.Nickname,
		RealName:           user.Name,
		Email:              user.Email,
		Phone:              user.Phone,
		Avatar:             user.Avatar,
		Gender:             user.Gender,
		Birthday:           birthday,
		Bio:                user.Bio,
		Roles:              roleNames,
		HomePath:           homePath,
		MustChangePassword: user.MustChangePassword,
	}, nil
}

// DeviceMeta 登录设备扩展信息
// H5/App/小程序可通过 Header 上报这些字段，用于设备管理和审计展示。
type DeviceMeta struct {
	AppVersion    string
	SystemVersion string
	DeviceModel   string
	Platform      string
	Channel       string
}

// RecordLoginDevice 记录登录设备
func (s *AuthService) RecordLoginDevice(userID uint, ip, userAgent, deviceID, clientType string, meta DeviceMeta) {
	if s.loginDeviceSvc != nil {
		// 优先使用前端传来的 X-Device-ID，没有则用 UA+IP 生成
		if deviceID == "" {
			deviceID = generateDeviceID(userAgent, ip)
		}
		clientType = normalizeClientType(clientType)
		platform := normalizePlatform(meta.Platform, clientType)
		now := time.Now()
		device := &model.LoginDevice{
			UserID:        userID,
			DeviceID:      deviceID,
			DeviceType:    clientType,
			DeviceName:    parseDeviceName(userAgent),
			Browser:       parseBrowser(userAgent),
			OS:            parseOS(userAgent),
			IP:            ip,
			Location:      "",
			AppVersion:    trimDeviceMeta(meta.AppVersion, 50),
			SystemVersion: trimDeviceMeta(meta.SystemVersion, 100),
			DeviceModel:   trimDeviceMeta(meta.DeviceModel, 100),
			Platform:      platform,
			Channel:       trimDeviceMeta(meta.Channel, 50),
			LastActiveAt:  &now,
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
// Gender 和 Bio 使用指针类型，以区分「用户显式设置零值」和「用户未提供该字段」。
// nil 表示未提供（不更新），非 nil 表示显式设置（即使是零值也会更新）。
type UpdateProfileRequest struct {
	Nickname string  `json:"nickname"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Gender   *int    `json:"gender"` // nil=不更新, 非nil=显式设置（0=未知 1=男 2=女）
	Birthday string  `json:"birthday"`
	Bio      *string `json:"bio"` // nil=不更新, 非nil=显式设置（包括清空）
	Avatar   string  `json:"avatar"`
}

// validateAvatarURL 验证头像 URL 格式（只允许相对路径或本站 URL）
func validateAvatarURL(url string) bool {
	if url == "" {
		return true // 空值允许（不清除头像）
	}
	// 允许相对路径
	if strings.HasPrefix(url, "/uploads/") ||
		strings.HasPrefix(url, "/api/v1/files/") {
		return true
	}
	// 允许 http/https 开头的 URL（外部头像如 OAuth、云存储 presigned URL）
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return true
	}
	return false
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
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.Birthday != "" {
		if t, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			user.Birthday = &t
		}
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.Avatar != "" {
		if !validateAvatarURL(req.Avatar) {
			return fmt.Errorf("头像 URL 格式不正确")
		}
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
// accessToken 参数为当前 access token 字符串，修改密码后会加入黑名单使其立即失效
func (s *AuthService) ChangePassword(userID uint, req *ChangePasswordRequest, ip, userAgent string, accessToken string) error {
	// 验证码校验（/auth/ 路径被 CaptchaGuard 跳过，此处自行校验）
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

	// 保存旧密码到历史记录
	phs := NewPasswordHistoryService()
	if err := phs.SavePassword(userID, user.Password); err != nil {
		// 记录日志但不阻断密码修改流程
		fmt.Printf("[auth] 保存密码历史记录失败: userID=%d, err=%v\n", userID, err)
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
	user.MustChangePassword = false // 解除强制修改密码标记

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// 使现有 Token 失效，强制所有设备重新登录
	rdb := database.GetRedis()
	ctx := context.Background()

	// 将当前 access token 加入黑名单（立即失效）
	if accessToken != "" {
		claims, err := jwt.Parse(accessToken)
		if err == nil && claims.ExpiresAt != nil {
			remaining := time.Until(claims.ExpiresAt.Time)
			if remaining > 0 {
				// 使用 SHA-256 哈希存储，减少内存占用并降低 Token 泄露风险
				tokenHash := sha256.Sum256([]byte(accessToken))
				blacklistKey := fmt.Sprintf("token_blacklist:%s", hex.EncodeToString(tokenHash[:]))
				rdb.Set(ctx, blacklistKey, "1", remaining)
			}
		}
	}

	// 清除 refresh token
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
	RegisterSource  string `json:"registerSource"`
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

func normalizeClientType(clientType string) string {
	clientType = strings.ToLower(strings.TrimSpace(clientType))
	switch clientType {
	case "web", "h5", "app", "miniapp":
		return clientType
	default:
		return "web"
	}
}

func normalizePlatform(platform, fallback string) string {
	platform = strings.ToLower(strings.TrimSpace(platform))
	switch platform {
	case "web", "h5", "ios", "android", "miniapp", "windows", "macos", "linux":
		return platform
	default:
		return normalizeClientType(fallback)
	}
}

func trimDeviceMeta(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) > max {
		return value[:max]
	}
	return value
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

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("密码加密失败: %w", err)
	}

	registerSource := normalizeClientType(req.RegisterSource)

	// 创建用户（事务保护：用户名/邮箱唯一性检查 + 用户创建 + 角色分配必须原子）
	user := &model.User{
		Name:           req.Username,
		Email:          req.Email,
		Password:       string(hashedPassword),
		Status:         1,
		Nickname:       req.Username,
		RegisterSource: registerSource,
	}

	db := database.GetMySQL()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 在事务内检查用户名是否已存在（避免并发注册导致的 TOCTOU 竞态）
		var existingUser model.User
		if err := tx.Where("name = ?", req.Username).First(&existingUser).Error; err == nil {
			return errors.New("用户名已存在")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("检查用户名失败: %w", err)
		}

		// 在事务内检查邮箱是否已存在（一个邮箱只能注册一个账号）
		var existingByEmail model.User
		if err := tx.Where("email = ?", req.Email).First(&existingByEmail).Error; err == nil {
			return errors.New("该邮箱已被注册，一个邮箱只能绑定一个账号")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("检查邮箱失败: %w", err)
		}

		// 创建用户
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("创建用户失败: %w", err)
		}

		// 分配默认角色（必须存在 "user" 角色）
		var defaultRole model.Role
		if err := tx.Where("name = ?", "user").First(&defaultRole).Error; err != nil {
			logger.Warn("默认角色 'user' 不存在，新用户将无角色，无法登录")
		} else {
			// 直接在事务中创建用户-角色关联
			userRole := &model.UserRole{
				UserID: user.ID,
				RoleID: defaultRole.ID,
			}
			if err := tx.Create(userRole).Error; err != nil {
				return fmt.Errorf("分配默认角色失败: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	// 为用户创建头像文件夹（非核心操作，失败不回滚）
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

	// 4. 发送安全通知邮件（失败仅记录日志，不阻断主流程）
	siteName := getSiteName()
	notifySubject := fmt.Sprintf("【%s】密码重置通知", siteName)
	notifyBody := fmt.Sprintf(`
		<p>尊敬的 %s：</p>
		<p>您的账号密码已于 %s 成功重置。</p>
		<p>如果不是您本人操作，请立即联系管理员或修改密码。</p>
		<p>此邮件为系统自动发送，请勿回复。</p>
	`, user.Name, now.Format("2006-01-02 15:04:05"))
	if err := email.SendHTMLEmail(req.Email, notifySubject, notifyBody); err != nil {
		logger.Error("密码重置通知邮件发送失败", zap.String("email", req.Email), zap.Error(err))
	}

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

// InitDefaultSettings 初始化默认验证码配置
// 如果 sys_system_settings 表中 captcha 组无数据，则写入默认值（captcha_enabled 默认为 true）
func InitDefaultSettings() error {
	db := database.GetMySQL()

	// 检查是否已有 captcha 配置
	var count int64
	if err := db.Model(&model.SystemSetting{}).Where("group_key = ?", "captcha").Count(&count).Error; err != nil {
		return fmt.Errorf("查询验证码配置失败: %w", err)
	}
	if count > 0 {
		return nil // 已有配置，不覆盖
	}

	// 默认验证码配置项（captcha_enabled 默认开启）
	defaults := []model.SystemSetting{
		{GroupKey: "captcha", Key: "captcha_enabled", Value: "true", Label: "启用验证码", Type: "boolean", Tip: "登录时是否需要验证码", Sort: 1, IsPublic: 1},
		{GroupKey: "captcha", Key: "captcha_type", Value: `"slider"`, Label: "验证码类型", Type: "select", Options: `[{"label":"滑块","value":"slider"},{"label":"拼图","value":"puzzle"},{"label":"旋转","value":"rotation"},{"label":"点选","value":"point"},{"label":"数字","value":"numeric"}]`, Tip: "验证码展示形式", Sort: 2, IsPublic: 1},
		{GroupKey: "captcha", Key: "captcha_expire", Value: "120", Label: "验证码有效期(秒)", Type: "number", Tip: "验证码有效时间", Sort: 3},
		{GroupKey: "captcha", Key: "captcha_max_fail", Value: "5", Label: "最大失败次数", Type: "number", Tip: "连续验证失败后刷新验证码", Sort: 4},
		{GroupKey: "captcha", Key: "captcha_login_trigger", Value: "3", Label: "触发阈值", Type: "number", Tip: "登录失败几次后开始要求验证码（0=始终开启）", Sort: 5},
		{GroupKey: "captcha", Key: "captcha_min_duration", Value: "500", Label: "最短操作时间(ms)", Type: "number", Tip: "操作时间小于此值判定为机器人", Sort: 6},
	}

	// 批量插入
	return db.Transaction(func(tx *gorm.DB) error {
		for i := range defaults {
			if err := tx.Create(&defaults[i]).Error; err != nil {
				return fmt.Errorf("插入验证码配置 %s 失败: %w", defaults[i].Key, err)
			}
		}
		return nil
	})
}

// InitDefaultUsers 初始化默认用户
func (s *AuthService) InitDefaultUsers() error {
	// 检查是否已有用户
	var count int64
	database.GetMySQL().Model(&model.User{}).Count(&count)
	if count > 0 {
		// 角色申请菜单是后续新增功能，已有数据库也需要补齐菜单与权限按钮
		return s.ensureRoleApplicationMenus()
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
	// 优先从环境变量读取密码（格式: DEFAULT_USER_{大写用户名}_PASSWORD），
	// 未设置时生成随机密码并打印到日志，所有默认用户首次登录必须修改密码
	defaultUsers := []struct {
		Name     string
		RoleName string
	}{
		{"vben", "super"},
		{"admin", "admin"},
		{"jack", "user"},
	}

	for _, u := range defaultUsers {
		password := getDefaultUserPassword(u.Name)

		// 加密密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		user := &model.User{
			Name:               u.Name,
			Password:           string(hashedPassword),
			Status:             1,
			MustChangePassword: true, // 首次登录必须修改密码
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

func (s *AuthService) ensureRoleApplicationMenus() error {
	system, err := s.ensureMenu(model.Menu{Name: "System", Path: "/system", Type: "catalog", Status: 1, Icon: "lucide:settings", Meta: `{"order":2,"title":"系统管理"}`}, 0)
	if err != nil {
		return err
	}

	if legacy, err := s.menuRepo.GetByName("RoleApplication"); err == nil && legacy != nil {
		if _, newErr := s.menuRepo.GetByName("SystemRoleApplication"); newErr == nil {
			if err := s.menuRepo.UpdateFields(legacy.ID, map[string]interface{}{"status": 0}); err != nil {
				return err
			}
		} else {
			if err := s.menuRepo.UpdateFields(legacy.ID, map[string]interface{}{
				"name":      "SystemRoleApplication",
				"pid":       system.ID,
				"path":      "/system/role-application",
				"component": "/system/role-application/list",
				"type":      "menu",
				"auth_code": "system:roleapp:view",
				"icon":      "lucide:clipboard-check",
				"meta":      `{"order":6,"title":"角色申请"}`,
				"status":    1,
			}); err != nil {
				return err
			}
		}
	}

	roleApp, err := s.ensureMenu(model.Menu{Name: "SystemRoleApplication", Path: "/system/role-application", Component: "/system/role-application/list", Type: "menu", Status: 1, Icon: "lucide:clipboard-check", AuthCode: "system:roleapp:view", Meta: `{"order":6,"title":"角色申请"}`}, system.ID)
	if err != nil {
		return err
	}

	if _, err := s.ensureMenu(model.Menu{Name: "SystemRoleApplicationView", Type: "button", Status: 1, AuthCode: "system:roleapp:view", Meta: `{"title":"查看角色申请"}`}, roleApp.ID); err != nil {
		return err
	}
	_, err = s.ensureMenu(model.Menu{Name: "SystemRoleApplicationReview", Type: "button", Status: 1, AuthCode: "system:roleapp:review", Meta: `{"title":"审核角色申请"}`}, roleApp.ID)
	return err
}

func (s *AuthService) ensureMenu(menu model.Menu, pid uint) (*model.Menu, error) {
	menu.PID = pid
	existing, err := s.menuRepo.GetByName(menu.Name)
	if err == nil && existing != nil {
		updates := map[string]interface{}{
			"pid":       menu.PID,
			"path":      menu.Path,
			"component": menu.Component,
			"type":      menu.Type,
			"status":    menu.Status,
			"auth_code": menu.AuthCode,
			"icon":      menu.Icon,
			"meta":      menu.Meta,
		}
		if err := s.menuRepo.UpdateFields(existing.ID, updates); err != nil {
			return nil, err
		}
		existing.PID = menu.PID
		existing.Path = menu.Path
		existing.Component = menu.Component
		existing.Type = menu.Type
		existing.Status = menu.Status
		existing.AuthCode = menu.AuthCode
		existing.Icon = menu.Icon
		existing.Meta = menu.Meta
		return existing, nil
	}

	if err := s.menuRepo.Create(&menu); err != nil {
		return nil, err
	}
	return &menu, nil
}

// getDefaultUserPassword 获取默认用户的密码
// 优先级：环境变量 DEFAULT_USER_{大写用户名}_PASSWORD > 随机生成
func getDefaultUserPassword(username string) string {
	envKey := "DEFAULT_USER_" + strings.ToUpper(username) + "_PASSWORD"
	if pw := strings.TrimSpace(os.Getenv(envKey)); pw != "" {
		logger.Info(fmt.Sprintf("使用环境变量 %s 为用户 %s 设置密码", envKey, username))
		return pw
	}

	// 随机生成 16 字符密码（保证强度）
	randomPw, err := generateRandomPassword(16)
	if err != nil {
		logger.Error("生成随机密码失败，使用兜底密码", zap.Error(err))
		return "ChangeMe!" + fmt.Sprintf("%d", time.Now().UnixNano()%10000)
	}

	logger.Warn(fmt.Sprintf("用户 %s 的密码已随机生成，请查看日志并妥善保管（首次登录需修改密码）", username),
		zap.String("password", randomPw))
	return randomPw
}

// generateRandomPassword 生成指定长度的随机密码（包含大小写字母、数字、特殊字符）
func generateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	// 确保包含至少一个大写、小写、数字和特殊字符
	b[0] = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"[int(b[0])%26]
	b[1] = "abcdefghijklmnopqrstuvwxyz"[int(b[1])%26]
	b[2] = "0123456789"[int(b[2])%10]
	b[3] = "!@#$%^&*"[int(b[3])%8]
	return string(b), nil
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
		{PID: system.ID, Name: "SystemUser", Path: "/system/user", Component: "/system/user/list", Type: "menu", Status: 1, Icon: "lucide:users", AuthCode: "system:user:view", Meta: `{"order":1,"title":"用户管理"}`},
		{PID: system.ID, Name: "SystemRole", Path: "/system/role", Component: "/system/role/list", Type: "menu", Status: 1, Icon: "lucide:shield", AuthCode: "system:role:view", Meta: `{"order":2,"title":"角色管理"}`},
		{PID: system.ID, Name: "SystemMenu", Path: "/system/menu", Component: "/system/menu/list", Type: "menu", Status: 1, Icon: "lucide:list", AuthCode: "system:menu:view", Meta: `{"order":3,"title":"菜单管理"}`},
		{PID: system.ID, Name: "SystemGroup", Path: "/system/group", Component: "/system/group/list", Type: "menu", Status: 1, Icon: "lucide:boxes", AuthCode: "system:group:view", Meta: `{"order":4,"title":"分组管理"}`},
		{PID: system.ID, Name: "SystemSetting", Path: "/system/settings", Component: "/system/settings/index", Type: "menu", Status: 1, Icon: "lucide:sliders-horizontal", AuthCode: "system:setting:list", Meta: `{"order":5,"title":"系统设置"}`},
		{PID: system.ID, Name: "SystemRoleApplication", Path: "/system/role-application", Component: "/system/role-application/list", Type: "menu", Status: 1, Icon: "lucide:clipboard-check", AuthCode: "system:roleapp:view", Meta: `{"order":6,"title":"角色申请"}`},
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
		{&systemMenus[5], []model.Menu{
			{Name: "SystemRoleApplicationView", AuthCode: "system:roleapp:view", Meta: `{"title":"查看角色申请"}`},
			{Name: "SystemRoleApplicationReview", AuthCode: "system:roleapp:review", Meta: `{"title":"审核角色申请"}`},
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

	// --- 文件管理 ---
	fileMgmt := &model.Menu{PID: 0, Name: "FileMgmt", Path: "/file", Type: "catalog", Status: 1, Icon: "lucide:folder", Meta: `{"order":4,"title":"文件管理"}`}
	if err := create(fileMgmt); err != nil {
		return err
	}

	fileMenus := []model.Menu{
		{PID: fileMgmt.ID, Name: "FileList", Path: "/file/list", Component: "/file/list/index", Type: "menu", Status: 1, Icon: "lucide:files", AuthCode: "file:list", Meta: `{"order":1,"title":"文件列表"}`},
		{PID: fileMgmt.ID, Name: "FileShare", Path: "/file/share", Component: "/file/share/index", Type: "menu", Status: 1, Icon: "lucide:share-2", AuthCode: "file:list", Meta: `{"order":2,"title":"分享管理"}`},
	}
	for i := range fileMenus {
		if err := create(&fileMenus[i]); err != nil {
			return err
		}
	}

	// 文件管理按钮权限
	fileButtonGroups := []struct {
		parent *model.Menu
		items  []model.Menu
	}{
		{&fileMenus[0], []model.Menu{
			{Name: "FileListView", AuthCode: "file:view:own", Meta: `{"title":"查看自己的文件"}`},
			{Name: "FileListViewAll", AuthCode: "file:view:all", Meta: `{"title":"查看所有文件"}`},
			{Name: "FileUpload", AuthCode: "file:upload", Meta: `{"title":"上传文件"}`},
			{Name: "FileDownload", AuthCode: "file:download", Meta: `{"title":"下载文件"}`},
			{Name: "FileDelete", AuthCode: "file:delete", Meta: `{"title":"删除文件"}`},
			{Name: "FileShareCreate", AuthCode: "file:share", Meta: `{"title":"创建分享"}`},
			{Name: "FileManage", AuthCode: "file:manage", Meta: `{"title":"文件管理（移动、重命名）"}`},
			{Name: "FileRecycleDelete", AuthCode: "file:recycle:delete", Meta: `{"title":"永久删除（回收站）"}`},
		}},
		{&fileMenus[1], []model.Menu{
			{Name: "ShareViewAll", AuthCode: "share:view:all", Meta: `{"title":"查看所有分享"}`},
			{Name: "ShareDelete", AuthCode: "share:delete", Meta: `{"title":"删除分享"}`},
			{Name: "ShareManage", AuthCode: "share:manage", Meta: `{"title":"管理分享（续期、过期等）"}`},
		}},
	}
	for _, bg := range fileButtonGroups {
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
