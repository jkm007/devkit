package router

import (
	"context"

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
func Setup(ctx context.Context, cfg *config.Config, hub *ws.Hub) *gin.Engine {
	// 设置运行模式
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Logger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.CORS))
	r.Use(middleware.RateLimiter(ctx, cfg.RateLimit))
	r.Use(middleware.DBRateLimiter())
	r.Use(middleware.CSRF())
	r.Use(gin.Recovery())

	// 初始化处理器
	authHandler := handler.NewAuthHandler()
	userHandler := handler.NewUserHandler()
	roleHandler := handler.NewRoleHandler()
	dashboardHandler := handler.NewDashboardHandler()
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
	notificationHandler := handler.NewNotificationHandler()
	userHomeHandler := handler.NewUserHomeHandler()
	studyHandler := handler.NewStudyHandler()
	bannerHandler := handler.NewBannerHandler()
	feedbackHandler := handler.NewQuestionFeedbackHandler()
	mobileCategoryHandler := handler.NewMobileCategoryHandler()

	// 移动端配置
	mobileConfigRepo := repository.NewMobileConfigRepo(database.GetMySQL())
	mobileConfigService := service.NewMobileConfigService(mobileConfigRepo)
	mobileConfigHandler := handler.NewMobileConfigHandler(mobileConfigService)

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
		sharePublic.POST("/:code/verify", shareHandler.VerifySharePassword)
		sharePublic.GET("/:code/files", shareHandler.GetShareFolderFiles)
		sharePublic.GET("/:code/file", shareHandler.GetShareFile)
		sharePublic.GET("/:code/file/:fileId", shareHandler.GetShareFile)
	}

	// 轮播图（公开接口，移动端首页用）
	apiV1.GET("/banners", bannerHandler.GetBanners)

	// 公开文件URL（无需认证，用于轮播图等公开资源）
	apiV1.GET("/files/:id/public-url", mediaHandler.GetPublicURL)
	apiV1.POST("/files/batch-public-url", mediaHandler.BatchGetPublicURL)

	// 移动端配置（公开接口，移动端用）
	apiV1.GET("/mobile/quick-menus", mobileConfigHandler.GetActiveQuickMenus)
	apiV1.GET("/mobile/my-page-menus", mobileConfigHandler.GetActiveMyPageMenus)
	apiV1.GET("/mobile/settings", mobileConfigHandler.GetMobileSettings)

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

	// WebSocket 连接（独立于 JWT 中间件，自行从 query 参数验证 token）
	apiV1.GET("/ws", wsHandler.Handle)

	// 移动端分类树（公开接口，无需认证）
	apiV1.GET("/mobile/category-tree", mobileCategoryHandler.GetCategoryTree)

	// 题库管理处理器
	examCategoryHandler := handler.NewExamCategoryHandler()
	examHandler := handler.NewExamHandler()
	subjectHandler := handler.NewSubjectHandler()
	questionCategoryHandler := handler.NewQuestionCategoryHandler()
	knowledgePointHandler := handler.NewKnowledgePointHandler()
	questionSourceHandler := handler.NewQuestionSourceHandler()
	questionHandler := handler.NewQuestionHandler()
	questionImportHandler := handler.NewQuestionImportHandler()
	questionShareHandler := handler.NewQuestionShareHandler()

	// 公开接口（移动端用，无需认证）
	apiV1.GET("/exam-categories/all", examCategoryHandler.GetAll)

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

		// 用户首页数据
		authorized.GET("/user/home", userHomeHandler.GetHomeData)
		authorized.GET("/user/stats", userHomeHandler.GetUserStats)

		// ==================== 移动端学习 API ====================
		// 题目学习
		authorized.GET("/study/questions", studyHandler.ListQuestions)
		authorized.GET("/study/questions/:id", studyHandler.GetQuestion)
		authorized.POST("/study/questions/:id/favorite", studyHandler.AddFavorite)
		authorized.DELETE("/study/questions/:id/favorite", studyHandler.RemoveFavorite)

		// 题目搜索（移动端）
		authorized.GET("/questions/search", questionHandler.Search)

		// 收藏管理
		authorized.GET("/user/favorites", studyHandler.ListFavorites)

		// 笔记管理
		authorized.GET("/user/notes", studyHandler.ListNotes)
		authorized.POST("/user/notes", studyHandler.CreateNote)
		authorized.PUT("/user/notes/:id", studyHandler.UpdateNote)
		authorized.DELETE("/user/notes/:id", studyHandler.DeleteNote)

		// 练习
		authorized.POST("/study/practice/questions", studyHandler.GetPracticeQuestions)
		authorized.POST("/study/practice/submit", studyHandler.SubmitPractice)
		authorized.GET("/study/practice/history", studyHandler.GetPracticeHistory)

		// 错题本
		authorized.GET("/study/wrong", studyHandler.GetWrongBooks)
		authorized.GET("/study/wrong/stats", studyHandler.GetWrongBookStats)
		authorized.GET("/study/wrong/random", studyHandler.GetWrongBookRandomQuestions)
		authorized.POST("/study/wrong", studyHandler.AddWrongBook)
		authorized.POST("/study/wrong/batch", studyHandler.BatchAddWrongBook)
		authorized.PUT("/study/wrong/:questionId/mastered", studyHandler.MarkWrongMastered)
		authorized.POST("/study/wrong/batch-mastered", studyHandler.BatchMarkMastered)
		authorized.DELETE("/study/wrong/:questionId", studyHandler.DeleteWrongBook)

		// 分类绑定
		authorized.GET("/user/category-bindings", userHandler.GetCategoryBindings)
		authorized.POST("/user/category-bindings", userHandler.BindCategory)
		authorized.PUT("/user/category-bindings/:id", userHandler.SetPrimaryCategory)
		authorized.DELETE("/user/category-bindings/:id", userHandler.UnbindCategory)

		// 智能练习
		authorized.POST("/study/practice/smart", studyHandler.GetSmartPractice)
		authorized.GET("/study/practice/analysis", studyHandler.GetPracticeAnalysis)

		// 题目纠错
		authorized.POST("/study/feedback", feedbackHandler.Create)
		authorized.GET("/study/feedback", feedbackHandler.List)
		authorized.GET("/study/feedback/:id", feedbackHandler.GetDetail)
		authorized.DELETE("/study/feedback/:id", feedbackHandler.Delete)

		// 通知消息
		authorized.GET("/notifications", notificationHandler.List)
		authorized.GET("/notifications/unread-count", notificationHandler.GetUnreadCount)
		authorized.PUT("/notifications/:id/read", notificationHandler.MarkRead)
		authorized.PUT("/notifications/read-all", notificationHandler.MarkAllRead)
		authorized.DELETE("/notifications/:id", notificationHandler.Delete)

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
		authorized.GET("/auth/role-applications/available-roles", roleApplicationHandler.GetAvailableRoles)
		authorized.POST("/auth/role-applications", roleApplicationHandler.Create)
		authorized.GET("/auth/role-applications", roleApplicationHandler.GetMyList)

		// 用户资料
		authorized.GET("/user/privacy", userPrivacyHandler.Get)
		authorized.PUT("/user/privacy", userPrivacyHandler.Update)
		authorized.GET("/user/real-name", userRealNameHandler.GetStatus)
		authorized.POST("/user/real-name", userRealNameHandler.Submit)

		// 文件上传（分片上传）
		authorized.POST("/files/upload/check", middleware.Permission("file:upload"), uploadHandler.CheckUpload)
		authorized.POST("/files/upload/init", middleware.Permission("file:upload"), uploadHandler.InitUpload)
		authorized.POST("/files/upload/part", middleware.Permission("file:upload"), uploadHandler.UploadPart)
		authorized.POST("/files/upload/complete", middleware.Permission("file:upload"), uploadHandler.CompleteUpload)
		authorized.POST("/files/upload/abort", middleware.Permission("file:upload"), uploadHandler.AbortUpload)
		authorized.GET("/files/upload/status", middleware.Permission("file:upload"), uploadHandler.GetUploadStatus)
		authorized.GET("/files/upload/tasks", middleware.Permission("file:upload"), uploadHandler.GetUserUploadTasks)
		authorized.GET("/files/upload/tasks/:id", middleware.Permission("file:upload"), uploadHandler.GetUploadTaskByID)

		// 文件管理
		authorized.POST("/files/folder", middleware.Permission("file:manage"), fileHandler.CreateFolder)
		authorized.GET("/files/tree", middleware.Permission("file:view:own"), fileHandler.GetFolderTree)
		authorized.PUT("/files/folder/:id", middleware.Permission("file:manage"), fileHandler.RenameFolder)
		authorized.DELETE("/files/folder/:id", middleware.Permission("file:manage"), fileHandler.DeleteFolder)
		authorized.GET("/files/list", middleware.Permission("file:view:own"), fileHandler.ListFiles)
		authorized.POST("/files/move", middleware.Permission("file:manage"), fileHandler.MoveFile)
		authorized.POST("/files/batch-delete", middleware.Permission("file:delete"), fileHandler.BatchDeleteFiles)
		authorized.POST("/files/batch-move", middleware.Permission("file:manage"), fileHandler.BatchMoveFiles)

		// 回收站（必须在 :id 路由之前）
		authorized.GET("/files/recycle/list", middleware.Permission("file:view:own"), recycleBinHandler.List)
		authorized.GET("/files/recycle/count", middleware.Permission("file:view:own"), recycleBinHandler.GetCount)
		authorized.POST("/files/recycle/restore/:id", middleware.Permission("file:manage"), recycleBinHandler.Restore)
		authorized.POST("/files/recycle/batch-restore", middleware.Permission("file:manage"), recycleBinHandler.BatchRestore)
		authorized.DELETE("/files/recycle/:id", middleware.Permission("file:delete"), recycleBinHandler.PermanentDelete)
		authorized.POST("/files/recycle/batch-delete", middleware.Permission("file:delete"), recycleBinHandler.BatchPermanentDelete)
		authorized.DELETE("/files/recycle/empty", middleware.Permission("file:delete"), recycleBinHandler.Empty)

		// 文件操作（:id 路由）
		authorized.DELETE("/files/:id", middleware.Permission("file:delete"), fileHandler.DeleteFile)
		authorized.GET("/files/:id/shares", middleware.Permission("file:view:own"), fileHandler.CheckFileShares)

		// 文件标签管理
		authorized.GET("/files/:id/tags", middleware.Permission("file:view:own"), tagHandler.GetFileTags)
		authorized.POST("/files/:id/tags", middleware.Permission("file:manage"), tagHandler.AddFileTag)
		authorized.DELETE("/files/:id/tags/:tagId", middleware.Permission("file:manage"), tagHandler.RemoveFileTag)
		authorized.PUT("/files/:id/tags", middleware.Permission("file:manage"), tagHandler.BatchUpdateFileTags)

		// 分享管理
		authorized.POST("/files/:id/share", middleware.Permission("file:share"), shareHandler.CreateFileShare)
		authorized.POST("/folders/:id/share", middleware.Permission("file:share"), shareHandler.CreateFolderShare)
		authorized.GET("/my-shares", middleware.Permission("file:view:own"), shareHandler.GetMyShares)
		authorized.DELETE("/shares/:id", middleware.Permission("file:share"), shareHandler.DeleteShare)
		authorized.GET("/files/shares", middleware.Permission("file:view:own"), shareHandler.GetUserShares)
		authorized.PUT("/files/shares/:id/renew", middleware.Permission("file:share"), shareHandler.RenewShare)
		authorized.PUT("/files/shares/:id/expire", middleware.Permission("file:share"), shareHandler.ExpireShare)
		authorized.PUT("/files/shares/:id/expiry", middleware.Permission("file:share"), shareHandler.UpdateShareExpiry)
		authorized.PUT("/files/shares/:id/disable", middleware.Permission("file:share"), shareHandler.DisableShare)
		authorized.PUT("/files/shares/:id/enable", middleware.Permission("file:share"), shareHandler.EnableShare)

		// 媒体文件
		authorized.GET("/files/:id/metadata", middleware.Permission("file:view:own"), mediaHandler.GetMediaInfo)
		authorized.GET("/files/:id/stream", middleware.Permission("file:download"), mediaHandler.GetStream)
		authorized.GET("/files/:id/download", middleware.Permission("file:download"), mediaHandler.DownloadFile)
		authorized.GET("/files/:id/view", middleware.Permission("file:view:own"), mediaHandler.ViewFile)
		authorized.GET("/files/:id/direct-url", middleware.Permission("file:download"), mediaHandler.GetDirectURL)
		authorized.GET("/files/:id/preview-url", middleware.Permission("file:view:own"), mediaHandler.GetPreviewURL)
		authorized.GET("/files/:id/hls", middleware.Permission("file:download"), mediaHandler.GetHLS)

		// 临时预览（使用 token 认证，支持 Range 请求）
		// 设计说明：该路由注册在 apiV1 顶层而非 authorized 组中，有意绕过 JWT 认证中间件。
		// 原因：预览 URL 需要支持在浏览器、iframe、邮件等外部场景中直接访问，
		//       使用一次性临时 token（带过期时间）替代 JWT，无需登录即可预览文件。
		//       handler 内部通过 ValidatePreviewToken() 验证 token 合法性和归属权限，安全性由 token 机制保证。
		apiV1.GET("/files/:id/preview", mediaHandler.PreviewFile)

		// 系统管理
		system := authorized.Group("/system")
		{
			// 仪表盘
			system.GET("/dashboard/stats", dashboardHandler.GetStats)

			// 用户管理
			system.GET("/user/list", middleware.Permission("system:user:view"), userHandler.List)
			system.GET("/user/storage-stats", middleware.Permission("system:user:view"), userHandler.GetStorageStats)
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
			system.GET("/role-applications/list", middleware.Permission("system:roleapp:list"), roleApplicationHandler.GetAllList)
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
			system.POST("/storage-configs/mark-orphaned", middleware.Permission("storage:config:edit"), storageConfigHandler.MarkOrphanedAssets)

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

			// 通知公告管理（管理员）
			system.GET("/notifications", middleware.Permission("system:notification:view"), notificationHandler.AdminList)
			system.POST("/notifications/announcement", middleware.Permission("system:notification:publish"), notificationHandler.PublishAnnouncement)

			// 题库管理 - 分类科目
			// 考试大类
			system.GET("/exam-categories", middleware.Permission("question:category"), examCategoryHandler.List)
			system.GET("/exam-categories/all", middleware.Permission("question:category"), examCategoryHandler.GetAll)
			system.GET("/exam-categories/:id", middleware.Permission("question:category"), examCategoryHandler.GetDetail)
			system.POST("/exam-categories", middleware.Permission("question:category:add"), examCategoryHandler.Create)
			system.PUT("/exam-categories/:id", middleware.Permission("question:category:edit"), examCategoryHandler.Update)
			system.DELETE("/exam-categories/:id", middleware.Permission("question:category:delete"), examCategoryHandler.Delete)

			// 具体考试
			system.GET("/exams", middleware.Permission("question:category"), examHandler.List)
			system.GET("/exams/all", middleware.Permission("question:category"), examHandler.GetAll)
			system.GET("/exams/:id", middleware.Permission("question:category"), examHandler.GetDetail)
			system.POST("/exams", middleware.Permission("question:exam:manage"), examHandler.Create)
			system.PUT("/exams/:id", middleware.Permission("question:exam:manage"), examHandler.Update)
			system.DELETE("/exams/:id", middleware.Permission("question:exam:manage"), examHandler.Delete)

			// 科目
			system.GET("/subjects", middleware.Permission("question:category"), subjectHandler.List)
			system.GET("/subjects/all", middleware.Permission("question:category"), subjectHandler.GetAll)
			system.GET("/subjects/:id", middleware.Permission("question:category"), subjectHandler.GetDetail)
			system.POST("/subjects", middleware.Permission("question:subject:manage"), subjectHandler.Create)
			system.PUT("/subjects/:id", middleware.Permission("question:subject:manage"), subjectHandler.Update)
			system.DELETE("/subjects/:id", middleware.Permission("question:subject:manage"), subjectHandler.Delete)

			// 章节分类
			system.GET("/question-categories", middleware.Permission("question:category"), questionCategoryHandler.List)
			system.GET("/question-categories/all", middleware.Permission("question:category"), questionCategoryHandler.GetAll)
			system.GET("/question-categories/:id", middleware.Permission("question:category"), questionCategoryHandler.GetDetail)
			system.POST("/question-categories", middleware.Permission("question:category:add"), questionCategoryHandler.Create)
			system.PUT("/question-categories/:id", middleware.Permission("question:category:edit"), questionCategoryHandler.Update)
			system.DELETE("/question-categories/:id", middleware.Permission("question:category:delete"), questionCategoryHandler.Delete)

			// 知识考点
			system.GET("/knowledge-points", middleware.Permission("question:knowledge"), knowledgePointHandler.List)
			system.GET("/knowledge-points/all", middleware.Permission("question:knowledge"), knowledgePointHandler.GetAll)
			system.GET("/knowledge-points/:id", middleware.Permission("question:knowledge"), knowledgePointHandler.GetDetail)
			system.POST("/knowledge-points", middleware.Permission("question:knowledge:add"), knowledgePointHandler.Create)
			system.PUT("/knowledge-points/:id", middleware.Permission("question:knowledge:edit"), knowledgePointHandler.Update)
			system.DELETE("/knowledge-points/:id", middleware.Permission("question:knowledge:delete"), knowledgePointHandler.Delete)

			// 来源标签
			system.GET("/question-sources", middleware.Permission("question:source:view"), questionSourceHandler.List)
			system.GET("/question-sources/:id", middleware.Permission("question:source:view"), questionSourceHandler.GetDetail)
			system.POST("/question-sources", middleware.Permission("question:source:manage"), questionSourceHandler.Create)
			system.PUT("/question-sources/:id", middleware.Permission("question:source:manage"), questionSourceHandler.Update)
			system.DELETE("/question-sources/:id", middleware.Permission("question:source:manage"), questionSourceHandler.Delete)

			// 题目管理
			system.GET("/questions", middleware.Permission("question:view"), questionHandler.List)
			system.GET("/questions/search", middleware.Permission("question:view"), questionHandler.Search)
			system.GET("/questions/types", middleware.Permission("question:view"), questionHandler.GetTypes)
			system.GET("/questions/stats", middleware.Permission("question:view"), questionHandler.GetStats)
			system.GET("/questions/:id", middleware.Permission("question:view"), questionHandler.GetDetail)
			system.POST("/questions", middleware.Permission("question:create"), questionHandler.Create)
			system.PUT("/questions/:id", middleware.Permission("question:edit"), questionHandler.Update)
			system.DELETE("/questions/:id", middleware.Permission("question:delete"), questionHandler.Delete)
			system.POST("/questions/:id/publish", middleware.Permission("question:publish"), questionHandler.Publish)
			system.POST("/questions/:id/archive", middleware.Permission("question:archive"), questionHandler.Archive)
			system.POST("/questions/:id/submit-audit", middleware.Permission("question:audit:submit"), questionHandler.SubmitAudit)
			system.POST("/questions/:id/audit/approve", middleware.Permission("question:audit:approve"), questionHandler.Approve)
			system.POST("/questions/:id/audit/reject", middleware.Permission("question:audit:reject"), questionHandler.Reject)
			system.POST("/questions/:id/withdraw", middleware.Permission("question:publish"), questionHandler.Withdraw)
			system.POST("/questions/:id/reactivate", middleware.Permission("question:publish"), questionHandler.Reactivate)

			// 题目导入
			system.GET("/question-imports", middleware.Permission("question:import"), questionImportHandler.List)
			system.GET("/question-imports/:id", middleware.Permission("question:import"), questionImportHandler.GetDetail)
			system.GET("/question-imports/:id/items", middleware.Permission("question:import"), questionImportHandler.GetItems)
			system.POST("/question-imports", middleware.Permission("question:import"), questionImportHandler.Create)
			system.DELETE("/question-imports/:id", middleware.Permission("question:import:delete"), questionImportHandler.Delete)

			// 题目分享
			system.GET("/question-shares", middleware.Permission("question:share:view"), questionShareHandler.List)
			system.GET("/question-shares/:id", middleware.Permission("question:share:view"), questionShareHandler.GetDetail)
			system.POST("/question-shares", middleware.Permission("question:share:create"), questionShareHandler.Create)
			system.PUT("/question-shares/:id/disable", middleware.Permission("question:share:disable"), questionShareHandler.Disable)
			system.PUT("/question-shares/:id/enable", middleware.Permission("question:share:enable"), questionShareHandler.Enable)
			system.DELETE("/question-shares/:id", middleware.Permission("question:share:delete"), questionShareHandler.Delete)

			// 题目纠错管理（管理端）
			system.GET("/feedbacks", middleware.Permission("question:feedback:view"), feedbackHandler.AdminList)
			system.PUT("/feedbacks/:id", middleware.Permission("question:feedback:edit"), feedbackHandler.AdminUpdate)

			// 定时任务管理
			system.GET("/scheduled-tasks", middleware.Permission("system:task:view"), scheduledTaskHandler.List)
			system.POST("/scheduled-tasks", middleware.Permission("system:task:edit"), scheduledTaskHandler.Create)
			system.GET("/scheduled-tasks/:id", middleware.Permission("system:task:view"), scheduledTaskHandler.GetByID)
			system.PUT("/scheduled-tasks/:id", middleware.Permission("system:task:edit"), scheduledTaskHandler.Update)
			system.PUT("/scheduled-tasks/:id/enabled", middleware.Permission("system:task:edit"), scheduledTaskHandler.UpdateEnabled)
			system.DELETE("/scheduled-tasks/:id", middleware.Permission("system:task:delete"), scheduledTaskHandler.Delete)
			system.POST("/scheduled-tasks/:id/run", middleware.Permission("system:task:run"), scheduledTaskHandler.Run)

			// 轮播图管理（管理端）
			system.GET("/banners", middleware.Permission("system:banner:view"), bannerHandler.AdminList)
			system.POST("/banners", middleware.Permission("system:banner:add"), bannerHandler.AdminCreate)
			system.PUT("/banners/:id", middleware.Permission("system:banner:edit"), bannerHandler.AdminUpdate)
			system.DELETE("/banners/:id", middleware.Permission("system:banner:delete"), bannerHandler.AdminDelete)
			system.PUT("/banners/:id/status", middleware.Permission("system:banner:edit"), bannerHandler.AdminUpdateStatus)
			system.PUT("/banners/:id/sort", middleware.Permission("system:banner:edit"), bannerHandler.AdminUpdateSort)

			// 移动端快捷菜单管理（管理端）
			system.GET("/quick-menus", middleware.Permission("system:banner:view"), mobileConfigHandler.GetQuickMenus)
			system.POST("/quick-menus", middleware.Permission("system:banner:add"), mobileConfigHandler.CreateQuickMenu)
			system.PUT("/quick-menus/:id", middleware.Permission("system:banner:edit"), mobileConfigHandler.UpdateQuickMenu)
			system.DELETE("/quick-menus/:id", middleware.Permission("system:banner:delete"), mobileConfigHandler.DeleteQuickMenu)

			// 移动端我的页面菜单管理（管理端）
			system.GET("/my-page-menus", middleware.Permission("system:banner:view"), mobileConfigHandler.GetMyPageMenus)
			system.POST("/my-page-menus", middleware.Permission("system:banner:add"), mobileConfigHandler.CreateMyPageMenu)
			system.PUT("/my-page-menus/:id", middleware.Permission("system:banner:edit"), mobileConfigHandler.UpdateMyPageMenu)
			system.DELETE("/my-page-menus/:id", middleware.Permission("system:banner:delete"), mobileConfigHandler.DeleteMyPageMenu)

			// 移动端设置管理（管理端）
			system.GET("/mobile-settings", middleware.Permission("system:banner:view"), mobileConfigHandler.GetMobileSettings)
			system.PUT("/mobile-settings", middleware.Permission("system:banner:edit"), mobileConfigHandler.UpdateMobileSettings)
		}
	}

	return r
}
