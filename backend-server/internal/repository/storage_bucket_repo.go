package repository

import (
	"backend-server/internal/model"
	"gorm.io/gorm"
)

// StorageBucketRepo 存储桶仓库
type StorageBucketRepo struct {
	db *gorm.DB
}

// NewStorageBucketRepo 创建存储桶仓库
func NewStorageBucketRepo(db *gorm.DB) *StorageBucketRepo {
	return &StorageBucketRepo{db: db}
}

// GetAll 获取所有存储桶
func (r *StorageBucketRepo) GetAll() ([]model.StorageBucket, error) {
	var buckets []model.StorageBucket
	err := r.db.Order("is_default DESC, id ASC").Find(&buckets).Error
	return buckets, err
}

// GetByID 根据ID获取存储桶
func (r *StorageBucketRepo) GetByID(id int64) (*model.StorageBucket, error) {
	var bucket model.StorageBucket
	err := r.db.Where("id = ?", id).First(&bucket).Error
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

// GetByName 根据名称获取存储桶
func (r *StorageBucketRepo) GetByName(name string) (*model.StorageBucket, error) {
	var bucket model.StorageBucket
	err := r.db.Where("name = ?", name).First(&bucket).Error
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

// GetByDriver 根据驱动获取存储桶列表
func (r *StorageBucketRepo) GetByDriver(driver string) ([]model.StorageBucket, error) {
	var buckets []model.StorageBucket
	err := r.db.Where("driver = ? AND status = ?", driver, 1).Find(&buckets).Error
	return buckets, err
}

// GetByPurpose 根据用途获取存储桶列表
func (r *StorageBucketRepo) GetByPurpose(purpose string) ([]model.StorageBucket, error) {
	var buckets []model.StorageBucket
	err := r.db.Where("purpose = ? AND status = ?", purpose, 1).Find(&buckets).Error
	return buckets, err
}

// GetDefault 获取默认存储桶
func (r *StorageBucketRepo) GetDefault() (*model.StorageBucket, error) {
	var bucket model.StorageBucket
	err := r.db.Where("is_default = ? AND status = ?", true, 1).First(&bucket).Error
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

// Create 创建存储桶
func (r *StorageBucketRepo) Create(bucket *model.StorageBucket) error {
	return r.db.Create(bucket).Error
}

// Update 更新存储桶
func (r *StorageBucketRepo) Update(bucket *model.StorageBucket) error {
	return r.db.Save(bucket).Error
}

// Delete 删除存储桶
func (r *StorageBucketRepo) Delete(id int64) error {
	return r.db.Delete(&model.StorageBucket{}, id).Error
}

// SetDefault 设置默认存储桶（取消其他默认）
func (r *StorageBucketRepo) SetDefault(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 取消所有默认
		if err := tx.Model(&model.StorageBucket{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		// 设置新的默认
		return tx.Model(&model.StorageBucket{}).Where("id = ?", id).Update("is_default", true).Error
	})
}

// NameExists 检查名称是否存在
func (r *StorageBucketRepo) NameExists(name string, excludeID int64) (bool, error) {
	var count int64
	query := r.db.Model(&model.StorageBucket{}).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
