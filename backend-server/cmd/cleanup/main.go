package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"backend-server/config"
	"backend-server/pkg/database"
	"backend-server/pkg/logger"
	"backend-server/pkg/storage"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	configPath := "config/config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	// 初始化数据库
	if err := database.InitMySQL(cfg.MySQL); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	db := database.GetMySQL()

	// 初始化存储管理器
	storage.InitStorage(cfg.Storage)

	ctx := context.Background()

	// 查询孤立的文件资产（没有对应文件条目的资产）
	type OrphanedAsset struct {
		ID          uint   `gorm:"column:id"`
		ObjectKey   string `gorm:"column:object_key"`
		StorageType string `gorm:"column:storage_type"`
		FileSize    int64  `gorm:"column:file_size"`
		ContentType string `gorm:"column:content_type"`
	}

	var orphanedAssets []OrphanedAsset
	query := `
		SELECT a.id, a.object_key, a.storage_type, a.file_size, a.content_type
		FROM sys_file_assets a
		LEFT JOIN sys_file_entries e ON a.id = e.file_asset_id
		WHERE e.id IS NULL
	`
	if err := db.Raw(query).Scan(&orphanedAssets).Error; err != nil {
		log.Fatalf("查询孤立资产失败: %v", err)
	}

	fmt.Printf("找到 %d 个孤立的文件资产\n", len(orphanedAssets))

	// 按存储类型分组统计
	stats := make(map[string]struct {
		count int
		size  int64
	})
	for _, asset := range orphanedAssets {
		s := stats[asset.StorageType]
		s.count++
		s.size += asset.FileSize
		stats[asset.StorageType] = s
	}

	fmt.Println("\n存储统计:")
	for storageType, s := range stats {
		fmt.Printf("  %s: %d 个文件, %.2f MB\n", storageType, s.count, float64(s.size)/1024/1024)
	}

	// 确认删除
	if len(os.Args) > 1 && os.Args[1] == "--confirm" {
		fmt.Println("\n开始删除孤立文件...")

		deletedCount := 0
		failedCount := 0
		deletedSize := int64(0)

		for _, asset := range orphanedAssets {
			// 获取对应存储驱动的实例
			storageDriver := storage.GetStorageByDriver(asset.StorageType)

			// 从存储中删除文件
			if err := storageDriver.Delete(ctx, asset.ObjectKey); err != nil {
				logger.Warn("删除存储文件失败",
					zap.String("object_key", asset.ObjectKey),
					zap.String("storage_type", asset.StorageType),
					zap.Error(err),
				)
				failedCount++
				continue
			}

			// 从数据库中删除资产记录
			if err := db.Exec("DELETE FROM sys_file_assets WHERE id = ?", asset.ID).Error; err != nil {
				logger.Warn("删除资产记录失败",
					zap.Uint("id", asset.ID),
					zap.Error(err),
				)
				failedCount++
				continue
			}

			deletedCount++
			deletedSize += asset.FileSize
		}

		fmt.Printf("\n删除完成:\n")
		fmt.Printf("  成功: %d 个文件, %.2f MB\n", deletedCount, float64(deletedSize)/1024/1024)
		fmt.Printf("  失败: %d 个文件\n", failedCount)
	} else {
		fmt.Println("\n⚠️  这是预览模式，不会实际删除文件")
		fmt.Println("要执行删除，请运行: go run cmd/cleanup/main.go --confirm")
	}
}
