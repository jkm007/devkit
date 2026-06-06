package oauth

import (
	"fmt"
	"sync"
)

// Provider OAuth 登录提供商接口
type Provider interface {
	// Name 返回提供商名称（如 "github", "google", "wechat"）
	Name() string
	// AuthURL 生成授权 URL
	AuthURL(state string) string
	// ExchangeToken 用授权码换取 Token
	ExchangeToken(code string) (*Token, error)
	// GetUserInfo 用 Token 获取用户信息
	GetUserInfo(token *Token) (*UserInfo, error)
}

// Token OAuth Token
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

// UserInfo 第三方用户信息
type UserInfo struct {
	ProviderUserID string `json:"provider_user_id"`
	Username       string `json:"username"`
	Email          string `json:"email,omitempty"`
	Avatar         string `json:"avatar,omitempty"`
}

// ProviderConfig OAuth 提供商配置
type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// 全局提供商注册表
var (
	providers = make(map[string]Provider)
	mu        sync.RWMutex
)

// Register 注册 OAuth 提供商
func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers[p.Name()] = p
}

// Get 获取已注册的 OAuth 提供商
func Get(name string) (Provider, error) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", name)
	}
	return p, nil
}

// List 列出所有已注册的提供商名称
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	return names
}
