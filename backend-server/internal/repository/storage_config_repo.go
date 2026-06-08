package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// StorageConfigRepo 存储连接配置仓库
type StorageConfigRepo struct {
	db *gorm.DB
}

func NewStorageConfigRepo(db *gorm.DB) *StorageConfigRepo {
	return &StorageConfigRepo{db: db}
}

// GetAll 获取所有存储配置
func (r *StorageConfigRepo) GetAll() ([]model.StorageConfig, error) {
	var configs []model.StorageConfig
	err := r.db.Order("driver ASC, is_default DESC, id ASC").Find(&configs).Error
	return configs, err
}

// GetByID 根据ID获取
func (r *StorageConfigRepo) GetByID(id int64) (*model.StorageConfig, error) {
	var config model.StorageConfig
	err := r.db.Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByDriver 根据驱动获取
func (r *StorageConfigRepo) GetByDriver(driver string) ([]model.StorageConfig, error) {
	var configs []model.StorageConfig
	err := r.db.Where("driver = ? AND status = 1", driver).Order("is_default DESC, id ASC").Find(&configs).Error
	return configs, err
}

// GetDefaultByDriver 获取指定驱动的默认配置
func (r *StorageConfigRepo) GetDefaultByDriver(driver string) (*model.StorageConfig, error) {
	var config model.StorageConfig
	err := r.db.Where("driver = ? AND is_default = 1 AND status = 1", driver).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetEnabled 获取所有启用的配置
func (r *StorageConfigRepo) GetEnabled() ([]model.StorageConfig, error) {
	var configs []model.StorageConfig
	err := r.db.Where("status = 1").Order("driver ASC, is_default DESC, id ASC").Find(&configs).Error
	return configs, err
}

// Create 创建
func (r *StorageConfigRepo) Create(config *model.StorageConfig) error {
	return r.db.Create(config).Error
}

// Update 更新
func (r *StorageConfigRepo) Update(config *model.StorageConfig) error {
	return r.db.Save(config).Error
}

// Delete 删除
func (r *StorageConfigRepo) Delete(id int64) error {
	return r.db.Delete(&model.StorageConfig{}, id).Error
}

// ClearDefault 清除指定驱动的默认标记
func (r *StorageConfigRepo) ClearDefault(driver string) error {
	return r.db.Model(&model.StorageConfig{}).
		Where("driver = ? AND is_default = 1", driver).
		Update("is_default", false).Error
}

// SetDefault 设置默认
func (r *StorageConfigRepo) SetDefault(id int64, driver string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 清除该驱动的其他默认
		if err := tx.Model(&model.StorageConfig{}).
			Where("driver = ? AND is_default = 1", driver).
			Update("is_default", false).Error; err != nil {
			return err
		}
		// 设置新的默认
		return tx.Model(&model.StorageConfig{}).
			Where("id = ?", id).
			Update("is_default", true).Error
	})
}

// GetEnabledDrivers 获取已启用的驱动列表（去重）
func (r *StorageConfigRepo) GetEnabledDrivers() ([]string, error) {
	var drivers []string
	err := r.db.Model(&model.StorageConfig{}).
		Where("status = 1").
		Distinct("driver").
		Pluck("driver", &drivers).Error
	return drivers, err
}
