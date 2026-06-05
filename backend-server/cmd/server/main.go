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

	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 初始化日志
	if err := logger.Init(cfg.Log); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 初始化验证码加密密钥
	if cfg.Captcha.Secret != "" {
		captcha.SetSecret([]byte(cfg.Captcha.Secret))
	}

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

	// 7. 启动 WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// 8. 初始化路由
	r := router.Setup(cfg, hub)

	// 9. 启动 HTTP 服务
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

	// 10. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务...")

	// 等待最多 5 秒处理完当前请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务关闭异常", zap.Error(err))
	}
	logger.Info("服务已关闭")
}
