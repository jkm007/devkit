package middleware

import (
	"backend-server/internal/service"
	"backend-server/pkg/response"
	"sync"

	"github.com/gin-gonic/gin"
)

// 全局单例 AuthService（懒加载模式）
var (
	authServiceInstance *service.AuthService
	once                sync.Once
)

// getAuthService 获取 AuthService 单例（懒加载）
func getAuthService() *service.AuthService {
	once.Do(func() {
		authServiceInstance = service.NewAuthService()
	})
	return authServiceInstance
}

// Permission 权限码校验中间件
// 检查当前用户是否拥有所需的权限码
func Permission(requiredCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetCurrentUserID(c)
		if userID == 0 {
			response.Forbidden(c, "无法获取用户信息")
			c.Abort()
			return
		}

		// 通过 Service 层获取用户权限码（自带 Redis 缓存）
		codes, err := getAuthService().GetPermissionCodes(userID)
		if err != nil {
			response.Forbidden(c, "获取权限信息失败")
			c.Abort()
			return
		}

		// 检查是否拥有所需权限码
		for _, code := range codes {
			if code == requiredCode {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "权限不足，需要权限: "+requiredCode)
		c.Abort()
	}
}
