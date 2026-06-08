package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"backend-server/config"
	"backend-server/internal/model"
	"backend-server/pkg/database"
)

var (
	currentStorage Storage
	storageMutex   sync.RWMutex
	// storageCache 缓存不同驱动的存储实例，用于访问历史文件
	storageCache = make(map[string]Storage)

	// 路由引擎相关
	routingEngine *RoutingEngine
	autoTagger    *AutoTagger
)

// InitStorage 初始化存储（启动时调用）
// 优先从 DB 加载配置，如果 DB 无配置则使用 config.yaml 的值
// 本地存储始终初始化，如果有启用的外部存储则使用它作为主要存储
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

	// 本地存储始终初始化
	localStorage := NewLocalStorage(cfg.Local)
	storageCache["local"] = localStorage
	log.Printf("[INFO] 本地存储已初始化: path=%s", cfg.Local.Path)

	// 检查是否有启用的外部存储
	activeDriver := determineActiveDriver(cfg)

	if activeDriver == "local" {
		currentStorage = localStorage
		log.Printf("[INFO] 使用本地存储作为主要存储")
	} else {
		// 尝试初始化外部存储
		cfg.Driver = activeDriver
		s, err := New(cfg)
		if err != nil {
			log.Printf("[WARN] 外部存储初始化失败 (%s)，回退到本地存储: %v", activeDriver, err)
			currentStorage = localStorage
		} else {
			currentStorage = s
			storageCache[activeDriver] = s
			log.Printf("[INFO] 使用 %s 作为主要存储", activeDriver)
		}
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

// determineActiveDriver 确定应该使用哪个存储驱动
// 优先级：COS > OSS > MinIO > Local（按启用顺序，最后启用的优先）
// 如果没有启用任何外部存储，则返回 "local"
func determineActiveDriver(cfg config.StorageConfig) string {
	// 检查是否有启用的外部存储（按优先级）
	// 这里我们检查配置中是否有有效的连接信息
	// 用户通过前端启用某个存储时，会设置 enabled = true

	// 从数据库设置中获取启用状态
	db := database.GetMySQL()
	if db == nil {
		return "local"
	}

	// 查询所有 storage 组的设置
	rows, err := db.Raw("SELECT `key`, value FROM sys_system_settings WHERE group_key = 'storage' AND deleted_at IS NULL").Rows()
	if err != nil {
		log.Printf("[WARN] 查询存储设置失败: %v", err)
		return "local"
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

	// 检查各存储的启用状态（按优先级从高到低）
	// 如果多个存储同时启用，使用优先级最高的
	if isStorageEnabled(settings, "storage_cos_enabled") && hasRequiredConfig(settings, "cos") {
		return "cos"
	}
	if isStorageEnabled(settings, "storage_oss_enabled") && hasRequiredConfig(settings, "oss") {
		return "oss"
	}
	if isStorageEnabled(settings, "storage_minio_enabled") && hasRequiredConfig(settings, "minio") {
		return "minio"
	}

	return "local"
}

// isStorageEnabled 检查存储是否启用
func isStorageEnabled(settings map[string]string, key string) bool {
	val, ok := settings[key]
	if !ok || val == "" {
		return false
	}
	val = strings.TrimSpace(val)
	// 去除 JSON 引号
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		val = val[1 : len(val)-1]
	}
	return strings.ToLower(val) == "true"
}

// hasRequiredConfig 检查存储是否有必要的配置
func hasRequiredConfig(settings map[string]string, driver string) bool {
	switch driver {
	case "minio":
		endpoint := getSettingStr2(settings, "storage_minio_endpoint", "minio_endpoint")
		bucket := getSettingStr2(settings, "storage_minio_bucket", "minio_bucket")
		return endpoint != "" && bucket != ""
	case "oss":
		endpoint := getSettingStr2(settings, "storage_oss_endpoint", "oss_endpoint")
		bucket := getSettingStr2(settings, "storage_oss_bucket", "oss_bucket")
		return endpoint != "" && bucket != ""
	case "cos":
		region := getSettingStr2(settings, "storage_cos_region", "cos_region")
		bucket := getSettingStr2(settings, "storage_cos_bucket", "cos_bucket")
		return region != "" && bucket != ""
	}
	return false
}

// RefreshStorage 从 DB 重新加载存储配置并重建实例
func RefreshStorage() error {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	dbCfg := loadStorageConfigFromDB()
	if dbCfg == nil {
		return fmt.Errorf("数据库中无存储配置")
	}

	// 本地存储始终初始化
	localStorage := NewLocalStorage(dbCfg.Local)
	storageCache["local"] = localStorage

	// 检查是否有启用的外部存储
	activeDriver := determineActiveDriver(*dbCfg)

	if activeDriver == "local" {
		currentStorage = localStorage
		log.Printf("[INFO] 存储配置已热重载: driver=local")
		return nil
	}

	// 尝试初始化外部存储
	dbCfg.Driver = activeDriver
	s, err := New(*dbCfg)
	if err != nil {
		log.Printf("[WARN] 外部存储初始化失败 (%s)，回退到本地存储: %v", activeDriver, err)
		currentStorage = localStorage
		return nil
	}

	// 旧实例如果是 LocalStorage，不需要特殊清理
	// MinIO/OSS/COS 客户端会被 GC 回收
	currentStorage = s
	storageCache[activeDriver] = s
	log.Printf("[INFO] 存储配置已热重载: driver=%s", activeDriver)
	return nil
}

// InitRoutingEngine 初始化路由引擎
func InitRoutingEngine() error {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	rules, err := loadRoutingRulesFromDB()
	if err != nil {
		return fmt.Errorf("加载路由规则失败: %w", err)
	}

	routingEngine = NewRoutingEngine(rules)
	autoTagger = NewAutoTagger()
	log.Printf("[INFO] 路由引擎已初始化，共 %d 条规则", len(rules))
	return nil
}

// RefreshRoutingEngine 刷新路由引擎（路由规则变更时调用）
func RefreshRoutingEngine() error {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	rules, err := loadRoutingRulesFromDB()
	if err != nil {
		return fmt.Errorf("加载路由规则失败: %w", err)
	}

	routingEngine = NewRoutingEngine(rules)
	log.Printf("[INFO] 路由引擎已刷新，共 %d 条规则", len(rules))
	return nil
}

// GetRoutingEngine 获取路由引擎实例
func GetRoutingEngine() *RoutingEngine {
	storageMutex.RLock()
	defer storageMutex.RUnlock()
	return routingEngine
}

// GetAutoTagger 获取自动打标签器实例
func GetAutoTagger() *AutoTagger {
	storageMutex.RLock()
	defer storageMutex.RUnlock()
	return autoTagger
}

// Route 根据文件信息路由到目标存储
func Route(filename, contentType, source string) (*RoutingResult, []RoutingTag, error) {
	storageMutex.RLock()
	defer storageMutex.RUnlock()

	if routingEngine == nil || autoTagger == nil {
		return nil, nil, fmt.Errorf("路由引擎未初始化")
	}

	// 1. 自动生成标签
	tags := autoTagger.GenerateTags(filename, contentType, source)

	// 2. 匹配路由规则
	result := routingEngine.Match(tags)
	if result == nil {
		return nil, nil, fmt.Errorf("no matching routing rule")
	}

	return result, tags, nil
}

// RouteWithPurpose 根据文件信息和用途路由到目标存储
func RouteWithPurpose(filename, contentType, source, purpose string) (*RoutingResult, []RoutingTag, error) {
	storageMutex.RLock()
	defer storageMutex.RUnlock()

	if routingEngine == nil || autoTagger == nil {
		return nil, nil, fmt.Errorf("路由引擎未初始化")
	}

	// 1. 自动生成标签（包含用途）
	tags := autoTagger.GenerateTagsWithPurpose(filename, contentType, source, purpose)

	// 2. 匹配路由规则
	result := routingEngine.Match(tags)
	if result == nil {
		return nil, nil, fmt.Errorf("no matching routing rule")
	}

	return result, tags, nil
}

// RouteWithTags 根据已有标签路由
func RouteWithTags(tags []RoutingTag) (*RoutingResult, error) {
	storageMutex.RLock()
	defer storageMutex.RUnlock()

	if routingEngine == nil {
		return nil, fmt.Errorf("路由引擎未初始化")
	}

	result := routingEngine.Match(tags)
	if result == nil {
		return nil, fmt.Errorf("no matching routing rule")
	}
	return result, nil
}

// GetStorageByRouting 根据路由结果获取存储实例
func GetStorageByRouting(result *RoutingResult) Storage {
	storageMutex.RLock()
	defer storageMutex.RUnlock()

	switch result.Driver {
	case "local":
		if s, ok := storageCache["local"]; ok {
			return s
		}
		return currentStorage
	case "minio":
		// 对于 MinIO，可能需要根据 bucket 创建不同的实例
		// 暂时返回缓存的实例
		if s, ok := storageCache["minio"]; ok {
			return s
		}
		return currentStorage
	case "oss":
		if s, ok := storageCache["oss"]; ok {
			return s
		}
		return currentStorage
	case "cos":
		if s, ok := storageCache["cos"]; ok {
			return s
		}
		return currentStorage
	default:
		return currentStorage
	}
}

// loadRoutingRulesFromDB 从数据库加载路由规则
func loadRoutingRulesFromDB() ([]model.TagRouting, error) {
	db := database.GetMySQL()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var rules []model.TagRouting
	if err := db.Where("status = ?", 1).Order("priority DESC, id ASC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
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

	// MinIO 配置 - 同时支持带前缀和不带前缀的 key
	cfg.MinIO.Endpoint = getSettingStr2(settings, "storage_minio_endpoint", "minio_endpoint")
	cfg.MinIO.AccessKey = getSettingStr2(settings, "storage_minio_access_key", "minio_access_key")
	cfg.MinIO.SecretKey = getSettingStr2(settings, "storage_minio_secret_key", "minio_secret_key")
	cfg.MinIO.Bucket = getSettingStr2(settings, "storage_minio_bucket", "minio_bucket")
	cfg.MinIO.UseSSL = getSettingBool2(settings, "storage_minio_use_ssl", "minio_use_ssl")

	// OSS 配置 - 同时支持带前缀和不带前缀的 key
	cfg.OSS.Endpoint = getSettingStr2(settings, "storage_oss_endpoint", "oss_endpoint")
	cfg.OSS.AccessKeyID = getSettingStr2(settings, "storage_oss_access_key_id", "oss_access_key_id")
	cfg.OSS.AccessKeySecret = getSettingStr3(settings, "storage_oss_access_key_secret", "oss_access_key_secret", "oss_secret_key")
	cfg.OSS.Bucket = getSettingStr2(settings, "storage_oss_bucket", "oss_bucket")
	cfg.OSS.CDNDomain = getSettingStr2(settings, "storage_oss_cdn_domain", "oss_cdn_domain")

	// COS 配置 - 同时支持带前缀和不带前缀的 key
	cfg.COS.Region = getSettingStr2(settings, "storage_cos_region", "cos_region")
	cfg.COS.SecretID = getSettingStr2(settings, "storage_cos_secret_id", "cos_secret_id")
	cfg.COS.SecretKey = getSettingStr2(settings, "storage_cos_secret_key", "cos_secret_key")
	cfg.COS.Bucket = getSettingStr2(settings, "storage_cos_bucket", "cos_bucket")
	cfg.COS.CDNDomain = getSettingStr2(settings, "storage_cos_cdn_domain", "cos_cdn_domain")

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

// getSettingStr2 从设置 map 中获取字符串值，尝试两个 key
func getSettingStr2(settings map[string]string, key1, key2 string) string {
	if val := getSettingStr(settings, key1, ""); val != "" {
		return val
	}
	return getSettingStr(settings, key2, "")
}

// getSettingStr3 从设置 map 中获取字符串值，尝试三个 key
func getSettingStr3(settings map[string]string, key1, key2, key3 string) string {
	if val := getSettingStr(settings, key1, ""); val != "" {
		return val
	}
	if val := getSettingStr(settings, key2, ""); val != "" {
		return val
	}
	return getSettingStr(settings, key3, "")
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

// getSettingBool2 从设置 map 中获取布尔值，尝试两个 key
func getSettingBool2(settings map[string]string, key1, key2 string) bool {
	val1, ok1 := settings[key1]
	if ok1 && val1 != "" {
		return getSettingBool(settings, key1, false)
	}
	return getSettingBool(settings, key2, false)
}
