package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// WeChatProvider 微信 OAuth 提供商
type WeChatProvider struct {
	appID     string
	appSecret string
	redirectURL string
}

// NewWeChatProvider 创建微信 OAuth 提供商
func NewWeChatProvider(cfg ProviderConfig) *WeChatProvider {
	return &WeChatProvider{
		appID:       cfg.ClientID,
		appSecret:   cfg.ClientSecret,
		redirectURL: cfg.RedirectURL,
	}
}

func (p *WeChatProvider) Name() string { return "wechat" }

func (p *WeChatProvider) AuthURL(state string) string {
	params := url.Values{}
	params.Set("appid", p.appID)
	params.Set("redirect_uri", p.redirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "snsapi_login")
	params.Set("state", state)
	return "https://open.weixin.qq.com/connect/qrconnect?" + params.Encode() + "#wechat_redirect"
}

func (p *WeChatProvider) ExchangeToken(code string) (*Token, error) {
	params := url.Values{}
	params.Set("appid", p.appID)
	params.Set("secret", p.appSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")

	reqURL := "https://api.weixin.qq.com/sns/oauth2/access_token?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("WeChat token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		OpenID       string `json:"openid"`
		UnionID      string `json:"unionid"`
		ExpiresIn    int    `json:"expires_in"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse WeChat token response failed: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("WeChat OAuth error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &Token{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}, nil
}

func (p *WeChatProvider) GetUserInfo(token *Token) (*UserInfo, error) {
	// 微信需要先获取 openid（从 ExchangeToken 中获取）
	// 这里通过 access_token + openid 获取用户信息
	// 注意：微信的 access_token 响应中包含 openid，需要在 ExchangeToken 中传递
	// 简化处理：通过 refresh_token 机制或直接调用 userinfo 接口

	// 微信获取用户信息需要 openid，但 Token 结构中没有存储
	// 实际使用时需要在 ExchangeToken 中保存 openid
	// 这里使用 unionid 作为唯一标识
	return nil, fmt.Errorf("WeChat GetUserInfo requires openid from token exchange, use GetUserInfoWithOpenID instead")
}

// GetUserInfoWithOpenID 使用 access_token 和 openid 获取微信用户信息
func (p *WeChatProvider) GetUserInfoWithOpenID(accessToken, openID string) (*UserInfo, error) {
	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("openid", openID)
	params.Set("lang", "zh_CN")

	reqURL := "https://api.weixin.qq.com/sns/userinfo?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("WeChat user info request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var user struct {
		OpenID     string `json:"openid"`
		UnionID    string `json:"unionid"`
		Nickname   string `json:"nickname"`
		HeadImgURL string `json:"headimgurl"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("parse WeChat user info failed: %w", err)
	}
	if user.ErrCode != 0 {
		return nil, fmt.Errorf("WeChat user info error: %d - %s", user.ErrCode, user.ErrMsg)
	}

	// 优先使用 unionid，没有则用 openid
	providerID := user.UnionID
	if providerID == "" {
		providerID = user.OpenID
	}

	return &UserInfo{
		ProviderUserID: providerID,
		Username:       user.Nickname,
		Avatar:         user.HeadImgURL,
	}, nil
}

// WeChatTokenResponse 微信 Token 响应（包含 openid，供外部使用）
type WeChatTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	UnionID      string `json:"unionid"`
	ExpiresIn    int    `json:"expires_in"`
}

// ExchangeTokenWithOpenID 用授权码换取 Token（返回包含 openid 的完整响应）
func (p *WeChatProvider) ExchangeTokenWithOpenID(code string) (*WeChatTokenResponse, error) {
	params := url.Values{}
	params.Set("appid", p.appID)
	params.Set("secret", p.appSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")

	reqURL := "https://api.weixin.qq.com/sns/oauth2/access_token?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("WeChat token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result WeChatTokenResponse
	var errResult struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse WeChat token response failed: %w", err)
	}
	if errResult.ErrCode != 0 {
		return nil, fmt.Errorf("WeChat OAuth error: %d - %s", errResult.ErrCode, errResult.ErrMsg)
	}

	return &result, nil
}
