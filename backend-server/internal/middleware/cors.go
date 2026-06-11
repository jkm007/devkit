package middleware

import (
	"log"
	"strings"
	"time"

	"backend-server/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	// 安全校验：AllowCredentials 为 true 时，AllowOrigins 禁止使用通配符 "*"
	// 这是 CORS 规范要求：https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	if cfg.AllowCredentials {
		for _, origin := range cfg.AllowOrigins {
			if origin == "*" {
				log.Fatal("[CORS] 安全错误：AllowCredentials=true 时，AllowOrigins 不允许包含通配符 \"*\"，请在配置中指定明确的域名")
			}
		}
	}

	// 防御性检查：过滤空字符串，避免意外行为
	origins := make([]string, 0, len(cfg.AllowOrigins))
	for _, o := range cfg.AllowOrigins {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Second,
	})
}
