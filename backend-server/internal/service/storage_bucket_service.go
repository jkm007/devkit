package service

import (
	"backend-server/config"
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"backend-server/pkg/storage"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// StorageBucketService 存储桶服务
type StorageBucketService struct {
	repo *repository.StorageBucketRepo
}

// NewStorageBucketService 创建存储桶服务
func NewStorageBucketService(repo *repository.StorageBucketRepo) *StorageBucketService {
	return &StorageBucketService{repo: repo}
}

// GetAll 获取所有存储桶
func (s *StorageBucketService) GetAll() ([]model.StorageBucket, error) {
	return s.repo.GetAll()
}

// GetByID 根据ID获取存储桶
func (s *StorageBucketService) GetByID(id int64) (*model.StorageBucket, error) {
	return s.repo.GetByID(id)
}

// GetByDriver 根据驱动获取存储桶
func (s *StorageBucketService) GetByDriver(driver string) ([]model.StorageBucket, error) {
	return s.repo.GetByDriver(driver)
}

// GetByPurpose 根据用途获取存储桶
func (s *StorageBucketService) GetByPurpose(purpose string) ([]model.StorageBucket, error) {
	return s.repo.GetByPurpose(purpose)
}

// GetDefault 获取默认存储桶
func (s *StorageBucketService) GetDefault() (*model.StorageBucket, error) {
	return s.repo.GetDefault()
}

// Create 创建存储桶
// 如果是外部驱动且未填写凭据，自动从存储配置中获取
func (s *StorageBucketService) Create(bucket *model.StorageBucket) error {
	// 检查名称是否存在
	exists, err := s.repo.NameExists(bucket.Name, 0)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("存储桶名称 %s 已存在", bucket.Name)
	}

	// 外部驱动自动填充凭据（如果未填写）
	if bucket.Driver != "local" && bucket.AccessKey == "" {
		endpoint, accessKey, secretKey, region, cdnDomain, useSSL := GetCredentialsForDriver(bucket.Driver)
		if bucket.Endpoint == "" {
			bucket.Endpoint = endpoint
		}
		if bucket.AccessKey == "" {
			bucket.AccessKey = accessKey
		}
		if bucket.SecretKey == "" {
			bucket.SecretKey = secretKey
		}
		if bucket.Region == "" {
			bucket.Region = region
		}
		if bucket.CDNDomain == "" {
			bucket.CDNDomain = cdnDomain
		}
		if !bucket.UseSSL {
			bucket.UseSSL = useSSL
		}
	}

	return s.repo.Create(bucket)
}

// Update 更新存储桶
func (s *StorageBucketService) Update(bucket *model.StorageBucket) error {
	// 检查是否存在
	existing, err := s.repo.GetByID(bucket.ID)
	if err != nil {
		return fmt.Errorf("存储桶不存在: %d", bucket.ID)
	}

	// 检查名称是否重复
	exists, err := s.repo.NameExists(bucket.Name, bucket.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("存储桶名称 %s 已存在", bucket.Name)
	}

	// 敏感字段处理：如果传入的是空字符串或 ******，保留原值
	if bucket.AccessKey == "" || bucket.AccessKey == "******" {
		bucket.AccessKey = existing.AccessKey
	}
	if bucket.SecretKey == "" || bucket.SecretKey == "******" {
		bucket.SecretKey = existing.SecretKey
	}

	// 保留不可修改的字段
	bucket.CreatedAt = existing.CreatedAt

	return s.repo.Update(bucket)
}

// Delete 删除存储桶
func (s *StorageBucketService) Delete(id int64) error {
	// 检查是否默认存储桶
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("存储桶不存在: %d", id)
	}
	if existing.IsDefault {
		return fmt.Errorf("默认存储桶不允许删除")
	}

	return s.repo.Delete(id)
}

// SetDefault 设置默认存储桶
func (s *StorageBucketService) SetDefault(id int64) error {
	// 检查是否存在
	_, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("存储桶不存在: %d", id)
	}

	if err := s.repo.SetDefault(id); err != nil {
		return err
	}

	// 切换默认桶后刷新存储驱动
	return storage.RefreshStorage()
}

// SyncDefaultBuckets 从存储配置同步默认桶到存储桶管理
// 读取 sys_system_settings 中 storage 组的配置，为每个启用的驱动创建/更新默认桶
func SyncDefaultBuckets() error {
	db := database.GetMySQL()

	// 读取 storage 组的所有配置
	rows, err := db.Raw("SELECT `key`, value FROM sys_system_settings WHERE group_key = 'storage' AND deleted_at IS NULL").Rows()
	if err != nil {
		return fmt.Errorf("读取存储配置失败: %w", err)
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

	// 同步每个驱动的默认桶
	drivers := []struct {
		name       string
		enabledKey string
		label      string
	}{
		{"minio", "storage_minio_enabled", "MinIO"},
		{"oss", "storage_oss_enabled", "阿里云 OSS"},
		{"cos", "storage_cos_enabled", "腾讯云 COS"},
	}

	repo := repository.NewStorageBucketRepo(db)

	for _, d := range drivers {
		enabled := getSettingBoolFromMap(settings, d.enabledKey)
		endpoint := getSettingStrFromMap(settings, fmt.Sprintf("storage_%s_endpoint", d.name))
		bucketName := getSettingStrFromMap(settings, fmt.Sprintf("storage_%s_bucket", d.name))

		log.Printf("[INFO] 同步存储桶检查: driver=%s, enabled=%v, endpoint=%s, bucket=%s", d.name, enabled, endpoint, bucketName)

		// 从 storage_bucket 表查找该驱动的桶（优先找默认桶，没有则找任意一个）
		existing, err := repo.GetDefaultByDriver(d.name)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Printf("[WARN] 查询 %s 默认桶失败: %v", d.name, err)
			continue
		}
		if existing == nil {
			// 没有默认桶，查找该驱动的任意一个桶
			buckets, err := repo.GetByDriver(d.name)
			if err == nil && len(buckets) > 0 {
				existing = &buckets[0]
			}
		}

		if !enabled {
			// 未启用：如果存在默认桶，标记为禁用
			if existing != nil && existing.Status == 1 {
				existing.Status = 0
				existing.Description = fmt.Sprintf("%s 未启用，请在存储配置中启用", d.label)
				if err := repo.Update(existing); err != nil {
					log.Printf("[ERROR] 禁用 %s 默认桶失败: %v", d.name, err)
				}
			}
			continue
		}

		// 已启用：构建桶配置（不自动设为默认，由用户手动选择）
		newBucket := &model.StorageBucket{
			Driver:     d.name,
			Endpoint:   endpoint,
			Bucket:     bucketName,
			Purpose:    "file",
			IsDefault:  existing != nil && existing.IsDefault, // 保留已有的默认状态
			Status:     1,
			AccessKey:  getSettingStrFromMap(settings, fmt.Sprintf("storage_%s_access_key", d.name)),
			SecretKey:  getSettingStrFromMap(settings, fmt.Sprintf("storage_%s_secret_key", d.name)),
			CDNDomain:  getSettingStrFromMap(settings, fmt.Sprintf("storage_%s_cdn_domain", d.name)),
			UseSSL:     getSettingBoolFromMap(settings, fmt.Sprintf("storage_%s_use_ssl", d.name)),
			Description: fmt.Sprintf("由存储配置自动同步的 %s 桶", d.label),
		}

		// OSS 特殊字段
		if d.name == "oss" {
			newBucket.AccessKey = getSettingStrFromMap(settings, "storage_oss_access_key_id")
			newBucket.SecretKey = getSettingStrFromMap(settings, "storage_oss_access_key_secret")
		}
		// COS 特殊字段
		if d.name == "cos" {
			newBucket.AccessKey = getSettingStrFromMap(settings, "storage_cos_secret_id")
			newBucket.SecretKey = getSettingStrFromMap(settings, "storage_cos_secret_key")
			newBucket.Region = getSettingStrFromMap(settings, "storage_cos_region")
		}

		// AccessKey 完全脱敏，日志中不打印任何 AccessKey 内容
		ak := "******"
		if newBucket.AccessKey == "" {
			ak = "(空)"
		}
		log.Printf("[INFO] 同步存储桶: driver=%s, endpoint=%s, bucket=%s, accessKey=%s", d.name, newBucket.Endpoint, newBucket.Bucket, ak)

		if existing != nil {
			// 更新：保留用户自定义的名称和路径前缀
			newBucket.ID = existing.ID
			newBucket.Name = existing.Name
			newBucket.PathPrefix = existing.PathPrefix
			newBucket.CreatedAt = existing.CreatedAt
			if newBucket.Name == "" {
				newBucket.Name = fmt.Sprintf("%s 默认桶", d.label)
			}
			if err := repo.Update(newBucket); err != nil {
				log.Printf("[ERROR] 更新 %s 默认桶失败: %v", d.name, err)
			}
		} else {
			// 新建
			newBucket.Name = fmt.Sprintf("%s 默认桶", d.label)
			if err := repo.Create(newBucket); err != nil {
				log.Printf("[ERROR] 创建 %s 默认桶失败: %v", d.name, err)
			} else {
				log.Printf("[INFO] 已创建 %s 默认桶: %s", d.name, newBucket.Name)
			}
		}
	}

	// 同步本地存储的默认桶
	localPath := getSettingStrFromMap(settings, "storage_local_path")
	localExisting, err := repo.GetDefaultByDriver("local")
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("[WARN] 查询本地默认桶失败: %v", err)
	} else if localExisting == nil {
		// 本地默认桶不存在，创建
		if err := repo.Create(&model.StorageBucket{
			Name:        "本地默认存储",
			Driver:      "local",
			PathPrefix:  localPath,
			Purpose:     "file",
			IsDefault:   true,
			Status:      1,
			Description: "系统内置的本地文件存储",
		}); err != nil {
			log.Printf("[ERROR] 创建本地默认桶失败: %v", err)
		}
	}

	return nil
}

// getSettingStrFromMap 从设置 map 中获取字符串值（去除 JSON 引号）
// 自动尝试带 storage_ 前缀和不带前缀两种 key
func getSettingStrFromMap(settings map[string]string, key string) string {
	// 先尝试原始 key
	val, ok := settings[key]
	if !ok || val == "" {
		// 尝试去掉 storage_ 前缀
		altKey := strings.TrimPrefix(key, "storage_")
		if altKey != key {
			val, ok = settings[altKey]
		}
	}
	if !ok || val == "" {
		// 尝试加上 storage_ 前缀
		if !strings.HasPrefix(key, "storage_") {
			altKey := "storage_" + key
			val, ok = settings[altKey]
		}
	}
	if !ok {
		return ""
	}
	val = strings.TrimSpace(val)
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		val = val[1 : len(val)-1]
	}
	return val
}

// getSettingBoolFromMap 从设置 map 中获取布尔值
func getSettingBoolFromMap(settings map[string]string, key string) bool {
	val := getSettingStrFromMap(settings, key)
	return strings.ToLower(val) == "true"
}

// GetCredentialsForDriver 从 sys_storage_config 获取指定驱动的连接凭据
// 优先使用默认配置，没有默认则使用第一个启用的配置
func GetCredentialsForDriver(driver string) (endpoint, accessKey, secretKey, region, cdnDomain string, useSSL bool) {
	db := database.GetMySQL()
	if db == nil {
		return
	}

	// 先查默认配置
	var cfg struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		Region    string
		CDNDomain string
		UseSSL    bool
	}
	err := db.Raw("SELECT endpoint, access_key, secret_key, region, cdn_domain, use_ssl FROM sys_storage_config WHERE driver = ? AND status = 1 ORDER BY is_default DESC LIMIT 1", driver).Scan(&cfg).Error
	if err != nil {
		// 回退到旧的 sys_system_settings
		return getCredentialsFromSettings(driver)
	}

	endpoint = cfg.Endpoint
	accessKey = cfg.AccessKey
	secretKey = cfg.SecretKey
	region = cfg.Region
	cdnDomain = cfg.CDNDomain
	useSSL = cfg.UseSSL
	return
}

// getCredentialsFromSettings 从旧的 sys_system_settings 获取凭据（兼容回退）
func getCredentialsFromSettings(driver string) (endpoint, accessKey, secretKey, region, cdnDomain string, useSSL bool) {
	db := database.GetMySQL()
	if db == nil {
		return
	}

	rows, err := db.Raw("SELECT `key`, value FROM sys_system_settings WHERE group_key = 'storage' AND deleted_at IS NULL").Rows()
	if err != nil {
		return
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

	endpoint = getSettingStrFromMap(settings, fmt.Sprintf("storage_%s_endpoint", driver))
	useSSL = getSettingBoolFromMap(settings, fmt.Sprintf("storage_%s_use_ssl", driver))
	cdnDomain = getSettingStrFromMap(settings, fmt.Sprintf("storage_%s_cdn_domain", driver))

	switch driver {
	case "oss":
		accessKey = getSettingStrFromMap(settings, "storage_oss_access_key_id")
		secretKey = getSettingStrFromMap(settings, "storage_oss_access_key_secret")
	case "cos":
		accessKey = getSettingStrFromMap(settings, "storage_cos_secret_id")
		secretKey = getSettingStrFromMap(settings, "storage_cos_secret_key")
		region = getSettingStrFromMap(settings, "storage_cos_region")
	default:
		accessKey = getSettingStrFromMap(settings, fmt.Sprintf("storage_%s_access_key", driver))
		secretKey = getSettingStrFromMap(settings, fmt.Sprintf("storage_%s_secret_key", driver))
	}

	return
}

// TestConnection 测试存储桶连接是否可用
func TestConnection(bucketID int64) (string, error) {
	db := database.GetMySQL()
	repo := repository.NewStorageBucketRepo(db)

	bucket, err := repo.GetByID(bucketID)
	if err != nil {
		return "", fmt.Errorf("存储桶不存在: %d", bucketID)
	}

	if bucket.Driver == "local" {
		return "本地存储无需测试连接", nil
	}

	// 获取凭据（桶自身的凭据优先，否则从存储配置获取）
	endpoint := bucket.Endpoint
	accessKey := bucket.AccessKey
	secretKey := bucket.SecretKey
	region := bucket.Region
	useSSL := bucket.UseSSL

	if accessKey == "" {
		ep, ak, sk, rg, _, ssl := GetCredentialsForDriver(bucket.Driver)
		if endpoint == "" {
			endpoint = ep
		}
		accessKey = ak
		secretKey = sk
		region = rg
		useSSL = ssl
	}

	// COS 不需要 endpoint，需要 region
	if bucket.Driver == "cos" {
		if region == "" || accessKey == "" {
			return "", fmt.Errorf("缺少连接信息，请先在存储配置中配置 COS 的 Region 和密钥")
		}
	} else if endpoint == "" || accessKey == "" {
		return "", fmt.Errorf("缺少连接信息，请先在存储配置中配置 %s 的 Endpoint 和密钥", bucket.Driver)
	}

	// 构建配置并创建存储实例
	cfg := config.StorageConfig{
		Driver: bucket.Driver,
		MinIO: config.MinIOConfig{
			Endpoint:  endpoint,
			AccessKey: accessKey,
			SecretKey: secretKey,
			Bucket:    bucket.Bucket,
			UseSSL:    useSSL,
		},
		OSS: config.OSSConfig{
			Endpoint:      endpoint,
			AccessKeyID:   accessKey,
			AccessKeySecret: secretKey,
			Bucket:        bucket.Bucket,
			CDNDomain:     bucket.CDNDomain,
		},
		COS: config.COSConfig{
			Region:    region,
			SecretID:  accessKey,
			SecretKey: secretKey,
			Bucket:    bucket.Bucket,
			CDNDomain: bucket.CDNDomain,
		},
	}

	s, err := storage.New(cfg)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
	}

	// 尝试上传一个小测试文件
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testKey := fmt.Sprintf(".connection-test/%d", time.Now().UnixNano())
	testContent := strings.NewReader("devkit-connection-test")

	_, err = s.Upload(ctx, testKey, testContent, "text/plain")
	if err != nil {
		return "", fmt.Errorf("上传测试失败: %w", err)
	}

	// 清理测试文件
	_ = s.Delete(ctx, testKey)

	return fmt.Sprintf("连接成功！%s 桶 [%s] 可正常读写", bucket.Driver, bucket.Bucket), nil
}

// TestConnectionByDriver 根据驱动和桶名测试连接（无需先保存）
func TestConnectionByDriver(driver, bucketName, region string) (string, error) {
	if driver == "local" {
		return "本地存储无需测试连接", nil
	}

	if bucketName == "" {
		return "", fmt.Errorf("请输入 Bucket 名称")
	}

	// 从存储配置获取凭据
	endpoint, accessKey, secretKey, cfgRegion, cdnDomain, useSSL := GetCredentialsForDriver(driver)
	if driver == "cos" && region != "" {
		cfgRegion = region
	}

	if driver == "cos" {
		if cfgRegion == "" || accessKey == "" {
			return "", fmt.Errorf("缺少连接信息，请先在存储配置中配置 COS 的 Region 和密钥")
		}
	} else if endpoint == "" || accessKey == "" {
		return "", fmt.Errorf("缺少连接信息，请先在存储配置中配置 %s 的 Endpoint 和密钥", driver)
	}

	cfg := config.StorageConfig{
		Driver: driver,
		MinIO: config.MinIOConfig{
			Endpoint:  endpoint,
			AccessKey: accessKey,
			SecretKey: secretKey,
			Bucket:    bucketName,
			UseSSL:    useSSL,
		},
		OSS: config.OSSConfig{
			Endpoint:        endpoint,
			AccessKeyID:     accessKey,
			AccessKeySecret: secretKey,
			Bucket:          bucketName,
			CDNDomain:       cdnDomain,
		},
		COS: config.COSConfig{
			Region:    cfgRegion,
			SecretID:  accessKey,
			SecretKey: secretKey,
			Bucket:    bucketName,
			CDNDomain: cdnDomain,
		},
	}

	s, err := storage.New(cfg)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testKey := fmt.Sprintf(".connection-test/%d", time.Now().UnixNano())
	testContent := strings.NewReader("devkit-connection-test")

	_, err = s.Upload(ctx, testKey, testContent, "text/plain")
	if err != nil {
		return "", fmt.Errorf("上传测试失败: %w", err)
	}

	_ = s.Delete(ctx, testKey)

	return fmt.Sprintf("连接成功！%s 桶 [%s] 可正常读写", driver, bucketName), nil
}

// GetEnabledDrivers 获取已启用的存储驱动列表
// 从 sys_storage_config 表读取，回退到 sys_system_settings
func GetEnabledDrivers() []map[string]interface{} {
	db := database.GetMySQL()
	if db == nil {
		return nil
	}

	drivers := []map[string]interface{}{
		{"value": "local", "label": "本地存储", "icon": "💻", "enabled": true},
	}

	// 从 sys_storage_config 获取已启用的驱动
	externalDrivers := []struct {
		name  string
		label string
		icon  string
	}{
		{"minio", "MinIO", "📦"},
		{"oss", "阿里云 OSS", "☁️"},
		{"cos", "腾讯云 COS", "🌊"},
	}

	for _, d := range externalDrivers {
		var count int64
		db.Raw("SELECT COUNT(*) FROM sys_storage_config WHERE driver = ? AND status = 1", d.name).Scan(&count)
		enabled := count > 0

		// 如果新表没有数据，回退到旧设置
		if !enabled {
			var settingCount int64
			db.Raw("SELECT COUNT(*) FROM sys_system_settings WHERE group_key = 'storage' AND `key` = ? AND value = 'true'", "storage_"+d.name+"_enabled").Scan(&settingCount)
			enabled = settingCount > 0
		}

		drivers = append(drivers, map[string]interface{}{
			"value":   d.name,
			"label":   d.label,
			"icon":    d.icon,
			"enabled": enabled,
		})
	}

	return drivers
}

// InitDefaultStorageBuckets 初始化默认存储桶
// 从存储配置同步默认桶，并确保本地默认桶存在
func InitDefaultStorageBuckets() error {
	db := database.GetMySQL()

	// 先同步外部存储驱动的默认桶
	if err := SyncDefaultBuckets(); err != nil {
		return fmt.Errorf("同步默认存储桶失败: %w", err)
	}

	// 确保本地默认桶存在
	repo := repository.NewStorageBucketRepo(db)
	localDefault, err := repo.GetDefaultByDriver("local")
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("查询本地默认桶失败: %w", err)
	}
	if localDefault == nil {
		if err := repo.Create(&model.StorageBucket{
			Name:        "本地默认存储",
			Driver:      "local",
			Purpose:     "file",
			IsDefault:   true,
			Status:      1,
			Description: "系统内置的本地文件存储",
		}); err != nil {
			return fmt.Errorf("创建本地默认桶失败: %w", err)
		}
	}

	return nil
}
