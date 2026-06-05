package middleware

import (
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// RBAC 基于角色的访问控制中间件
func RBAC(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetCurrentRole(c)
		if role == "" {
			response.Forbidden(c, "无法获取用户角色")
			c.Abort()
			return
		}

		// 检查角色是否在允许列表中
		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "权限不足")
		c.Abort()
	}
}
