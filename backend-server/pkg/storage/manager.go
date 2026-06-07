package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"backend-server/config"
	"backend-server/pkg/database"
)

var (
	currentStorage Storage
	storageMutex   sync.RWMutex
	// storageCache 缓存不同驱动的存储实例，用于访问历史文件
	storageCache = make(map[string]Storage)
)

// InitStorage 初始化存储（启动时调用）
// 优先从 DB 加载配置，如果 DB 无配置则使用 config.yaml 的值
func InitStorage(cfg config.StorageConfig) {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	// 尝试从 DB 加载配置
	dbCfg := loadStorageConfigFromDB()
	if dbCfg != nil {
		cfg = *dbCfg
		log.Printf("[INFO] 从数据库加载存储配置: driver=%s", cfg.Driver)
	} else {
		log.Printf("[INFO] 数据库无存储配置，使用 config.yaml: driver=%s", cfg.Driver)
	}

	s, err := New(cfg)
	if err != nil {
		log.Printf("[WARN] 存储初始化失败 (%s)，回退到本地存储: %v", cfg.Driver, err)
		s = NewLocalStorage(cfg.Local)
	}
	currentStorage = s

	// 缓存当前存储实例
	storageCache[cfg.Driver] = s
	// 本地存储始终可用
	if cfg.Driver != "local" {
		storageCache["local"] = NewLocalStorage(cfg.Local)
	}
}

// GetStorage 获取当前存储实例（线程安全）
func GetStorage() Storage {
	storageMutex.RLock()
	defer storageMutex.RUnlock()
	return currentStorage
}

// GetStorageDriver 获取当前存储驱动名称
func GetStorageDriver() string {
	storageMutex.RLock()
	defer storageMutex.RUnlock()
	if currentStorage == nil {
		return "local"
	}
	// 根据实例类型判断驱动
	switch currentStorage.(type) {
	case *MinIOStorage:
		return "minio"
	case *OSSStorage:
		return "oss"
	case *COStorage:
		return "cos"
	default:
		return "local"
	}
}

// GetStorageByDriver 根据驱动名称获取存储实例
// 用于访问历史文件（切换存储后旧文件仍需可访问）
func GetStorageByDriver(driver string) Storage {
	storageMutex.RLock()
	defer storageMutex.RUnlock()

	// 先从缓存中获取
	if s, ok := storageCache[driver]; ok {
		return s
	}

	// 缓存中没有，使用当前存储（可能是相同驱动）
	return currentStorage
}

// RefreshStorage 从 DB 重新加载存储配置并重建实例
func RefreshStorage() error {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	dbCfg := loadStorageConfigFromDB()
	if dbCfg == nil {
		return fmt.Errorf("数据库中无存储配置")
	}

	s, err := New(*dbCfg)
	if err != nil {
		return fmt.Errorf("重建存储实例失败: %w", err)
	}

	// 旧实例如果是 LocalStorage，不需要特殊清理
	// MinIO/OSS/COS 客户端会被 GC 回收
	currentStorage = s
	log.Printf("[INFO] 存储配置已热重载: driver=%s", dbCfg.Driver)
	return nil
}

// loadStorageConfigFromDB 从 sys_system_settings 表加载存储配置
func loadStorageConfigFromDB() *config.StorageConfig {
	db := database.GetMySQL()
	if db == nil {
		return nil
	}

	rows, err := db.Raw("SELECT `key`, value FROM sys_system_settings WHERE group_key = 'storage' AND deleted_at IS NULL").Rows()
	if err != nil {
		log.Printf("[WARN] 从数据库加载存储配置失败: %v", err)
		return nil
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		settings[key] = value
	}

	if len(settings) == 0 {
		return nil
	}

	cfg := &config.StorageConfig{}

	// 解析 driver
	cfg.Driver = getSettingStr(settings, "storage_driver", "local")

	// 本地存储配置
	cfg.Local.Path = getSettingStr(settings, "storage_local_path", "./uploads")
	cfg.Local.URLPrefix = getSettingStr(settings, "storage_local_url_prefix", "/uploads")

	// MinIO 配置
	cfg.MinIO.Endpoint = getSettingStr(settings, "storage_minio_endpoint", "")
	cfg.MinIO.AccessKey = getSettingStr(settings, "storage_minio_access_key", "")
	cfg.MinIO.SecretKey = getSettingStr(settings, "storage_minio_secret_key", "")
	cfg.MinIO.Bucket = getSettingStr(settings, "storage_minio_bucket", "")
	cfg.MinIO.UseSSL = getSettingBool(settings, "storage_minio_use_ssl", false)

	// OSS 配置
	cfg.OSS.Endpoint = getSettingStr(settings, "storage_oss_endpoint", "")
	cfg.OSS.AccessKeyID = getSettingStr(settings, "storage_oss_access_key_id", "")
	cfg.OSS.AccessKeySecret = getSettingStr(settings, "storage_oss_access_key_secret", "")
	cfg.OSS.Bucket = getSettingStr(settings, "storage_oss_bucket", "")
	cfg.OSS.CDNDomain = getSettingStr(settings, "storage_oss_cdn_domain", "")

	// COS 配置
	cfg.COS.Region = getSettingStr(settings, "storage_cos_region", "")
	cfg.COS.SecretID = getSettingStr(settings, "storage_cos_secret_id", "")
	cfg.COS.SecretKey = getSettingStr(settings, "storage_cos_secret_key", "")
	cfg.COS.Bucket = getSettingStr(settings, "storage_cos_bucket", "")
	cfg.COS.CDNDomain = getSettingStr(settings, "storage_cos_cdn_domain", "")

	return cfg
}

// getSettingStr 从设置 map 中获取字符串值，去除 JSON 引号
func getSettingStr(settings map[string]string, key, defaultVal string) string {
	val, ok := settings[key]
	if !ok || val == "" {
		return defaultVal
	}
	// 去除 JSON 字符串引号
	val = strings.TrimSpace(val)
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		var s string
		if err := json.Unmarshal([]byte(val), &s); err == nil {
			return s
		}
		return val[1 : len(val)-1]
	}
	return val
}

// getSettingBool 从设置 map 中获取布尔值
func getSettingBool(settings map[string]string, key string, defaultVal bool) bool {
	val, ok := settings[key]
	if !ok || val == "" {
		return defaultVal
	}
	val = strings.TrimSpace(val)
	// 去除 JSON 引号
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		val = val[1 : len(val)-1]
	}
	if strings.ToLower(val) == "true" {
		return true
	}
	if strings.ToLower(val) == "false" {
		return false
	}
	return defaultVal
}
