package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 安全响应头中间件
// 设置 HTTP 安全头以防止常见 Web 攻击
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止 MIME 类型嗅探攻击
		// 浏览器将严格遵循 Content-Type 头，不会尝试猜测内容类型
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止点击劫持攻击
		// 禁止页面在 iframe 中加载（DENY）或仅允许同源加载（SAMEORIGIN）
		c.Header("X-Frame-Options", "DENY")

		// 控制 Referrer 信息泄露
		// strict-origin-when-cross-origin: 仅在同源请求时发送完整 URL，跨域仅发送源
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 启用 XSS 过滤器（现代浏览器已内置，但仍作为兼容性保障）
		c.Header("X-XSS-Protection", "1; mode=block")

		// 内容安全策略
		// 默认只允许加载同源资源，脚本只允许同源执行
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:;")

		// 严格传输安全
		// 强制浏览器使用 HTTPS 访问（仅在生产环境启用）
		// 注意：开发环境通常使用 HTTP，所以需要根据实际情况调整
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// 权限策略
		// 限制浏览器功能访问
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		c.Next()
	}
}
