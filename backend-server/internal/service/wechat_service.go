package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"backend-server/pkg/jwt"
	"backend-server/pkg/logger"
	"backend-server/pkg/wechat"
)

// WeChatService 微信登录服务
type WeChatService struct {
	userRepo      *repository.UserRepo
	oauthRepo     *repository.OAuthUserRepo
	roleRepo      *repository.RoleRepo
	settingRepo   *repository.SystemSettingRepo
	securityLogSvc *SecurityLogService
}

// NewWeChatService 创建微信登录服务
func NewWeChatService() *WeChatService {
	db := database.GetMySQL()
	return &WeChatService{
		userRepo:       repository.NewUserRepo(db),
		oauthRepo:      repository.NewOAuthUserRepo(db),
		roleRepo:       repository.NewRoleRepo(db),
		settingRepo:    repository.NewSystemSettingRepo(db),
		securityLogSvc: NewSecurityLogService(),
	}
}

// RecordSecurityLog 记录安全日志
func (s *WeChatService) RecordSecurityLog(userID uint, eventType, detail, ip, userAgent string, status int) {
	if s.securityLogSvc != nil {
		_ = s.securityLogSvc.Record(userID, eventType, detail, ip, userAgent, status)
	}
}

// wechatConfig 微信配置（从数据库加载）
type wechatConfig struct {
	// 网站扫码
	OAuthEnabled    bool
	OAuthAppID      string
	OAuthSecret     string
	OAuthRedirectURL string
	// 小程序
	MiniAppEnabled bool
	MiniAppAppID   string
	MiniAppSecret  string
	// 公众号
	OfficialEnabled    bool
	OfficialAppID      string
	OfficialSecret     string
	OfficialRedirectURL string
}

// loadConfig 从数据库加载微信配置
func (s *WeChatService) loadConfig() (*wechatConfig, error) {
	settings, err := s.settingRepo.GetByGroup("wechat")
	if err != nil {
		return nil, fmt.Errorf("加载微信配置失败: %w", err)
	}

	cfg := &wechatConfig{}
	for _, setting := range settings {
		val := setting.Value
		// 去除 JSON 引号
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}

		switch setting.Key {
		case "wechat_oauth_enabled":
			cfg.OAuthEnabled = val == "true"
		case "wechat_oauth_appid":
			cfg.OAuthAppID = val
		case "wechat_oauth_secret":
			cfg.OAuthSecret = val
		case "wechat_oauth_redirect_url":
			cfg.OAuthRedirectURL = val
		case "wechat_miniapp_enabled":
			cfg.MiniAppEnabled = val == "true"
		case "wechat_miniapp_appid":
			cfg.MiniAppAppID = val
		case "wechat_miniapp_secret":
			cfg.MiniAppSecret = val
		case "wechat_official_enabled":
			cfg.OfficialEnabled = val == "true"
		case "wechat_official_appid":
			cfg.OfficialAppID = val
		case "wechat_official_secret":
			cfg.OfficialSecret = val
		case "wechat_official_redirect_url":
			cfg.OfficialRedirectURL = val
		}
	}

	return cfg, nil
}

// LoginByMiniProgram 小程序登录
func (s *WeChatService) LoginByMiniProgram(code, clientIP string) (*LoginResponse, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.MiniAppEnabled {
		return nil, errors.New("小程序登录未启用")
	}
	if cfg.MiniAppAppID == "" || cfg.MiniAppSecret == "" {
		return nil, errors.New("小程序配置不完整")
	}

	provider := wechat.NewMiniApp(cfg.MiniAppAppID, cfg.MiniAppSecret)
	result, err := provider.Login(code)
	if err != nil {
		return nil, fmt.Errorf("小程序登录失败: %w", err)
	}

	return s.findOrCreateUser(result.OpenID, result.UnionID, result.Nickname, result.Avatar, "wechat_mini", clientIP)
}

// GetOfficialAuthorizeURL 获取公众号授权 URL
func (s *WeChatService) GetOfficialAuthorizeURL(scope string) (string, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return "", err
	}
	if !cfg.OfficialEnabled {
		return "", errors.New("公众号登录未启用")
	}
	if cfg.OfficialAppID == "" || cfg.OfficialSecret == "" || cfg.OfficialRedirectURL == "" {
		return "", errors.New("公众号配置不完整")
	}

	provider := wechat.NewOfficialAccount(cfg.OfficialAppID, cfg.OfficialSecret, cfg.OfficialRedirectURL)
	state, err := generateWeChatState()
	if err != nil {
		return "", fmt.Errorf("生成 state 失败: %w", err)
	}

	rdb := database.GetRedis()
	rdb.Set(context.Background(), "wechat_state:"+state, "official", 5*time.Minute)

	return provider.GetAuthorizeURL(scope, state), nil
}

// LoginByOfficial 公众号 H5 登录
func (s *WeChatService) LoginByOfficial(code, state, clientIP string) (*LoginResponse, error) {
	// 验证 state
	rdb := database.GetRedis()
	storedType, err := rdb.Get(context.Background(), "wechat_state:"+state).Result()
	if err != nil {
		return nil, errors.New("无效或过期的 state 参数")
	}
	if storedType != "official" {
		return nil, errors.New("state 参数不匹配")
	}
	rdb.Del(context.Background(), "wechat_state:"+state)

	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.OfficialEnabled {
		return nil, errors.New("公众号登录未启用")
	}

	provider := wechat.NewOfficialAccount(cfg.OfficialAppID, cfg.OfficialSecret, cfg.OfficialRedirectURL)
	result, err := provider.Login(code)
	if err != nil {
		return nil, fmt.Errorf("公众号登录失败: %w", err)
	}

	return s.findOrCreateUser(result.OpenID, result.UnionID, result.Nickname, result.Avatar, "wechat_mp", clientIP)
}

// GetWebAuthorizeURL 获取网站扫码授权 URL
func (s *WeChatService) GetWebAuthorizeURL() (string, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return "", err
	}
	if !cfg.OAuthEnabled {
		return "", errors.New("微信扫码登录未启用")
	}
	if cfg.OAuthAppID == "" || cfg.OAuthSecret == "" || cfg.OAuthRedirectURL == "" {
		return "", errors.New("网站扫码配置不完整")
	}

	provider := wechat.NewWeb(cfg.OAuthAppID, cfg.OAuthSecret, cfg.OAuthRedirectURL)
	state, err := generateWeChatState()
	if err != nil {
		return "", fmt.Errorf("生成 state 失败: %w", err)
	}

	rdb := database.GetRedis()
	rdb.Set(context.Background(), "wechat_state:"+state, "web", 5*time.Minute)

	return provider.GetAuthorizeURL(state), nil
}

// LoginByWeb 网站扫码登录
func (s *WeChatService) LoginByWeb(code, state, clientIP string) (*LoginResponse, error) {
	// 验证 state
	rdb := database.GetRedis()
	storedType, err := rdb.Get(context.Background(), "wechat_state:"+state).Result()
	if err != nil {
		return nil, errors.New("无效或过期的 state 参数")
	}
	if storedType != "web" {
		return nil, errors.New("state 参数不匹配")
	}
	rdb.Del(context.Background(), "wechat_state:"+state)

	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.OAuthEnabled {
		return nil, errors.New("微信扫码登录未启用")
	}

	provider := wechat.NewWeb(cfg.OAuthAppID, cfg.OAuthSecret, cfg.OAuthRedirectURL)
	result, err := provider.Login(code)
	if err != nil {
		return nil, fmt.Errorf("微信扫码登录失败: %w", err)
	}

	return s.findOrCreateUser(result.OpenID, result.UnionID, result.Nickname, result.Avatar, "wechat", clientIP)
}

// findOrCreateUser 查找或创建用户（三种微信登录共用）
func (s *WeChatService) findOrCreateUser(openID, unionID, nickname, avatar, source, clientIP string) (*LoginResponse, error) {
	// 优先用 unionID 查找（跨应用统一身份），其次用 openid + provider
	providerUserID := unionID
	if providerUserID == "" {
		providerUserID = openID
	}

	// 查找已绑定的用户
	existingBinding, _ := s.oauthRepo.GetByProvider(source, providerUserID)
	if existingBinding != nil && existingBinding.ID > 0 {
		user, err := s.userRepo.GetByID(existingBinding.UserID)
		if err != nil || user == nil {
			return nil, errors.New("绑定的用户不存在")
		}
		if user.Status != 1 {
			return nil, errors.New("账号已被禁用")
		}
		return s.buildLoginResponse(user, clientIP)
	}

	// 自动注册新用户
	newUser := &model.User{
		Name:           source + "_" + openID,
		Nickname:       nickname,
		Avatar:         avatar,
		Status:         1,
		RegisterSource: source,
	}
	if newUser.Nickname == "" {
		newUser.Nickname = newUser.Name
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 分配默认角色
	defaultRole, err := s.roleRepo.GetByName("user")
	if err == nil && defaultRole != nil {
		s.userRepo.SyncUserRoles(newUser.ID, []uint{defaultRole.ID})
	}

	// 创建 OAuth 绑定
	oauthUser := &model.OAuthUser{
		UserID:           newUser.ID,
		Provider:         source,
		ProviderType:     source,
		ProviderUserID:   providerUserID,
		ProviderUsername: nickname,
		ProviderAvatar:   avatar,
	}
	if err := s.oauthRepo.Create(oauthUser); err != nil {
		logger.Error("创建微信绑定失败")
	}

	logger.Info(fmt.Sprintf("微信自动注册: %s (%s)", newUser.Name, source))

	return s.buildLoginResponse(newUser, clientIP)
}

// buildLoginResponse 构建登录响应
func (s *WeChatService) buildLoginResponse(user *model.User, clientIP string) (*LoginResponse, error) {
	// 收集角色名称
	roleNames, err := s.collectRoleNames(user)
	if err != nil {
		return nil, err
	}
	if len(roleNames) == 0 {
		defaultRole, err := s.roleRepo.GetByName("user")
		if err == nil && defaultRole != nil {
			s.userRepo.SyncUserRoles(user.ID, []uint{defaultRole.ID})
			roleNames = []string{"user"}
		}
		if len(roleNames) == 0 {
			return nil, errors.New("用户未分配角色")
		}
	}

	// 生成 Token
	tokenPair, err := jwt.Generate(user.ID, user.Name, roleNames)
	if err != nil {
		return nil, err
	}

	// 存储 RefreshToken 哈希到 Redis
	rdb := database.GetRedis()
	rdb.Set(context.Background(), fmt.Sprintf("refresh_token:%d", user.ID), oauthHashToken(tokenPair.RefreshToken), 30*24*time.Hour)

	// 更新用户最后登录信息
	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = clientIP
	s.userRepo.Update(user)

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

// collectRoleNames 收集用户角色名称
func (s *WeChatService) collectRoleNames(user *model.User) ([]string, error) {
	roleIDSet := make(map[uint]bool)

	directRoles, err := s.userRepo.GetUserRoles(user.ID)
	if err != nil {
		return nil, err
	}
	for _, role := range directRoles {
		roleIDSet[role.ID] = true
	}

	if user.GroupID > 0 {
		groupRoleIDs, err := repository.NewGroupRepo(database.GetMySQL()).GetGroupRoleIDsRecursive(user.GroupID)
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

// generateWeChatState 生成随机 state
func generateWeChatState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
