package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// BannerRepo 轮播图数据访问
type BannerRepo struct {
	db *gorm.DB
}

// NewBannerRepo 创建轮播图仓库
func NewBannerRepo(db *gorm.DB) *BannerRepo {
	return &BannerRepo{db: db}
}

// ListEnabled 获取所有启用的轮播图（按排序顺序）
func (r *BannerRepo) ListEnabled() ([]model.Banner, error) {
	var banners []model.Banner
	err := r.db.Where("status = ?", model.BannerEnabled).
		Order("sort_order ASC, id ASC").
		Find(&banners).Error
	return banners, err
}

// Create 创建轮播图
func (r *BannerRepo) Create(banner *model.Banner) error {
	return r.db.Create(banner).Error
}

// Update 更新轮播图
func (r *BannerRepo) Update(banner *model.Banner) error {
	return r.db.Model(banner).Updates(banner).Error
}

// Delete 删除轮播图
func (r *BannerRepo) Delete(id uint) error {
	return r.db.Delete(&model.Banner{}, id).Error
}

// UpdateStatus 更新状态
func (r *BannerRepo) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Banner{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateSortOrder 更新排序
func (r *BannerRepo) UpdateSortOrder(id uint, sortOrder int) error {
	return r.db.Model(&model.Banner{}).Where("id = ?", id).Update("sort_order", sortOrder).Error
}

// ListAll 获取所有轮播图（管理用）
func (r *BannerRepo) ListAll() ([]model.Banner, error) {
	var banners []model.Banner
	err := r.db.Order("sort_order ASC, id ASC").Find(&banners).Error
	return banners, err
}

// GetByID 根据ID获取轮播图
func (r *BannerRepo) GetByID(id uint) (*model.Banner, error) {
	var banner model.Banner
	err := r.db.First(&banner, id).Error
	if err != nil {
		return nil, err
	}
	return &banner, nil
}

// CountByFileID 统计引用指定文件的轮播图数量
func (r *BannerRepo) CountByFileID(fileID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Banner{}).Where("file_id = ?", fileID).Count(&count).Error
	return count, err
}
