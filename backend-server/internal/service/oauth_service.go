package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"backend-server/pkg/jwt"
	"backend-server/pkg/logger"
	"backend-server/pkg/oauth"

	"gorm.io/gorm"
)

// OAuthService 第三方登录绑定服务
type OAuthService struct {
	repo      *repository.OAuthUserRepo
	userRepo  *repository.UserRepo
	roleRepo  *repository.RoleRepo
	groupRepo *repository.GroupRepo
}

// NewOAuthService 创建第三方登录绑定服务
func NewOAuthService() *OAuthService {
	db := database.GetMySQL()
	return &OAuthService{
		repo:      repository.NewOAuthUserRepo(db),
		userRepo:  repository.NewUserRepo(db),
		roleRepo:  repository.NewRoleRepo(db),
		groupRepo: repository.NewGroupRepo(db),
	}
}

// OAuthBindURLRequest 获取授权URL请求
type OAuthBindURLRequest struct {
	Provider    string `form:"provider" binding:"required"`
	RedirectURI string `form:"redirectUri"`
}

// OAuthUnbindRequest 解绑请求
type OAuthUnbindRequest struct {
	Provider string `json:"provider" binding:"required"`
}

// OAuthCallbackRequest 回调请求
type OAuthCallbackRequest struct {
	Provider string `form:"provider" binding:"required"`
	Code     string `form:"code" binding:"required"`
	State    string `form:"state" binding:"required"`
}

// OAuthBindingItem 绑定列表项
type OAuthBindingItem struct {
	ID               uint      `json:"id"`
	Provider         string    `json:"provider"`
	ProviderUsername string    `json:"providerUsername"`
	ProviderAvatar   string    `json:"providerAvatar"`
	CreatedAt        time.Time `json:"createdAt"`
}

// ListBindings 获取用户的第三方绑定列表
func (s *OAuthService) ListBindings(userID uint) ([]OAuthBindingItem, error) {
	bindings, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	var items []OAuthBindingItem
	for _, b := range bindings {
		items = append(items, OAuthBindingItem{
			ID:               b.ID,
			Provider:         b.Provider,
			ProviderUsername: b.ProviderUsername,
			ProviderAvatar:   b.ProviderAvatar,
			CreatedAt:        b.CreatedAt,
		})
	}
	return items, nil
}

// Unbind 解绑第三方账号
func (s *OAuthService) Unbind(userID uint, providerName string) error {
	binding, err := s.repo.GetByUserAndProvider(userID, providerName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("This provider is not bound.")
		}
		return err
	}

	// 如果用户没有设置密码，则至少保留一个 OAuth 绑定
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user.Password == "" {
		count, err := s.repo.CountByUser(userID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return errors.New("Cannot unbind the only login method.")
		}
	}

	return s.repo.Delete(binding.ID)
}

// GetAuthorizeURL 获取第三方授权 URL
func (s *OAuthService) GetAuthorizeURL(providerName string) (string, error) {
	p, err := oauth.Get(providerName)
	if err != nil {
		return "", err
	}

	// 生成随机 state 防 CSRF
	state, err := generateOAuthState()
	if err != nil {
		return "", fmt.Errorf("生成 state 失败: %w", err)
	}

	// 存入 Redis，5 分钟过期
	rdb := database.GetRedis()
	rdb.Set(context.Background(), "oauth_state:"+state, providerName, 5*time.Minute)

	return p.AuthURL(state), nil
}

// HandleCallback 处理 OAuth 回调
func (s *OAuthService) HandleCallback(providerName, code, state, clientIP, userAgent string) (*LoginResponse, error) {
	// 1. 验证 state
	rdb := database.GetRedis()
	storedProvider, err := rdb.Get(context.Background(), "oauth_state:"+state).Result()
	if err != nil {
		return nil, errors.New("无效或过期的 state 参数")
	}
	if storedProvider != providerName {
		return nil, errors.New("state 参数不匹配")
	}
	rdb.Del(context.Background(), "oauth_state:"+state)

	// 2. 获取 provider
	p, err := oauth.Get(providerName)
	if err != nil {
		return nil, err
	}

	// 3. 用 code 换 token + 获取用户信息
	var userInfo *oauth.UserInfo
	var token *oauth.Token

	if providerName == "wechat" {
		// 微信特殊处理
		wp := p.(*oauth.WeChatProvider)
		wt, err := wp.ExchangeTokenWithOpenID(code)
		if err != nil {
			return nil, fmt.Errorf("微信换取 token 失败: %w", err)
		}
		token = &oauth.Token{
			AccessToken:  wt.AccessToken,
			RefreshToken: wt.RefreshToken,
			ExpiresIn:    wt.ExpiresIn,
		}
		userInfo, err = wp.GetUserInfoWithOpenID(wt.AccessToken, wt.OpenID)
		if err != nil {
			return nil, fmt.Errorf("获取微信用户信息失败: %w", err)
		}
	} else {
		token, err = p.ExchangeToken(code)
		if err != nil {
			return nil, fmt.Errorf("换取 token 失败: %w", err)
		}
		userInfo, err = p.GetUserInfo(token)
		if err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %w", err)
		}
	}

	// 4. 查找是否已绑定
	existingBinding, _ := s.repo.GetByProvider(providerName, userInfo.ProviderUserID)
	if existingBinding != nil && existingBinding.ID > 0 {
		// 已绑定，直接登录
		user, err := s.userRepo.GetByID(existingBinding.UserID)
		if err != nil || user == nil {
			return nil, errors.New("绑定的用户不存在")
		}
		if user.Status != 1 {
			return nil, errors.New("账号已被禁用")
		}

		// 更新 token
		existingBinding.AccessToken = token.AccessToken
		existingBinding.RefreshToken = token.RefreshToken
		s.repo.Update(existingBinding)

		return s.oauthBuildLoginResponse(user, clientIP)
	}

	// 5. 未绑定，自动注册新用户
	// 生成规范化用户名：provider_随机8位字符
	randomSuffix, _ := generateRandomSuffix(8)
	username := providerName + "_" + randomSuffix
	newUser := &model.User{
		Name:           username,
		Nickname:       userInfo.Username,
		Avatar:         userInfo.Avatar,
		Email:          userInfo.Email,
		Status:         1,
		RegisterSource: providerName,
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
		Provider:         providerName,
		ProviderType:     providerName,
		ProviderUserID:   userInfo.ProviderUserID,
		ProviderUsername: userInfo.Username,
		ProviderAvatar:   userInfo.Avatar,
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
	}
	if token.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
		oauthUser.ExpiresAt = &exp
	}
	if err := s.repo.Create(oauthUser); err != nil {
		logger.Error("创建 OAuth 绑定失败")
	}

	logger.Info(fmt.Sprintf("OAuth 自动注册: %s (%s)", newUser.Name, providerName))

	return s.oauthBuildLoginResponse(newUser, clientIP)
}

// oauthBuildLoginResponse 为 OAuth 登录构建登录响应
func (s *OAuthService) oauthBuildLoginResponse(user *model.User, clientIP string) (*LoginResponse, error) {
	roleNames, err := s.oauthCollectRoleNames(user)
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
			return nil, errors.New("User has no assigned roles")
		}
	}

	tokenPair, err := jwt.Generate(user.ID, user.Name, roleNames)
	if err != nil {
		return nil, err
	}

	rdb := database.GetRedis()
	rdb.Set(context.Background(), fmt.Sprintf("refresh_token:%d", user.ID), oauthHashToken(tokenPair.RefreshToken), 30*24*time.Hour)

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

// oauthCollectRoleNames 收集用户角色名称
func (s *OAuthService) oauthCollectRoleNames(user *model.User) ([]string, error) {
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

// generateOAuthState 生成随机 state
func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateRandomSuffix 生成随机字符串后缀
func generateRandomSuffix(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// 使用 hex 编码并截取所需长度
	return hex.EncodeToString(b)[:length], nil
}

// oauthHashToken 对 token 做 SHA256 哈希
func oauthHashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}