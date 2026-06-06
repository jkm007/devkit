package wechat

// LoginResult 微信登录统一返回结果
type LoginResult struct {
	OpenID      string // 用户唯一标识
	UnionID     string // 开放平台唯一标识（需绑定开放平台才有）
	Nickname    string // 用户昵称
	Avatar      string // 用户头像 URL
	AccessToken string // 网页/公众号 access_token（小程序为空）
	SessionKey  string // 小程序 session_key（网页/公众号为空）
	ErrCode     int    // 错误码（0=成功）
	ErrMsg      string // 错误信息
}

// WebProvider 网站扫码登录 Provider
type WebProvider interface {
	// GetAuthorizeURL 获取授权 URL（生成二维码页面）
	GetAuthorizeURL(state string) string
	// Login 用授权码换取用户信息
	Login(code string) (*LoginResult, error)
}

// MiniProgramProvider 小程序登录 Provider
type MiniProgramProvider interface {
	// Login 用 wx.login() 获取的 code 换取 openid 和 session_key
	Login(code string) (*LoginResult, error)
}

// OfficialAccountProvider 公众号 H5 登录 Provider
type OfficialAccountProvider interface {
	// GetAuthorizeURL 获取公众号 OAuth 授权 URL
	// scope: snsapi_base（静默）或 snsapi_userinfo（弹窗授权）
	GetAuthorizeURL(scope, state string) string
	// Login 用授权码换取用户信息
	Login(code string) (*LoginResult, error)
}
