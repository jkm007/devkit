package service

import (
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

// MediaService 媒体管理服务
type MediaService struct {
	mediaRepo *repository.MediaRepo
}

func NewMediaService() *MediaService {
	db := database.GetMySQL()
	return &MediaService{
		mediaRepo: repository.NewMediaRepo(db),
	}
}

// GetMediaInfo 获取媒体元数据
func (s *MediaService) GetMediaInfo(fileAssetID uint) (*model.MediaAsset, error) {
	return s.mediaRepo.GetByFileAssetID(fileAssetID)
}

// CreateMediaRecord 创建媒体记录（上传完成后由 Worker 调用）
func (s *MediaService) CreateMediaRecord(media *model.MediaAsset) error {
	return s.mediaRepo.Create(media)
}

// UpdateTranscodeStatus 更新转码状态
func (s *MediaService) UpdateTranscodeStatus(fileAssetID uint, status string, hlsPath string) error {
	return s.mediaRepo.UpdateTranscodeStatus(fileAssetID, status, hlsPath)
}

// GetPendingTranscode 获取待转码的媒体
func (s *MediaService) GetPendingTranscode(limit int) ([]model.MediaAsset, error) {
	return s.mediaRepo.GetPendingTranscode(limit)
}
