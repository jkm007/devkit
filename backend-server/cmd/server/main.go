package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-server/config"
	"backend-server/internal/middleware"
	"backend-server/internal/router"
	"backend-server/internal/service"
	"backend-server/internal/ws"
	"backend-server/migrations"
	"backend-server/pkg/captcha"
	"backend-server/pkg/database"
	"backend-server/pkg/logger"
	"backend-server/pkg/oauth"
	"backend-server/pkg/storage"

	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 1.5 校验敏感密钥配置（必须通过环境变量设置）
	validateSecrets(cfg)

	// 2. 初始化日志
	if err := logger.Init(cfg.Log); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 初始化验证码加密密钥
	if cfg.Captcha.Secret != "" {
		captcha.SetSecret([]byte(cfg.Captcha.Secret))
	}

	// 初始化 OAuth 提供商
	initOAuthProviders(cfg)

	// 初始化风险评分配置获取器
	middleware.SetRiskConfigGetter(func() *middleware.RiskConfigGetter {
		rc := service.GetRiskConfig()
		rules := make([]middleware.RiskRuleItem, len(rc.Rules))
		for i, r := range rc.Rules {
			rules[i] = middleware.RiskRuleItem{
				Key:       r.Key,
				Enabled:   r.Enabled,
				Score:     r.Score,
				Threshold: r.Threshold,
				Keywords:  r.Keywords,
			}
		}
		return &middleware.RiskConfigGetter{
			Enabled:      rc.Enabled,
			TriggerScore: rc.TriggerScore,
			BlockScore:   rc.BlockScore,
			DecayMinutes: rc.DecayMinutes,
			DecayRate:    rc.DecayRate,
			Paths:        rc.Paths,
			Rules:        rules,
		}
	})

	defer logger.Sync()

	// 3. 初始化数据库
	if err := database.InitMySQL(cfg.MySQL); err != nil {
		logger.Fatal("初始化 MySQL 失败", zap.Error(err))
	}
	defer database.CloseMySQL()

	// 4. 初始化 Redis
	if err := database.InitRedis(cfg.Redis); err != nil {
		logger.Fatal("初始化 Redis 失败", zap.Error(err))
	}
	defer database.CloseRedis()

	// 5. 数据库迁移
	if err := migrations.Run(database.GetMySQL()); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 6. 初始化默认数据
	authService := service.NewAuthService()
	if err := authService.InitDefaultUsers(); err != nil {
		logger.Fatal("初始化默认用户失败", zap.Error(err))
	}

	// 7. 初始化默认存储配置
	if err := service.InitDefaultStorageSettings(cfg.Storage); err != nil {
		logger.Fatal("初始化默认存储配置失败", zap.Error(err))
	}

	// 7.1 初始化默认存储桶
	if err := service.InitDefaultStorageBuckets(); err != nil {
		logger.Fatal("初始化默认存储桶失败", zap.Error(err))
	}

	// 8. 启动 WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// 8.1 从数据库加载文件类型规则到 AutoTagger
	initFileTypeRules()

	// 9. 初始化存储（优先从 DB 加载配置）
	storage.InitStorage(cfg.Storage)

	// 9.1 初始化路由引擎
	if err := storage.InitRoutingEngine(); err != nil {
		logger.Error("初始化路由引擎失败，使用默认路由", zap.Error(err))
	}

	// 10. 初始化路由
	r := router.Setup(cfg, hub)

	// 10. 启动 HTTP 服务
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.Info(fmt.Sprintf("服务启动，监听端口 %d", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	// 11. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务...")

	// 停止 WebSocket Hub
	hub.Stop()

	// 等待最多 5 秒处理完当前请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务关闭异常", zap.Error(err))
	}
	logger.Info("服务已关闭")
}

// initOAuthProviders 初始化 OAuth 登录提供商
func initOAuthProviders(cfg *config.Config) {
	// GitHub
	if cfg.OAuth.GitHub.ClientID != "" && cfg.OAuth.GitHub.ClientSecret != "" {
		oauth.Register(oauth.NewGitHubProvider(oauth.ProviderConfig{
			ClientID:     cfg.OAuth.GitHub.ClientID,
			ClientSecret: cfg.OAuth.GitHub.ClientSecret,
			RedirectURL:  cfg.OAuth.GitHub.RedirectURL,
		}))
		logger.Info("OAuth 提供商已注册: github")
	}

	// Google
	if cfg.OAuth.Google.ClientID != "" && cfg.OAuth.Google.ClientSecret != "" {
		oauth.Register(oauth.NewGoogleProvider(oauth.ProviderConfig{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			RedirectURL:  cfg.OAuth.Google.RedirectURL,
		}))
		logger.Info("OAuth 提供商已注册: google")
	}

	// WeChat
	if cfg.OAuth.WeChat.ClientID != "" && cfg.OAuth.WeChat.ClientSecret != "" {
		oauth.Register(oauth.NewWeChatProvider(oauth.ProviderConfig{
			ClientID:     cfg.OAuth.WeChat.ClientID,
			ClientSecret: cfg.OAuth.WeChat.ClientSecret,
			RedirectURL:  cfg.OAuth.WeChat.RedirectURL,
		}))
		logger.Info("OAuth 提供商已注册: wechat")
	}
}

// initFileTypeRules 从数据库加载文件类型规则到 AutoTagger
func initFileTypeRules() {
	db := database.GetMySQL()
	if db == nil {
		return
	}

	type ruleRow struct {
		Extension string
		FileType  string
	}

	var rows []ruleRow
	if err := db.Raw("SELECT extension, file_type FROM sys_file_type_rules WHERE status = 1").Scan(&rows).Error; err != nil {
		logger.Info("未找到文件类型规则表，使用默认规则")
		return
	}

	if len(rows) == 0 {
		logger.Info("文件类型规则表为空，使用默认规则")
		return
	}

	rules := make([]storage.FileTypeRuleData, len(rows))
	for i, r := range rows {
		rules[i] = storage.FileTypeRuleData{
			Extension: r.Extension,
			FileType:  r.FileType,
		}
	}

	storage.GetGlobalAutoTagger().LoadFromDB(rules)
}

// validateSecrets 启动时校验敏感密钥，拒绝使用默认占位符值
func validateSecrets(cfg *config.Config) {
	const defaultPrefix = "changeme"

	var missing []string

	if cfg.JWT.Secret == "" || len(cfg.JWT.Secret) >= len(defaultPrefix) && cfg.JWT.Secret[:len(defaultPrefix)] == defaultPrefix {
		missing = append(missing, "JWT_SECRET")
	}
	if cfg.Crypto.AESKey == "" || len(cfg.Crypto.AESKey) >= len(defaultPrefix) && cfg.Crypto.AESKey[:len(defaultPrefix)] == defaultPrefix {
		missing = append(missing, "AES_KEY")
	}
	if cfg.Captcha.Secret == "" || len(cfg.Captcha.Secret) >= len(defaultPrefix) && cfg.Captcha.Secret[:len(defaultPrefix)] == defaultPrefix {
		missing = append(missing, "CAPTCHA_SECRET")
	}

	if len(missing) > 0 {
		log.Fatalf("安全校验失败：以下密钥仍为默认占位符，请通过环境变量设置真实密钥: %v", missing)
	}
}
