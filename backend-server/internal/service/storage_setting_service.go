package service

import (
	"fmt"

	"backend-server/config"
	"backend-server/internal/model"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

// storageDefaultSetting 存储配置项定义
type storageDefaultSetting struct {
	Key         string
	Value       string
	Label       string
	Type        string
	IsSensitive int8
	Sort        int
	Tip         string
}

// InitDefaultStorageSettings 初始化默认存储配置
// 如果 sys_system_settings 表中 storage 组无数据，则从 config.yaml 写入默认值
func InitDefaultStorageSettings(cfg config.StorageConfig) error {
	db := database.GetMySQL()

	// 检查是否已有 storage 配置
	var count int64
	if err := db.Model(&model.SystemSetting{}).Where("group_key = ?", "storage").Count(&count).Error; err != nil {
		return fmt.Errorf("查询存储配置失败: %w", err)
	}
	if count > 0 {
		return nil // 已有配置，不覆盖
	}

	// 构建默认配置项
	settings := buildDefaultStorageSettings(cfg)

	// 批量插入
	return db.Transaction(func(tx *gorm.DB) error {
		for _, s := range settings {
			setting := model.SystemSetting{
				GroupKey:    "storage",
				Key:         s.Key,
				Value:       s.Value,
				Label:       s.Label,
				Type:        s.Type,
				IsSensitive: s.IsSensitive,
				Sort:        s.Sort,
				Tip:         s.Tip,
			}
			if err := tx.Create(&setting).Error; err != nil {
				return fmt.Errorf("插入存储配置 %s 失败: %w", s.Key, err)
			}
		}
		return nil
	})
}

// buildDefaultStorageSettings 根据 config.yaml 构建默认存储配置
func buildDefaultStorageSettings(cfg config.StorageConfig) []storageDefaultSetting {
	return []storageDefaultSetting{
		// 存储驱动选择（保留兼容，但不再使用）
		{Key: "storage_driver", Value: jsonStr(cfg.Driver), Label: "存储驱动", Type: "select", Sort: 1, Tip: "可选值: local, minio, oss, cos（已弃用，请使用下方启用开关）"},
		// 本地存储（始终启用）
		{Key: "storage_local_enabled", Value: "true", Label: "本地存储", Type: "boolean", Sort: 2, Tip: "本地存储始终启用，作为默认存储"},
		{Key: "storage_local_path", Value: jsonStr(cfg.Local.Path), Label: "本地存储路径", Type: "string", Sort: 10, Tip: "本地文件存储的目录路径"},
		{Key: "storage_local_url_prefix", Value: jsonStr(cfg.Local.URLPrefix), Label: "本地 URL 前缀", Type: "string", Sort: 11, Tip: "访问本地文件的 URL 前缀"},
		// MinIO
		{Key: "storage_minio_enabled", Value: "false", Label: "启用 MinIO", Type: "boolean", Sort: 19, Tip: "启用后将使用 MinIO 作为主要存储"},
		{Key: "storage_minio_endpoint", Value: jsonStr(cfg.MinIO.Endpoint), Label: "MinIO 地址", Type: "string", Sort: 20, Tip: "MinIO 服务地址，如 localhost:9000"},
		{Key: "storage_minio_access_key", Value: jsonStr(cfg.MinIO.AccessKey), Label: "MinIO Access Key", Type: "string", IsSensitive: 1, Sort: 21},
		{Key: "storage_minio_secret_key", Value: jsonStr(cfg.MinIO.SecretKey), Label: "MinIO Secret Key", Type: "string", IsSensitive: 1, Sort: 22},
		{Key: "storage_minio_bucket", Value: jsonStr(cfg.MinIO.Bucket), Label: "MinIO 桶名", Type: "string", Sort: 23},
		{Key: "storage_minio_use_ssl", Value: jsonBool(cfg.MinIO.UseSSL), Label: "MinIO SSL", Type: "boolean", Sort: 24, Tip: "是否使用 HTTPS 连接 MinIO"},
		// OSS
		{Key: "storage_oss_enabled", Value: "false", Label: "启用 OSS", Type: "boolean", Sort: 29, Tip: "启用后将使用阿里云 OSS 作为主要存储"},
		{Key: "storage_oss_endpoint", Value: jsonStr(cfg.OSS.Endpoint), Label: "OSS Endpoint", Type: "string", Sort: 30, Tip: "阿里云 OSS 端点，如 oss-cn-hangzhou.aliyuncs.com"},
		{Key: "storage_oss_access_key_id", Value: jsonStr(cfg.OSS.AccessKeyID), Label: "OSS Access Key ID", Type: "string", IsSensitive: 1, Sort: 31},
		{Key: "storage_oss_access_key_secret", Value: jsonStr(cfg.OSS.AccessKeySecret), Label: "OSS Access Key Secret", Type: "string", IsSensitive: 1, Sort: 32},
		{Key: "storage_oss_bucket", Value: jsonStr(cfg.OSS.Bucket), Label: "OSS 桶名", Type: "string", Sort: 33},
		{Key: "storage_oss_cdn_domain", Value: jsonStr(cfg.OSS.CDNDomain), Label: "OSS CDN 域名", Type: "string", Sort: 34, Tip: "可选，自定义 CDN 域名"},
		// COS
		{Key: "storage_cos_enabled", Value: "false", Label: "启用 COS", Type: "boolean", Sort: 39, Tip: "启用后将使用腾讯云 COS 作为主要存储"},
		{Key: "storage_cos_region", Value: jsonStr(cfg.COS.Region), Label: "COS 地域", Type: "string", Sort: 40, Tip: "如 ap-guangzhou"},
		{Key: "storage_cos_secret_id", Value: jsonStr(cfg.COS.SecretID), Label: "COS Secret ID", Type: "string", IsSensitive: 1, Sort: 41},
		{Key: "storage_cos_secret_key", Value: jsonStr(cfg.COS.SecretKey), Label: "COS Secret Key", Type: "string", IsSensitive: 1, Sort: 42},
		{Key: "storage_cos_bucket", Value: jsonStr(cfg.COS.Bucket), Label: "COS 桶名", Type: "string", Sort: 43},
		{Key: "storage_cos_cdn_domain", Value: jsonStr(cfg.COS.CDNDomain), Label: "COS CDN 域名", Type: "string", Sort: 44, Tip: "可选，自定义 CDN 域名"},
	}
}

// jsonStr 将字符串转为 JSON 格式存储
func jsonStr(s string) string {
	if s == "" {
		return `""`
	}
	return fmt.Sprintf("%q", s)
}

// jsonBool 将布尔值转为 JSON 格式存储
func jsonBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
