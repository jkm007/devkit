package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"backend-server/config"
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"backend-server/pkg/storage"

	"gorm.io/gorm"
)

// StorageConfigService 存储连接配置服务
type StorageConfigService struct {
	repo *repository.StorageConfigRepo
}

func NewStorageConfigService() *StorageConfigService {
	db := database.GetMySQL()
	return &StorageConfigService{
		repo: repository.NewStorageConfigRepo(db),
	}
}

// GetAll 获取所有配置
func (s *StorageConfigService) GetAll() ([]model.StorageConfig, error) {
	return s.repo.GetAll()
}

// GetByID 根据ID获取
func (s *StorageConfigService) GetByID(id int64) (*model.StorageConfig, error) {
	return s.repo.GetByID(id)
}

// Create 创建配置
func (s *StorageConfigService) Create(config *model.StorageConfig) error {
	// 本地存储只能有一个
	if config.Driver == "local" {
		configs, err := s.repo.GetByDriver("local")
		if err == nil && len(configs) > 0 {
			return fmt.Errorf("本地存储配置已存在，不能重复创建")
		}
	}

	if err := s.repo.Create(config); err != nil {
		return fmt.Errorf("创建存储配置失败: %w", err)
	}

	// 如果设为默认，使用事务方法设置（repo.SetDefault 内部已包含清除旧默认+设置新默认）
	if config.IsDefault {
		if err := s.repo.SetDefault(config.ID, config.Driver); err != nil {
			return fmt.Errorf("设置默认失败: %w", err)
		}
	}

	// 刷新存储管理器
	if err := storage.RefreshStorage(); err != nil {
		log.Printf("[WARN] 刷新存储管理器失败: %v", err)
	}

	// 同步到桶管理
	if err := s.SyncBucketsFromConfig(); err != nil {
		log.Printf("[WARN] 同步存储桶失败: %v", err)
	}

	return nil
}

// Update 更新配置
func (s *StorageConfigService) Update(config *model.StorageConfig) error {
	existing, err := s.repo.GetByID(config.ID)
	if err != nil {
		return fmt.Errorf("存储配置不存在")
	}

	// 本地存储不允许修改驱动
	if existing.Driver == "local" && config.Driver != "local" {
		return fmt.Errorf("本地存储不允许修改驱动类型")
	}

	// 敏感字段处理：如果传入的是 ****** 或空字符串，保留原值
	if config.AccessKey == "******" || config.AccessKey == "" {
		config.AccessKey = existing.AccessKey
	}
	if config.SecretKey == "******" || config.SecretKey == "" {
		config.SecretKey = existing.SecretKey
	}

	if err := s.repo.Update(config); err != nil {
		return fmt.Errorf("更新存储配置失败: %w", err)
	}

	// 如果设为默认，使用事务方法设置
	if config.IsDefault {
		if err := s.repo.SetDefault(config.ID, config.Driver); err != nil {
			return fmt.Errorf("设置默认失败: %w", err)
		}
	}

	// 刷新存储管理器
	if err := storage.RefreshStorage(); err != nil {
		log.Printf("[WARN] 刷新存储管理器失败: %v", err)
	}

	// 同步到桶管理
	if err := s.SyncBucketsFromConfig(); err != nil {
		log.Printf("[WARN] 同步存储桶失败: %v", err)
	}

	return nil
}

// Delete 删除配置
func (s *StorageConfigService) Delete(id int64) error {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("存储配置不存在")
	}

	// 本地存储不允许删除
	if existing.Driver == "local" {
		return fmt.Errorf("本地存储配置不允许删除")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("删除存储配置失败: %w", err)
	}

	// 刷新存储管理器
	if err := storage.RefreshStorage(); err != nil {
		log.Printf("[WARN] 刷新存储管理器失败: %v", err)
	}

	// 同步到桶管理
	if err := s.SyncBucketsFromConfig(); err != nil {
		log.Printf("[WARN] 同步存储桶失败: %v", err)
	}

	return nil
}

// SetDefault 设置默认
func (s *StorageConfigService) SetDefault(id int64) error {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("存储配置不存在")
	}

	if config.Status != 1 {
		return fmt.Errorf("禁用的配置不能设为默认")
	}

	if err := s.repo.SetDefault(id, config.Driver); err != nil {
		return fmt.Errorf("设置默认失败: %w", err)
	}

	// 刷新存储管理器
	if err := storage.RefreshStorage(); err != nil {
		log.Printf("[WARN] 刷新存储管理器失败: %v", err)
	}
	return nil
}

// TestConnection 测试连接
func (s *StorageConfigService) TestConnection(id int64) error {
	config, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("存储配置不存在")
	}
	return s.testByConfig(config)
}

// TestConnectionByData 根据传入数据测试连接（用于创建前测试）
func (s *StorageConfigService) TestConnectionByData(driver, endpoint, accessKey, secretKey, bucket, region string, useSSL bool) error {
	config := &model.StorageConfig{
		Driver:    driver,
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
		Region:    region,
		UseSSL:    useSSL,
	}
	return s.testByConfig(config)
}

// testByConfig 根据配置测试实际连接（上传+删除测试文件）
func (s *StorageConfigService) testByConfig(config *model.StorageConfig) error {
	if config.Driver == "local" {
		return nil
	}

	if config.Bucket == "" {
		return fmt.Errorf("Bucket 名称不能为空")
	}

	cfg := configModelToConfig(config)
	st, err := storage.New(cfg)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	// 实际测试：上传一个小文件再删除
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testKey := fmt.Sprintf(".connection-test/%d", time.Now().UnixNano())
	testContent := strings.NewReader("devkit-connection-test")

	_, err = st.Upload(ctx, testKey, testContent, "text/plain")
	if err != nil {
		return fmt.Errorf("上传测试失败: %w", err)
	}

	// 清理测试文件
	_ = st.Delete(ctx, testKey)
	return nil
}

// configModelToConfig 将 model.StorageConfig 转换为 config.StorageConfig
func configModelToConfig(m *model.StorageConfig) config.StorageConfig {
	cfg := config.StorageConfig{Driver: m.Driver}
	switch m.Driver {
	case "local":
		cfg.Local.Path = "./uploads"
		cfg.Local.URLPrefix = "/uploads"
	case "minio":
		cfg.MinIO.Endpoint = m.Endpoint
		cfg.MinIO.AccessKey = m.AccessKey
		cfg.MinIO.SecretKey = m.SecretKey
		cfg.MinIO.Bucket = m.Bucket
		cfg.MinIO.UseSSL = m.UseSSL
	case "oss":
		cfg.OSS.Endpoint = m.Endpoint
		cfg.OSS.AccessKeyID = m.AccessKey
		cfg.OSS.AccessKeySecret = m.SecretKey
		cfg.OSS.Bucket = m.Bucket
		cfg.OSS.CDNDomain = m.CDNDomain
	case "cos":
		cfg.COS.Region = m.Region
		cfg.COS.SecretID = m.AccessKey
		cfg.COS.SecretKey = m.SecretKey
		cfg.COS.Bucket = m.Bucket
		cfg.COS.CDNDomain = m.CDNDomain
	}
	return cfg
}

// SyncBucketsFromConfig 从 sys_storage_config 同步到 sys_storage_bucket
// 当存储配置创建/更新/删除时调用，确保桶管理中有对应的默认桶
func (s *StorageConfigService) SyncBucketsFromConfig() error {
	db := database.GetMySQL()
	bucketRepo := repository.NewStorageBucketRepo(db)

	configs, err := s.repo.GetAll()
	if err != nil {
		return fmt.Errorf("读取存储配置失败: %w", err)
	}

	// 按 driver 分组，取默认配置
	driverConfigs := make(map[string]*model.StorageConfig)
	for i, c := range configs {
		if c.Status != 1 {
			continue
		}
		if _, ok := driverConfigs[c.Driver]; !ok || c.IsDefault {
			driverConfigs[c.Driver] = &configs[i]
		}
	}

	driverLabels := map[string]string{
		"local": "本地存储",
		"minio": "MinIO",
		"oss":   "阿里云 OSS",
		"cos":   "腾讯云 COS",
	}

	// 同步每个驱动
	for driver, cfg := range driverConfigs {
		label := driverLabels[driver]
		if label == "" {
			label = driver
		}

		// 查找该驱动的默认桶
		existing, err := bucketRepo.GetDefaultByDriver(driver)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Printf("[WARN] 查询 %s 默认桶失败: %v", driver, err)
			continue
		}

		if existing != nil {
			// 更新凭据
			existing.Endpoint = cfg.Endpoint
			existing.AccessKey = cfg.AccessKey
			existing.SecretKey = cfg.SecretKey
			existing.CDNDomain = cfg.CDNDomain
			existing.UseSSL = cfg.UseSSL
			if driver == "cos" {
				existing.Region = cfg.Region
			}
			existing.Status = 1
			existing.Description = fmt.Sprintf("由存储配置自动同步的 %s 桶", label)
			if err := bucketRepo.Update(existing); err != nil {
				log.Printf("[ERROR] 更新 %s 默认桶失败: %v", driver, err)
			} else {
				log.Printf("[INFO] 已更新 %s 默认桶凭据", driver)
			}
		} else {
			// 创建新桶
			newBucket := &model.StorageBucket{
				Name:       fmt.Sprintf("%s 默认桶", label),
				Driver:     driver,
				Endpoint:   cfg.Endpoint,
				Bucket:     cfg.Bucket,
				AccessKey:  cfg.AccessKey,
				SecretKey:  cfg.SecretKey,
				CDNDomain:  cfg.CDNDomain,
				UseSSL:     cfg.UseSSL,
				Purpose:    "file",
				IsDefault:  true,
				Status:     1,
				Description: fmt.Sprintf("由存储配置自动同步的 %s 桶", label),
			}
			if driver == "cos" {
				newBucket.Region = cfg.Region
			}
			if err := bucketRepo.Create(newBucket); err != nil {
				log.Printf("[ERROR] 创建 %s 默认桶失败: %v", driver, err)
			} else {
				log.Printf("[INFO] 已创建 %s 默认桶: %s", driver, newBucket.Name)
			}
		}
	}

	// 禁用没有配置的驱动的桶
	allDrivers := []string{"minio", "oss", "cos"}
	for _, driver := range allDrivers {
		if _, ok := driverConfigs[driver]; !ok {
			existing, err := bucketRepo.GetDefaultByDriver(driver)
			if err == nil && existing != nil && existing.Status == 1 {
				existing.Status = 0
				existing.Description = fmt.Sprintf("%s 未配置，请在存储配置中添加", driverLabels[driver])
				if err := bucketRepo.Update(existing); err != nil {
					log.Printf("[ERROR] 禁用 %s 默认桶失败: %v", driver, err)
				}
			}
		}
	}

	return nil
}

// driverOrder 驱动固定排序顺序
var driverOrder = []string{"local", "minio", "oss", "cos"}

// GetEnabledDrivers 获取已启用的驱动列表（固定顺序）
func (s *StorageConfigService) GetEnabledDrivers() ([]map[string]interface{}, error) {
	configs, err := s.repo.GetEnabled()
	if err != nil {
		return nil, err
	}

	// 去重
	driverMap := make(map[string]bool)
	for _, c := range configs {
		driverMap[c.Driver] = true
	}

	driverLabels := map[string]string{
		"local": "本地存储",
		"minio": "MinIO",
		"oss":   "阿里云 OSS",
		"cos":   "腾讯云 COS",
	}
	driverIcons := map[string]string{
		"local": "💻",
		"minio": "📦",
		"oss":   "☁️",
		"cos":   "🌊",
	}

	var result []map[string]interface{}
	for _, driver := range driverOrder {
		result = append(result, map[string]interface{}{
			"value":   driver,
			"label":   driverLabels[driver],
			"icon":    driverIcons[driver],
			"enabled": driverMap[driver] || driver == "local",
		})
	}
	return result, nil
}
