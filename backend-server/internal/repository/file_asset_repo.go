package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// FileAssetRepo 文件资产仓库
type FileAssetRepo struct {
	db *gorm.DB
}

func NewFileAssetRepo(db *gorm.DB) *FileAssetRepo {
	return &FileAssetRepo{db: db}
}

// Create 创建文件资产
func (r *FileAssetRepo) Create(asset *model.FileAsset) error {
	return r.db.Create(asset).Error
}

// GetByHash 根据哈希查找文件资产（秒传）
func (r *FileAssetRepo) GetByHash(fileHash string) (*model.FileAsset, error) {
	var asset model.FileAsset
	if err := r.db.Where("file_hash = ?", fileHash).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

// IncrementRefCount 增加引用计数
func (r *FileAssetRepo) IncrementRefCount(id uint) error {
	return r.db.Model(&model.FileAsset{}).Where("id = ?", id).Update("ref_count", gorm.Expr("ref_count + 1")).Error
}

// DecrementRefCount 减少引用计数
func (r *FileAssetRepo) DecrementRefCount(id uint) error {
	return r.db.Model(&model.FileAsset{}).Where("id = ?", id).Update("ref_count", gorm.Expr("ref_count - 1")).Error
}

// GetByID 根据 ID 获取文件资产
func (r *FileAssetRepo) GetByID(id uint) (*model.FileAsset, error) {
	var asset model.FileAsset
	if err := r.db.Where("id = ?", id).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}
