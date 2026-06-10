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
	r.Use(middleware.DBRateLimiter())
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
	storageBucketHandler := handler.NewStorageBucketHandler(service.NewStorageBucketService(repository.NewStorageBucketRepo(database.GetMySQL())))
	storageConfigHandler := handler.NewStorageConfigHandler()
	fileTypeRuleHandler := handler.NewFileTypeRuleHandler()
	rateLimitRuleHandler := handler.NewRateLimitRuleHandler()
	recycleBinHandler := handler.NewRecycleBinHandler()
	scheduledTaskHandler := handler.NewScheduledTaskHandler()

	// 健康检查（不需要 /api/v1 前缀）
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

	// ==================== API v1 路由组 ====================
	apiV1 := r.Group("/api/v1")

	// 公开接口（无需认证）
	apiV1.GET("/system/settings/public", systemSettingHandler.GetPublic)

	// 分享访问（公开，限流由数据库规则管理）
	sharePublic := apiV1.Group("/share")
	{
		sharePublic.GET("/:code", shareHandler.GetShareInfo)
		sharePublic.GET("/:code/files", shareHandler.GetShareFolderFiles)
		sharePublic.GET("/:code/file", shareHandler.GetShareFile)
		sharePublic.GET("/:code/file/:fileId", shareHandler.GetShareFile)
	}

	// 认证接口（无需认证，限流由数据库规则管理）
	captchaHandler := handler.NewCaptchaHandler()
	verifyCodeHandler := handler.NewVerifyCodeHandler()
	auth := apiV1.Group("/auth")
	{
		auth.GET("/captcha", captchaHandler.GetCaptcha)
		auth.POST("/login", authHandler.Login)
		auth.POST("/login-by-email", authHandler.LoginByEmail)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.GET("/oauth/authorize", oauthHandler.GetBindURL)
		auth.GET("/oauth/callback", oauthHandler.Callback)
		auth.POST("/send-code", verifyCodeHandler.SendCode)
		auth.POST("/verify-code", verifyCodeHandler.VerifyCode)
		auth.POST("/send-sms-code", verifyCodeHandler.SendSMSCode)
		auth.POST("/login-by-phone", authHandler.LoginByPhone)
		auth.POST("/register", authHandler.Register)
		auth.POST("/reset-password", authHandler.ResetPassword)

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
	authorized := apiV1.Group("")
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

		// 标签查询（所有认证用户可用，用于文件管理的标签筛选）
		authorized.GET("/tags", tagHandler.GetAllTags)
		authorized.GET("/tags/grouped", tagHandler.GetGroupedTags)
		authorized.GET("/tags/key/:key", tagHandler.GetTagsByKey)

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

		// 回收站（必须在 :id 路由之前）
		authorized.GET("/files/recycle/list", recycleBinHandler.List)
		authorized.GET("/files/recycle/count", recycleBinHandler.GetCount)
		authorized.POST("/files/recycle/restore/:id", recycleBinHandler.Restore)
		authorized.POST("/files/recycle/batch-restore", recycleBinHandler.BatchRestore)
		authorized.DELETE("/files/recycle/:id", recycleBinHandler.PermanentDelete)
		authorized.POST("/files/recycle/batch-delete", recycleBinHandler.BatchPermanentDelete)
		authorized.DELETE("/files/recycle/empty", recycleBinHandler.Empty)

		// 文件操作（:id 路由）
		authorized.DELETE("/files/:id", fileHandler.DeleteFile)
		authorized.GET("/files/:id/shares", fileHandler.CheckFileShares)

		// 文件标签管理
		authorized.GET("/files/:id/tags", tagHandler.GetFileTags)
		authorized.POST("/files/:id/tags", tagHandler.AddFileTag)
		authorized.DELETE("/files/:id/tags/:tagId", tagHandler.RemoveFileTag)
		authorized.PUT("/files/:id/tags", tagHandler.BatchUpdateFileTags)

		// 分享管理
		authorized.POST("/files/:id/share", shareHandler.CreateFileShare)
		authorized.POST("/folders/:id/share", shareHandler.CreateFolderShare)
		authorized.GET("/my-shares", shareHandler.GetMyShares)
		authorized.DELETE("/shares/:id", shareHandler.DeleteShare)
		authorized.GET("/files/shares", shareHandler.GetUserShares)
		authorized.PUT("/files/shares/:id/renew", shareHandler.RenewShare)
		authorized.PUT("/files/shares/:id/expire", shareHandler.ExpireShare)
		authorized.PUT("/files/shares/:id/expiry", shareHandler.UpdateShareExpiry)
		authorized.PUT("/files/shares/:id/disable", shareHandler.DisableShare)
		authorized.PUT("/files/shares/:id/enable", shareHandler.EnableShare)

		// 媒体文件
		authorized.GET("/files/:id/metadata", mediaHandler.GetMediaInfo)
		authorized.GET("/files/:id/stream", mediaHandler.GetStream)
		authorized.GET("/files/:id/download", mediaHandler.DownloadFile)
		authorized.GET("/files/:id/view", mediaHandler.ViewFile)
		authorized.GET("/files/:id/direct-url", mediaHandler.GetDirectURL)
		authorized.GET("/files/:id/preview-url", mediaHandler.GetPreviewURL)

		// 临时预览（使用 token 认证，支持 Range 请求）
		apiV1.GET("/files/:id/preview", mediaHandler.PreviewFile)

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
			system.GET("/security-logs", middleware.Permission("security:log:view"), securityLogHandler.GetAllLogs)

			// 实名认证管理（管理员）
			system.GET("/real-name/list", middleware.Permission("security:realname:view"), userRealNameHandler.List)
			system.PUT("/real-name/:id/approve", middleware.Permission("security:realname:approve"), userRealNameHandler.Approve)
			system.PUT("/real-name/:id/reject", middleware.Permission("security:realname:reject"), userRealNameHandler.Reject)

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
			system.GET("/risk/scores", middleware.Permission("security:risk:view"), riskScoreHandler.GetRiskScores)
			system.GET("/risk/score", middleware.Permission("security:risk:view"), riskScoreHandler.GetRiskScoreByIP)
			system.GET("/risk/stats", middleware.Permission("security:risk:view"), riskScoreHandler.GetRiskScoreStats)
			system.POST("/risk/clear", middleware.Permission("security:risk:edit"), riskScoreHandler.ClearRiskScore)

			// 标签管理
			system.GET("/tags", middleware.Permission("storage:bucket:view"), tagHandler.GetAllTags)
			system.GET("/tags/grouped", middleware.Permission("storage:bucket:view"), tagHandler.GetGroupedTags)
			system.GET("/tags/stats", middleware.Permission("storage:bucket:view"), tagHandler.GetUsageStats)
			system.GET("/tags/key/:key", middleware.Permission("storage:bucket:view"), tagHandler.GetTagsByKey)
			system.GET("/tags/:id", middleware.Permission("storage:bucket:view"), tagHandler.GetTagByID)
			system.POST("/tags", middleware.Permission("storage:bucket:edit"), tagHandler.CreateTag)
			system.PUT("/tags/:id", middleware.Permission("storage:bucket:edit"), tagHandler.UpdateTag)
			system.DELETE("/tags/:id", middleware.Permission("storage:bucket:edit"), tagHandler.DeleteTag)

			// 路由规则管理
			system.GET("/routing-rules", middleware.Permission("storage:bucket:view"), routingHandler.GetAllRules)
			system.GET("/routing-rules/:id", middleware.Permission("storage:bucket:view"), routingHandler.GetRuleByID)
			system.POST("/routing-rules", middleware.Permission("storage:bucket:edit"), routingHandler.CreateRule)
			system.PUT("/routing-rules/:id", middleware.Permission("storage:bucket:edit"), routingHandler.UpdateRule)
			system.DELETE("/routing-rules/:id", middleware.Permission("storage:bucket:edit"), routingHandler.DeleteRule)
			system.PUT("/routing-rules/:id/status", middleware.Permission("storage:bucket:edit"), routingHandler.UpdateStatus)
			system.PUT("/routing-rules/:id/priority", middleware.Permission("storage:bucket:edit"), routingHandler.UpdatePriority)
			system.POST("/routing-rules/batch-priority", middleware.Permission("storage:bucket:edit"), routingHandler.BatchUpdatePriority)
			system.POST("/routing-rules/:id/test", middleware.Permission("storage:bucket:view"), routingHandler.TestRule)
			system.POST("/routing-rules/test-route", middleware.Permission("storage:bucket:view"), routingHandler.TestRoute)

			// 存储桶管理
			system.GET("/storage-buckets", middleware.Permission("storage:bucket:view"), storageBucketHandler.GetAll)
			system.GET("/storage-buckets/default", middleware.Permission("storage:bucket:view"), storageBucketHandler.GetDefault)
			system.GET("/storage-buckets/enabled-drivers", middleware.Permission("storage:bucket:view"), storageBucketHandler.GetEnabledDrivers)
			system.GET("/storage-buckets/driver/:driver", middleware.Permission("storage:bucket:view"), storageBucketHandler.GetByDriver)
			system.GET("/storage-buckets/purpose/:purpose", middleware.Permission("storage:bucket:view"), storageBucketHandler.GetByPurpose)
			system.GET("/storage-buckets/:id", middleware.Permission("storage:bucket:view"), storageBucketHandler.GetByID)
			system.POST("/storage-buckets", middleware.Permission("storage:bucket:edit"), storageBucketHandler.Create)
			system.PUT("/storage-buckets/:id", middleware.Permission("storage:bucket:edit"), storageBucketHandler.Update)
			system.DELETE("/storage-buckets/:id", middleware.Permission("storage:bucket:delete"), storageBucketHandler.Delete)
			system.PUT("/storage-buckets/:id/default", middleware.Permission("storage:bucket:edit"), storageBucketHandler.SetDefault)
			system.POST("/storage-buckets/:id/test", middleware.Permission("storage:bucket:edit"), storageBucketHandler.TestConnection)
			system.POST("/storage-buckets/test-by-driver", middleware.Permission("storage:bucket:edit"), storageBucketHandler.TestConnectionByDriver)

			// 存储连接配置管理
			system.GET("/storage-configs", middleware.Permission("storage:config:view"), storageConfigHandler.GetAll)
			system.GET("/storage-configs/enabled-drivers", middleware.Permission("storage:config:view"), storageConfigHandler.GetEnabledDrivers)
			system.GET("/storage-configs/:id", middleware.Permission("storage:config:view"), storageConfigHandler.GetByID)
			system.POST("/storage-configs", middleware.Permission("storage:config:edit"), storageConfigHandler.Create)
			system.PUT("/storage-configs/:id", middleware.Permission("storage:config:edit"), storageConfigHandler.Update)
			system.DELETE("/storage-configs/:id", middleware.Permission("storage:config:delete"), storageConfigHandler.Delete)
			system.PUT("/storage-configs/:id/default", middleware.Permission("storage:config:edit"), storageConfigHandler.SetDefault)
			system.POST("/storage-configs/:id/test", middleware.Permission("storage:config:edit"), storageConfigHandler.TestConnection)
			system.POST("/storage-configs/test-by-data", middleware.Permission("storage:config:edit"), storageConfigHandler.TestConnectionByData)

			// 限流规则管理
			system.GET("/rate-limit-rules", middleware.Permission("security:ratelimit:view"), rateLimitRuleHandler.List)
			system.GET("/rate-limit-rules/:id", middleware.Permission("security:ratelimit:view"), rateLimitRuleHandler.GetByID)
			system.POST("/rate-limit-rules", middleware.Permission("security:ratelimit:edit"), rateLimitRuleHandler.Create)
			system.PUT("/rate-limit-rules/:id", middleware.Permission("security:ratelimit:edit"), rateLimitRuleHandler.Update)
			system.DELETE("/rate-limit-rules/:id", middleware.Permission("security:ratelimit:delete"), rateLimitRuleHandler.Delete)
			system.PUT("/rate-limit-rules/:id/status", middleware.Permission("security:ratelimit:edit"), rateLimitRuleHandler.UpdateStatus)

			// 文件类型规则管理
			system.GET("/file-type-rules", middleware.Permission("storage:file-type:view"), fileTypeRuleHandler.GetAll)
			system.GET("/file-type-rules/grouped", middleware.Permission("storage:file-type:view"), fileTypeRuleHandler.GetGrouped)
			system.POST("/file-type-rules", middleware.Permission("storage:file-type:edit"), fileTypeRuleHandler.Create)
			system.PUT("/file-type-rules/:id", middleware.Permission("storage:file-type:edit"), fileTypeRuleHandler.Update)
			system.DELETE("/file-type-rules/:id", middleware.Permission("storage:file-type:delete"), fileTypeRuleHandler.Delete)
			system.POST("/file-type-rules/refresh", middleware.Permission("storage:file-type:edit"), fileTypeRuleHandler.RefreshAutoTagger)

			// 定时任务管理
			system.GET("/scheduled-tasks", middleware.Permission("system:task:view"), scheduledTaskHandler.List)
			system.POST("/scheduled-tasks", middleware.Permission("system:task:edit"), scheduledTaskHandler.Create)
			system.GET("/scheduled-tasks/:id", middleware.Permission("system:task:view"), scheduledTaskHandler.GetByID)
			system.PUT("/scheduled-tasks/:id", middleware.Permission("system:task:edit"), scheduledTaskHandler.Update)
			system.PUT("/scheduled-tasks/:id/enabled", middleware.Permission("system:task:edit"), scheduledTaskHandler.UpdateEnabled)
			system.DELETE("/scheduled-tasks/:id", middleware.Permission("system:task:delete"), scheduledTaskHandler.Delete)
			system.POST("/scheduled-tasks/:id/run", middleware.Permission("system:task:run"), scheduledTaskHandler.Run)
		}
	}

	return r
}
