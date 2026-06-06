package wechat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Web 微信网站扫码登录
type Web struct {
	appID       string
	appSecret   string
	redirectURL string
}

// NewWeb 创建网站扫码登录 Provider
func NewWeb(appID, appSecret, redirectURL string) *Web {
	return &Web{
		appID:       appID,
		appSecret:   appSecret,
		redirectURL: redirectURL,
	}
}

// GetAuthorizeURL 获取微信扫码授权 URL
func (w *Web) GetAuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("appid", w.appID)
	params.Set("redirect_uri", w.redirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "snsapi_login")
	params.Set("state", state)
	return "https://open.weixin.qq.com/connect/qrconnect?" + params.Encode() + "#wechat_redirect"
}

// Login 用授权码换取用户信息
func (w *Web) Login(code string) (*LoginResult, error) {
	// 1. 用 code 换取 access_token + openid
	tokenResult, err := w.exchangeToken(code)
	if err != nil {
		return nil, err
	}

	// 2. 用 access_token + openid 获取用户信息
	userInfo, err := w.getUserInfo(tokenResult.AccessToken, tokenResult.OpenID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		OpenID:      tokenResult.OpenID,
		UnionID:     tokenResult.UnionID,
		Nickname:    userInfo.Nickname,
		Avatar:      userInfo.Avatar,
		AccessToken: tokenResult.AccessToken,
	}, nil
}

// wechatTokenResponse 微信 token 响应
type wechatTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	UnionID      string `json:"unionid"`
	ExpiresIn    int    `json:"expires_in"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// exchangeToken 用 code 换取 access_token
func (w *Web) exchangeToken(code string) (*wechatTokenResponse, error) {
	params := url.Values{}
	params.Set("appid", w.appID)
	params.Set("secret", w.appSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")

	reqURL := "https://api.weixin.qq.com/sns/oauth2/access_token?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("微信 token 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result wechatTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析微信 token 响应失败: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// wechatUserInfo 微信用户信息
type wechatUserInfo struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"headimgurl"`
	OpenID   string `json:"openid"`
	UnionID  string `json:"unionid"`
	ErrCode  int    `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
}

// getUserInfo 获取微信用户信息
func (w *Web) getUserInfo(accessToken, openID string) (*wechatUserInfo, error) {
	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("openid", openID)
	params.Set("lang", "zh_CN")

	reqURL := "https://api.weixin.qq.com/sns/userinfo?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("微信用户信息请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result wechatUserInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析微信用户信息失败: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}
