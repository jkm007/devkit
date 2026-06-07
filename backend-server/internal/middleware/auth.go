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

		// 解析 Token（已包含算法校验）
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

		// 检查 Token 是否在黑名单中（logout 后失效）
		blacklistKey := fmt.Sprintf("token_blacklist:%s", parts[1])
		val, err := database.GetRedis().Get(context.Background(), blacklistKey).Result()
		if err == nil && val != "" {
			response.Unauthorized(c, "Token 已失效，请重新登录")
			c.Abort()
			return
		}

		// 检查设备是否被踢出
		kickedKey := fmt.Sprintf("kicked_device:%d:%s", claims.UserID, deviceID)
		val, err = database.GetRedis().Get(context.Background(), kickedKey).Result()
		if err == nil && val != "" {
			response.Unauthorized(c, "该设备已被踢出，请重新登录")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("device_id", deviceID)

		c.Next()
	}
}

// GetCurrentUserID 从上下文获取当前用户 ID
func GetCurrentUserID(c *gin.Context) uint {
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(uint); ok {
			return id
		}
	}
	return 0
}

// GetCurrentRoles 从上下文获取当前用户角色列表
func GetCurrentRoles(c *gin.Context) []string {
	if roles, exists := c.Get("roles"); exists {
		if r, ok := roles.([]string); ok {
			return r
		}
	}
	return nil
}

// GetCurrentRole 从上下文获取当前用户第一个角色（向后兼容）
func GetCurrentRole(c *gin.Context) string {
	roles := GetCurrentRoles(c)
	if len(roles) > 0 {
		return roles[0]
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
