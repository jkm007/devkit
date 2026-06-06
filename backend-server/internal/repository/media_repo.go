package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// MediaRepo 媒体资产仓库
type MediaRepo struct {
	db *gorm.DB
}

func NewMediaRepo(db *gorm.DB) *MediaRepo {
	return &MediaRepo{db: db}
}

// Create 创建媒体资产
func (r *MediaRepo) Create(media *model.MediaAsset) error {
	return r.db.Create(media).Error
}

// GetByFileAssetID 根据文件资产ID获取媒体信息
func (r *MediaRepo) GetByFileAssetID(fileAssetID uint) (*model.MediaAsset, error) {
	var media model.MediaAsset
	if err := r.db.Where("file_asset_id = ?", fileAssetID).First(&media).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

// Update 更新媒体资产
func (r *MediaRepo) Update(media *model.MediaAsset) error {
	return r.db.Save(media).Error
}

// UpdateTranscodeStatus 更新转码状态
func (r *MediaRepo) UpdateTranscodeStatus(fileAssetID uint, status string, hlsPath string) error {
	return r.db.Model(&model.MediaAsset{}).
		Where("file_asset_id = ?", fileAssetID).
		Updates(map[string]interface{}{
			"transcode_status": status,
			"hls_path":         hlsPath,
		}).Error
}

// GetPendingTranscode 获取待转码的媒体
func (r *MediaRepo) GetPendingTranscode(limit int) ([]model.MediaAsset, error) {
	var medias []model.MediaAsset
	if err := r.db.Where("transcode_status = ?", "pending").Limit(limit).Find(&medias).Error; err != nil {
		return nil, err
	}
	return medias, nil
}
