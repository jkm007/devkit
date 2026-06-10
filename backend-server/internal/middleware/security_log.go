package middleware

import (
	"bytes"
	"io"
	"strings"

	"backend-server/internal/service"

	"github.com/gin-gonic/gin"
)

// responseWriter 包装 gin.ResponseWriter 以捕获状态码
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// SecurityLogMiddleware 自动记录写操作的安全日志中间件
func SecurityLogMiddleware() gin.HandlerFunc {
	securityLogSvc := service.NewSecurityLogService()

	return func(c *gin.Context) {
		method := c.Request.Method

		// 只记录写操作
		if method != "POST" && method != "PUT" && method != "DELETE" {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		// 跳过不需要记录的路径
		if shouldSkipPath(path) {
			c.Next()
			return
		}

		// 检查是否有匹配的事件类型
		eventType, detail := service.PathToEventType(method, path)
		if eventType == "" {
			c.Next()
			return
		}

		// 包装 response writer 捕获状态码
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = w

		// 读取请求体（用于记录详情）
		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// 执行请求
		c.Next()

		// 获取当前用户 ID
		userID := GetCurrentUserID(c)
		if userID == 0 {
			return
		}

		// 判断是否成功（2xx 状态码）
		status := w.Status()
		logStatus := 0
		if status >= 200 && status < 300 {
			logStatus = 1
		}

		// 构建详细信息
		fullDetail := detail
		if logStatus == 0 {
			fullDetail += "（失败）"
		}

		// 记录日志
		_ = securityLogSvc.Record(
			userID,
			eventType,
			fullDetail,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			logStatus,
		)
	}
}

// shouldSkipPath 判断是否跳过记录
func shouldSkipPath(path string) bool {
	// 认证相关路径（登录/注册/验证码等）不做安全日志
	if strings.HasPrefix(path, "/api/v1/auth/") {
		return true
	}
	// 文件上传不做日志
	if strings.HasPrefix(path, "/api/v1/files/upload") {
		return true
	}
	return false
}
