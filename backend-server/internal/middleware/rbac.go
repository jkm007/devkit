package middleware

import (
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// RBAC 基于角色的访问控制中间件
func RBAC(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := GetCurrentRoles(c)
		if len(roles) == 0 {
			response.Forbidden(c, "无法获取用户角色")
			c.Abort()
			return
		}

		// 构建允许角色的集合
		allowedSet := make(map[string]bool, len(allowedRoles))
		for _, r := range allowedRoles {
			allowedSet[r] = true
		}

		// 检查用户是否有任意一个角色在允许列表中
		for _, role := range roles {
			if allowedSet[role] {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "权限不足")
		c.Abort()
	}
}
