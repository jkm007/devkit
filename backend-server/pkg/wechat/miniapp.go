package wechat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// MiniApp 微信小程序登录
type MiniApp struct {
	appID     string
	appSecret string
}

// NewMiniApp 创建小程序登录 Provider
func NewMiniApp(appID, appSecret string) *MiniApp {
	return &MiniApp{
		appID:     appID,
		appSecret: appSecret,
	}
}

// miniAppSessionResponse 小程序 code2Session 响应
type miniAppSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// Login 用 wx.login() 获取的 code 换取 openid 和 session_key
func (m *MiniApp) Login(code string) (*LoginResult, error) {
	params := url.Values{}
	params.Set("appid", m.appID)
	params.Set("secret", m.appSecret)
	params.Set("js_code", code)
	params.Set("grant_type", "authorization_code")

	reqURL := "https://api.weixin.qq.com/sns/jscode2session?" + params.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("小程序登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result miniAppSessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析小程序登录响应失败: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &LoginResult{
		OpenID:     result.OpenID,
		UnionID:    result.UnionID,
		SessionKey: result.SessionKey,
	}, nil
}
