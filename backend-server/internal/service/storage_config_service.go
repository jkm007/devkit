package service

import (
	"fmt"
	"log"

	"backend-server/config"
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"backend-server/pkg/storage"
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

	// 如果设为默认，清除其他默认
	if config.IsDefault {
		if err := s.repo.ClearDefault(config.Driver); err != nil {
			log.Printf("[WARN] 清除默认标记失败: %v", err)
		}
		// 重新设置自己为默认
		s.repo.SetDefault(config.ID, config.Driver)
	}

	// 刷新存储管理器
	storage.RefreshStorage()
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

	// 敏感字段处理：如果传入的是 ******，保留原值
	if config.AccessKey == "******" {
		config.AccessKey = existing.AccessKey
	}
	if config.SecretKey == "******" {
		config.SecretKey = existing.SecretKey
	}

	if err := s.repo.Update(config); err != nil {
		return fmt.Errorf("更新存储配置失败: %w", err)
	}

	// 如果设为默认，清除其他默认
	if config.IsDefault {
		if err := s.repo.ClearDefault(config.Driver); err != nil {
			log.Printf("[WARN] 清除默认标记失败: %v", err)
		}
		s.repo.SetDefault(config.ID, config.Driver)
	}

	// 刷新存储管理器
	storage.RefreshStorage()
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
	storage.RefreshStorage()
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
	storage.RefreshStorage()
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

// testByConfig 根据配置测试连接
func (s *StorageConfigService) testByConfig(config *model.StorageConfig) error {
	cfg := configModelToConfig(config)
	st, err := storage.New(cfg)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	// 尝试获取 URL 来验证连接
	_ = st.GetURL("test-connection-check")
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

// GetEnabledDrivers 获取已启用的驱动列表
func (s *StorageConfigService) GetEnabledDrivers() ([]map[string]interface{}, error) {
	configs, err := s.repo.GetEnabled()
	if err != nil {
		return nil, err
	}

	// 去重并构建结果
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
	for driver := range driverLabels {
		result = append(result, map[string]interface{}{
			"value":   driver,
			"label":   driverLabels[driver],
			"icon":    driverIcons[driver],
			"enabled": driverMap[driver] || driver == "local",
		})
	}
	return result, nil
}
