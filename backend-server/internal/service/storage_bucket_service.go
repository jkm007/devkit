package service

import (
	"claude-manager/internal/model"
	"claude-manager/internal/repository"
	"fmt"
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
func (s *StorageBucketService) Create(bucket *model.StorageBucket) error {
	// 检查名称是否存在
	exists, err := s.repo.NameExists(bucket.Name, 0)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("存储桶名称 %s 已存在", bucket.Name)
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

	return s.repo.SetDefault(id)
}
