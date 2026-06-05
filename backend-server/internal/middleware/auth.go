package middleware

import (
	"context"
	"fmt"
	"strings"

	"backend-server/pkg/database"
	"backend-server/pkg/jwt"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少 Authorization 头")
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Authorization 格式错误")
			c.Abort()
			return
		}

		// 解析 Token
		claims, err := jwt.Parse(parts[1])
		if err != nil {
			response.Unauthorized(c, "Token 无效或已过期")
			c.Abort()
			return
		}

		// 从前端 header 获取设备ID
		deviceID := c.GetHeader("X-Device-ID")
		if deviceID == "" {
			// 兼容旧客户端：用 User-Agent + IP 生成
			data := []byte(c.GetHeader("User-Agent") + c.ClientIP())
			if len(data) > 16 {
				data = data[:16]
			}
			deviceID = fmt.Sprintf("web-%x", data)
		}

		// 检查设备是否被踢出
		blacklistKey := fmt.Sprintf("kicked_device:%d:%s", claims.UserID, deviceID)
		val, err := database.GetRedis().Get(context.Background(), blacklistKey).Result()
		if err == nil && val != "" {
			response.Unauthorized(c, "该设备已被踢出，请重新登录")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("device_id", deviceID)

		c.Next()
	}
}

// GetCurrentUserID 从上下文获取当前用户 ID
func GetCurrentUserID(c *gin.Context) uint {
	if uid, exists := c.Get("user_id"); exists {
		return uid.(uint)
	}
	return 0
}

// GetCurrentRole 从上下文获取当前用户角色
func GetCurrentRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		return role.(string)
	}
	return ""
}

// GetCurrentDeviceID 从上下文获取当前设备ID
func GetCurrentDeviceID(c *gin.Context) string {
	if did, exists := c.Get("device_id"); exists {
		return did.(string)
	}
	return ""
}
