package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"backend-server/pkg/database"
	"backend-server/pkg/jwt"
	"backend-server/pkg/logger"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		// 优先从 Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}

		// Header 没有 Token 时，从 Cookie 获取（支持浏览器图片请求等场景）
		if tokenStr == "" {
			if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
				tokenStr = cookie
			}
		}

		// Cookie 也没有时，从查询参数获取（支持 <img>/<video> 标签加载认证资源）
		if tokenStr == "" {
			if queryToken := c.Query("token"); queryToken != "" {
				tokenStr = queryToken
			}
		}

		if tokenStr == "" {
			response.Unauthorized(c, "缺少认证信息")
			c.Abort()
			return
		}

		// 解析 Token（已包含算法校验）
		claims, err := jwt.Parse(tokenStr)
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
		// 采用 fail-closed 策略：Redis 不可用时拒绝请求，防止已注销的 Token 继续使用
		// 使用 SHA-256 哈希存储，减少内存占用并降低 Token 泄露风险
		blacklistKey := tokenBlacklistKey(tokenStr)
		val, err := database.GetRedis().Get(context.Background(), blacklistKey).Result()
		if err == nil && val != "" {
			response.Unauthorized(c, "Token 已失效，请重新登录")
			c.Abort()
			return
		}
		// redis.Nil 表示 key 不存在（正常情况），其他错误表示 Redis 故障
		if err != nil && err != redis.Nil {
			logger.Error("Redis 黑名单检查失败(fail-closed)", zap.Error(err))
			response.InternalError(c, "服务暂时不可用，请稍后重试")
			c.Abort()
			return
		}

		// 检查设备是否被踢出（同样采用 fail-closed 策略）
		kickedKey := fmt.Sprintf("kicked_device:%d:%s", claims.UserID, deviceID)
		val, err = database.GetRedis().Get(context.Background(), kickedKey).Result()
		if err == nil && val != "" {
			response.Unauthorized(c, "该设备已被踢出，请重新登录")
			c.Abort()
			return
		}
		if err != nil && err != redis.Nil {
			logger.Error("Redis 设备踢出检查失败(fail-closed)", zap.Error(err))
			response.InternalError(c, "服务暂时不可用，请稍后重试")
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

// GeneratePreviewToken 生成预签名 Token
func GeneratePreviewToken(userID uint, fileID uint) (string, error) {
	return jwt.GeneratePreviewToken(userID, fileID)
}

// ValidatePreviewToken 验证预签名 Token
func ValidatePreviewToken(tokenStr string) (uint, uint, error) {
	return jwt.ValidatePreviewToken(tokenStr)
}

// GetCurrentDeviceID 从上下文获取当前设备ID
func GetCurrentDeviceID(c *gin.Context) string {
	if did, exists := c.Get("device_id"); exists {
		return did.(string)
	}
	return ""
}

// tokenBlacklistKey 生成 Token 黑名单的 Redis key
// 使用 SHA-256 哈希替代明文存储，减少内存占用并降低 Token 泄露风险
func tokenBlacklistKey(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("token_blacklist:%s", hex.EncodeToString(h[:]))
}
