package router

import (
	"backend-server/config"
	"backend-server/internal/handler"
	"backend-server/internal/middleware"
	"backend-server/internal/repository"
	"backend-server/internal/service"
	"backend-server/internal/ws"
	"backend-server/pkg/database"

	_ "backend-server/docs" // swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup 初始化路由
func Setup(cfg *config.Config, hub *ws.Hub) *gin.Engine {
	// 设置运行模式
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORS))
	r.Use(middleware.RateLimiter(cfg.RateLimit))
	r.Use(gin.Recovery())

	// 初始化处理器
	authHandler := handler.NewAuthHandler()
	userHandler := handler.NewUserHandler()
	roleHandler := handler.NewRoleHandler()
	menuHandler := handler.NewMenuHandler()
	groupHandler := handler.NewGroupHandler()
	wsHandler := handler.NewWSHandler(hub)
	securityLogHandler := handler.NewSecurityLogHandler()
	loginDeviceHandler := handler.NewLoginDeviceHandler()
	oauthHandler := handler.NewOAuthHandler()
	userPrivacyHandler := handler.NewUserPrivacyHandler()
	passwordHistoryHandler := handler.NewPasswordHistoryHandler()
	userRealNameHandler := handler.NewUserRealNameHandler()
	roleApplicationHandler := handler.NewRoleApplicationHandler()
	systemSettingHandler := handler.NewSystemSettingHandler()
	riskScoreHandler := handler.NewRiskScoreHandler()
	wechatHandler := handler.NewWeChatHandler()
	uploadHandler := handler.NewUploadHandler()
	fileHandler := handler.NewFileHandler()
	mediaHandler := handler.NewMediaHandler()
	shareHandler := handler.NewShareHandler()
	tagHandler := handler.NewTagHandler(service.NewTagService(repository.NewTagRepo(database.GetMySQL()), repository.NewFileTagRepo(database.GetMySQL())))
	routingHandler := handler.NewRoutingHandler(service.NewRoutingService(repository.NewTagRoutingRepo(database.GetMySQL())))

	// 健康检查
	// @Summary      健康检查
	// @Description  检查服务是否正常运行
	// @Tags         系统
	// @Produce      json
	// @Success      200  {object}  map[string]string "status: ok"
	// @Router       /health [get]
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger 文档（仅 debug 模式可访问）
	if cfg.Server.Mode == "debug" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 静态文件服务（上传文件公开访问，仅 local 存储模式）
	if cfg.Storage.Driver == "local" {
		r.Static(cfg.Storage.Local.URLPrefix, cfg.Storage.Local.Path)
	}

	// 公开接口（无需认证）
	r.GET("/system/settings/public", systemSettingHandler.GetPublic)

	// 分享访问（公开，独立限流：每秒 10 个，突发 20）
	sharePublic := r.Group("/share", middleware.NewIPRateLimiter(10, 20))
	{
		sharePublic.GET("/:code", shareHandler.GetShareInfo)
		sharePublic.GET("/:code/files", shareHandler.GetShareFolderFiles)
		sharePublic.GET("/:code/file", shareHandler.GetShareFile)
		sharePublic.GET("/:code/file/:fileId", shareHandler.GetShareFile)
	}

	// 认证接口（无需认证）
	captchaHandler := handler.NewCaptchaHandler()
	verifyCodeHandler := handler.NewVerifyCodeHandler()
	auth := r.Group("/auth")
	{
		// 验证码接口单独限流（每秒 5 个，突发 10）
		auth.GET("/captcha", middleware.NewIPRateLimiter(5, 10), captchaHandler.GetCaptcha)
		auth.POST("/login", authHandler.Login)
		// 邮箱登录（限流：每秒 1 个，突发 3）
		auth.POST("/login-by-email", middleware.NewIPRateLimiter(1, 3), authHandler.LoginByEmail)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.GET("/oauth/authorize", oauthHandler.GetBindURL)
		// OAuth 回调（限流：每秒 2 个，突发 5）
		auth.GET("/oauth/callback", middleware.NewIPRateLimiter(2, 5), oauthHandler.Callback)
		// 邮箱验证码（限流：每秒 1 个，突发 2）
		auth.POST("/send-code", middleware.NewIPRateLimiter(1, 2), verifyCodeHandler.SendCode)
		// 验证验证码（限流：每秒 5 个，突发 10）
		auth.POST("/verify-code", middleware.NewIPRateLimiter(5, 10), verifyCodeHandler.VerifyCode)
		// 短信验证码（限流：每秒 1 个，突发 2）
		auth.POST("/send-sms-code", middleware.NewIPRateLimiter(1, 2), verifyCodeHandler.SendSMSCode)
		// 手机号登录（限流：每秒 1 个，突发 3）
		auth.POST("/login-by-phone", middleware.NewIPRateLimiter(1, 3), authHandler.LoginByPhone)
		// 注册（限流：每秒 1 个，突发 3）
		auth.POST("/register", middleware.NewIPRateLimiter(1, 3), authHandler.Register)
		// 重置密码（限流：每秒 1 个，突发 3）
		auth.POST("/reset-password", middleware.NewIPRateLimiter(1, 3), authHandler.ResetPassword)


		// 微信登录
		wechat := auth.Group("/wechat")
		{
			wechat.POST("/miniapp-login", wechatHandler.LoginByMiniProgram)
			wechat.GET("/official-authorize", wechatHandler.GetOfficialAuthorizeURL)
			wechat.GET("/official-callback", wechatHandler.LoginByOfficial)
			wechat.GET("/web-authorize", wechatHandler.GetWebAuthorizeURL)
			wechat.GET("/web-callback", wechatHandler.LoginByWeb)
		}
	}

	// 需要认证的接口
	authorized := r.Group("")
	authorized.Use(middleware.JWTAuth())
	authorized.Use(middleware.SecurityLogMiddleware())
	authorized.Use(middleware.CaptchaGuard())
	{
		// 用户信息
		authorized.GET("/user/info", authHandler.GetUserInfo)
		authorized.PUT("/user/info", authHandler.UpdateProfile)
		authorized.PUT("/auth/change-password", authHandler.ChangePassword)
		authorized.GET("/auth/codes", authHandler.GetPermissionCodes)
		authorized.GET("/auth/permission-version", authHandler.GetPermissionVersion)

		// WebSocket 连接
		authorized.GET("/ws", wsHandler.Handle)

		// 菜单（用户菜单）
		authorized.GET("/menu/all", menuHandler.GetAll)

		// 安全日志（当前用户）
		authorized.GET("/auth/security-logs", securityLogHandler.GetMyLogs)

		// 登录设备管理
		authorized.GET("/auth/devices", loginDeviceHandler.List)
		authorized.DELETE("/auth/devices/kick-all", loginDeviceHandler.KickAllOther)
		authorized.DELETE("/auth/devices/:id", loginDeviceHandler.Kick)

		// OAuth 绑定管理
		authorized.GET("/auth/oauth/bindings", oauthHandler.GetBindings)
		authorized.GET("/auth/oauth/bind-url", oauthHandler.GetBindURL)
		authorized.POST("/auth/oauth/unbind", oauthHandler.Unbind)

		// 密码历史
		authorized.POST("/auth/password-history/check", passwordHistoryHandler.Check)

		// 角色申请（用户端）
		authorized.POST("/auth/role-applications", roleApplicationHandler.Create)
		authorized.GET("/auth/role-applications", roleApplicationHandler.GetMyList)

		// 用户资料
		authorized.GET("/user/privacy", userPrivacyHandler.Get)
		authorized.PUT("/user/privacy", userPrivacyHandler.Update)
		authorized.GET("/user/real-name", userRealNameHandler.GetStatus)
		authorized.POST("/user/real-name", userRealNameHandler.Submit)

		// 文件上传（分片上传）
		authorized.POST("/files/upload/check", uploadHandler.CheckUpload)
		authorized.POST("/files/upload/init", uploadHandler.InitUpload)
		authorized.POST("/files/upload/part", uploadHandler.UploadPart)
		authorized.POST("/files/upload/complete", uploadHandler.CompleteUpload)
		authorized.POST("/files/upload/abort", uploadHandler.AbortUpload)
		authorized.GET("/files/upload/status", uploadHandler.GetUploadStatus)
		authorized.GET("/files/upload/tasks", uploadHandler.GetUserUploadTasks)
		authorized.GET("/files/upload/tasks/:id", uploadHandler.GetUploadTaskByID)

		// 文件管理
		authorized.POST("/files/folder", fileHandler.CreateFolder)
		authorized.GET("/files/tree", fileHandler.GetFolderTree)
		authorized.PUT("/files/folder/:id", fileHandler.RenameFolder)
		authorized.DELETE("/files/folder/:id", fileHandler.DeleteFolder)
		authorized.GET("/files/list", fileHandler.ListFiles)
		authorized.POST("/files/move", fileHandler.MoveFile)
		authorized.POST("/files/batch-delete", fileHandler.BatchDeleteFiles)
		authorized.POST("/files/batch-move", fileHandler.BatchMoveFiles)
		authorized.DELETE("/files/:id", fileHandler.DeleteFile)

		// 文件标签管理
		authorized.GET("/files/:id/tags", tagHandler.GetFileTags)
		authorized.POST("/files/:id/tags", tagHandler.AddFileTag)
		authorized.DELETE("/files/:id/tags/:tagId", tagHandler.RemoveFileTag)
		authorized.PUT("/files/:id/tags", tagHandler.BatchUpdateFileTags)

		// 媒体文件
			authorized.POST("/files/:id/share", shareHandler.CreateFileShare)
			authorized.POST("/folders/:id/share", shareHandler.CreateFolderShare)
			authorized.GET("/my-shares", shareHandler.GetMyShares)
			authorized.DELETE("/shares/:id", shareHandler.DeleteShare)

			// 分享管理
			authorized.GET("/files/shares", shareHandler.GetUserShares)
			authorized.PUT("/files/shares/:id/renew", shareHandler.RenewShare)
			authorized.PUT("/files/shares/:id/expire", shareHandler.ExpireShare)
			authorized.PUT("/files/shares/:id/expiry", shareHandler.UpdateShareExpiry)
			authorized.PUT("/files/shares/:id/disable", shareHandler.DisableShare)
			authorized.PUT("/files/shares/:id/enable", shareHandler.EnableShare)

		authorized.GET("/files/:id/metadata", mediaHandler.GetMediaInfo)
		authorized.GET("/files/:id/stream", mediaHandler.GetStream)
		authorized.GET("/files/:id/download", mediaHandler.DownloadFile)
			authorized.GET("/files/:id/view", mediaHandler.ViewFile)
			authorized.GET("/files/:id/preview-url", mediaHandler.GetPreviewURL)

		// 临时预览（使用 token 认证，支持 Range 请求）
		r.GET("/files/:id/preview", mediaHandler.PreviewFile)

		// 系统管理
		system := authorized.Group("/system")
		{
			// 用户管理
			system.GET("/user/list", middleware.Permission("system:user:view"), userHandler.List)
			system.POST("/user", middleware.Permission("system:user:add"), userHandler.Create)
			system.PUT("/user/:id", middleware.Permission("system:user:edit"), userHandler.Update)
			system.DELETE("/user/:id", middleware.Permission("system:user:delete"), userHandler.Delete)

			// 角色管理
			system.GET("/role/list", middleware.Permission("system:role:view"), roleHandler.List)
			system.GET("/role/:id", middleware.Permission("system:role:view"), roleHandler.GetDetail)
			system.POST("/role", middleware.Permission("system:role:add"), roleHandler.Create)
			system.PUT("/role/:id", middleware.Permission("system:role:edit"), roleHandler.Update)
			system.DELETE("/role/:id", middleware.Permission("system:role:delete"), roleHandler.Delete)

			// 菜单管理
			system.GET("/menu/list", middleware.Permission("system:menu:view"), menuHandler.List)
			system.GET("/menu/name-exists", middleware.Permission("system:menu:view"), menuHandler.NameExists)
			system.GET("/menu/path-exists", middleware.Permission("system:menu:view"), menuHandler.PathExists)
			system.POST("/menu", middleware.Permission("system:menu:add"), menuHandler.Create)
			system.PUT("/menu/:id", middleware.Permission("system:menu:edit"), menuHandler.Update)
			system.DELETE("/menu/:id", middleware.Permission("system:menu:delete"), menuHandler.Delete)

			// 分组管理
			system.GET("/group/list", middleware.Permission("system:group:view"), groupHandler.List)
			system.POST("/group", middleware.Permission("system:group:add"), groupHandler.Create)
			system.PUT("/group/:id", middleware.Permission("system:group:edit"), groupHandler.Update)
			system.DELETE("/group/:id", middleware.Permission("system:group:delete"), groupHandler.Delete)

			// 安全日志（管理员）
			system.GET("/security-logs", middleware.Permission("system:security:view"), securityLogHandler.GetAllLogs)

			// 实名认证管理（管理员）
			system.GET("/real-name/list", middleware.Permission("system:realname:view"), userRealNameHandler.List)
			system.PUT("/real-name/:id/approve", middleware.Permission("system:realname:review"), userRealNameHandler.Approve)
			system.PUT("/real-name/:id/reject", middleware.Permission("system:realname:review"), userRealNameHandler.Reject)

			// 角色申请管理（管理员）
			system.GET("/role-applications/list", middleware.Permission("system:roleapp:view"), roleApplicationHandler.GetAllList)
			system.PUT("/role-applications/:id/approve", middleware.Permission("system:roleapp:review"), roleApplicationHandler.Approve)
			system.PUT("/role-applications/:id/reject", middleware.Permission("system:roleapp:review"), roleApplicationHandler.Reject)

			// 验证码测试（管理员）
			system.GET("/captcha/test", middleware.Permission("system:captcha:test"), captchaHandler.TestCaptcha)
			system.POST("/captcha/verify", middleware.Permission("system:captcha:test"), captchaHandler.VerifyCaptcha)

			// 系统设置（固定路径必须在 :group 参数路由之前）
			system.GET("/settings", middleware.Permission("system:setting:list"), systemSettingHandler.GetAll)
			system.POST("/settings/test-email", middleware.Permission("system:setting:edit"), systemSettingHandler.TestEmail)
			system.POST("/settings/test-sms", middleware.Permission("system:setting:edit"), systemSettingHandler.TestSMS)
			system.PUT("/settings", middleware.Permission("system:setting:edit"), systemSettingHandler.Update)
			system.GET("/settings/:group", middleware.Permission("system:setting:list"), systemSettingHandler.GetByGroup)
			system.PUT("/settings/:group", middleware.Permission("system:setting:edit"), systemSettingHandler.UpdateByGroup)

			// 风险评分管理
			system.GET("/risk/scores", middleware.Permission("system:setting:list"), riskScoreHandler.GetRiskScores)
			system.GET("/risk/score", middleware.Permission("system:setting:list"), riskScoreHandler.GetRiskScoreByIP)
			system.GET("/risk/stats", middleware.Permission("system:setting:list"), riskScoreHandler.GetRiskScoreStats)
			system.POST("/risk/clear", middleware.Permission("system:setting:edit"), riskScoreHandler.ClearRiskScore)

			// 标签管理
			system.GET("/tags", middleware.Permission("system:setting:list"), tagHandler.GetAllTags)
			system.GET("/tags/grouped", middleware.Permission("system:setting:list"), tagHandler.GetGroupedTags)
			system.GET("/tags/stats", middleware.Permission("system:setting:list"), tagHandler.GetUsageStats)
			system.GET("/tags/key/:key", middleware.Permission("system:setting:list"), tagHandler.GetTagsByKey)
			system.GET("/tags/:id", middleware.Permission("system:setting:list"), tagHandler.GetTagByID)
			system.POST("/tags", middleware.Permission("system:setting:edit"), tagHandler.CreateTag)
			system.PUT("/tags/:id", middleware.Permission("system:setting:edit"), tagHandler.UpdateTag)
			system.DELETE("/tags/:id", middleware.Permission("system:setting:edit"), tagHandler.DeleteTag)

			// 路由规则管理
			system.GET("/routing-rules", middleware.Permission("system:setting:list"), routingHandler.GetAllRules)
			system.GET("/routing-rules/:id", middleware.Permission("system:setting:list"), routingHandler.GetRuleByID)
			system.POST("/routing-rules", middleware.Permission("system:setting:edit"), routingHandler.CreateRule)
			system.PUT("/routing-rules/:id", middleware.Permission("system:setting:edit"), routingHandler.UpdateRule)
			system.DELETE("/routing-rules/:id", middleware.Permission("system:setting:edit"), routingHandler.DeleteRule)
			system.PUT("/routing-rules/:id/status", middleware.Permission("system:setting:edit"), routingHandler.UpdateStatus)
			system.PUT("/routing-rules/:id/priority", middleware.Permission("system:setting:edit"), routingHandler.UpdatePriority)
			system.POST("/routing-rules/batch-priority", middleware.Permission("system:setting:edit"), routingHandler.BatchUpdatePriority)
			system.POST("/routing-rules/:id/test", middleware.Permission("system:setting:list"), routingHandler.TestRule)
			system.POST("/routing-rules/test-route", middleware.Permission("system:setting:list"), routingHandler.TestRoute)
		}
	}

	return r
}