package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	// csrfCookieName CSRF Token Cookie 名称
	csrfCookieName = "csrf_token"
	// csrfHeaderName CSRF Token 请求头名称
	csrfHeaderName = "X-CSRF-Token"
	// csrfTokenLength CSRF Token 字节长度（生成 hex 后为 64 字符）
	csrfTokenLength = 32
)

// csrfSafeMethods 不需要 CSRF 保护的 HTTP 方法
var csrfSafeMethods = map[string]bool{
	"GET":     true,
	"HEAD":    true,
	"OPTIONS": true,
}

// generateCSRFToken 生成随机 CSRF Token
func generateCSRFToken() (string, error) {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CSRF 返回 CSRF 防护中间件（Double Submit Cookie 模式）
//
// 工作原理：
//  1. 首次请求时，服务端生成随机 Token 并通过 Set-Cookie 下发给浏览器
//  2. 浏览器后续请求自动携带 Cookie，前端 JS 从 Cookie 读取 Token 并放入请求头 X-CSRF-Token
//  3. 中间件校验 Cookie 和 Header 中的 Token 是否一致
//
// 安全性保证：
//   - 攻击者无法读取跨域 Cookie（受同源策略保护），因此无法构造合法的 X-CSRF-Token 头
//   - Cookie 设置了 SameSite=Lax，进一步限制跨站请求携带
//   - 仅对状态变更方法（POST/PUT/DELETE/PATCH）进行校验
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅对状态变更方法进行 CSRF 校验
		if csrfSafeMethods[c.Request.Method] {
			// 安全方法：确保 Cookie 存在即可，方便前端后续请求使用
			ensureCSRFCookie(c)
			c.Next()
			return
		}

		// 确保 CSRF Cookie 存在（首次请求时自动设置）
		ensureCSRFCookie(c)

		// 从 Cookie 获取 Token
		cookieToken, err := c.Cookie(csrfCookieName)
		if err != nil || cookieToken == "" {
			response.Forbidden(c, "CSRF Token 缺失，请刷新页面后重试")
			c.Abort()
			return
		}

		// 从请求头获取 Token
		headerToken := c.GetHeader(csrfHeaderName)
		if headerToken == "" {
			response.Forbidden(c, "CSRF Token 缺失，请刷新页面后重试")
			c.Abort()
			return
		}

		// 校验 Cookie 与 Header 中的 Token 是否一致（常量时间比较防止时序攻击）
		if !secureCompare(cookieToken, headerToken) {
			response.Forbidden(c, "CSRF Token 无效，请刷新页面后重试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ensureCSRFCookie 确保 CSRF Token Cookie 存在
// 如果 Cookie 不存在或为空，则生成新的 Token 并设置到 Cookie
func ensureCSRFCookie(c *gin.Context) {
	cookieToken, err := c.Cookie(csrfCookieName)
	if err == nil && cookieToken != "" {
		return
	}

	token, err := generateCSRFToken()
	if err != nil {
		// Token 生成失败不阻断请求，仅记录日志
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // 前端 JS 需要读取此 Cookie
		Secure:   false, // 开发环境为 HTTP，生产环境由反向代理处理
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 小时
	})
}

// secureCompare 常量时间字符串比较，防止时序攻击
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
