package wechat

import (
	"fmt"
	"net/url"
)

// OfficialAccount 微信公众号 H5 登录
type OfficialAccount struct {
	appID       string
	appSecret   string
	redirectURL string
}

// NewOfficialAccount 创建公众号 H5 登录 Provider
func NewOfficialAccount(appID, appSecret, redirectURL string) *OfficialAccount {
	return &OfficialAccount{
		appID:       appID,
		appSecret:   appSecret,
		redirectURL: redirectURL,
	}
}

// GetAuthorizeURL 获取公众号 OAuth 授权 URL
// scope: snsapi_base（静默授权，仅获取 openid）或 snsapi_userinfo（弹窗授权，获取用户信息）
func (o *OfficialAccount) GetAuthorizeURL(scope, state string) string {
	if scope == "" {
		scope = "snsapi_userinfo"
	}
	params := url.Values{}
	params.Set("appid", o.appID)
	params.Set("redirect_uri", o.redirectURL)
	params.Set("response_type", "code")
	params.Set("scope", scope)
	params.Set("state", state)
	return "https://open.weixin.qq.com/connect/oauth2/authorize?" + params.Encode() + "#wechat_redirect"
}

// officialTokenResponse 公众号 token 响应
type officialTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	UnionID      string `json:"unionid"`
	ExpiresIn    int    `json:"expires_in"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// Login 用授权码换取用户信息
func (o *OfficialAccount) Login(code string) (*LoginResult, error) {
	// 1. 用 code 换取 access_token + openid
	tokenResult, err := o.exchangeToken(code)
	if err != nil {
		return nil, err
	}

	// 2. 用 access_token + openid 获取用户信息
	userInfo, err := o.getUserInfo(tokenResult.AccessToken, tokenResult.OpenID)
	if err != nil {
		// 静默授权模式下（snsapi_base）无法获取用户信息，只返回 openid
		// ErrCode 48001 = api unauthorized (snsapi_base 模式无权调用 userinfo)
		if tokenResult.ErrCode == 0 {
			return &LoginResult{
				OpenID:      tokenResult.OpenID,
				UnionID:     tokenResult.UnionID,
				AccessToken: tokenResult.AccessToken,
			}, nil
		}
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

// exchangeToken 用 code 换取 access_token
func (o *OfficialAccount) exchangeToken(code string) (*officialTokenResponse, error) {
	params := url.Values{}
	params.Set("appid", o.appID)
	params.Set("secret", o.appSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")

	var result officialTokenResponse
	if err := doWeChatRequest("https://api.weixin.qq.com/sns/oauth2/access_token?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// officialUserInfo 公众号用户信息
type officialUserInfo struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"headimgurl"`
	OpenID   string `json:"openid"`
	UnionID  string `json:"unionid"`
	ErrCode  int    `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
}

// getUserInfo 获取公众号用户信息
func (o *OfficialAccount) getUserInfo(accessToken, openID string) (*officialUserInfo, error) {
	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("openid", openID)
	params.Set("lang", "zh_CN")

	var result officialUserInfo
	if err := doWeChatRequest("https://api.weixin.qq.com/sns/userinfo?"+params.Encode(), &result); err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}
