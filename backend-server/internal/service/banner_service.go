package service

import (
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

// BannerService 轮播图服务
type BannerService struct {
	repo *repository.BannerRepo
}

// NewBannerService 创建轮播图服务
func NewBannerService() *BannerService {
	return &BannerService{
		repo: repository.NewBannerRepo(database.GetMySQL()),
	}
}

// BannerResponse 轮播图响应
type BannerResponse struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Image    string `json:"image"`
	Link     string `json:"link"`
	LinkType string `json:"linkType"`
}

// ListEnabled 获取启用的轮播图列表
func (s *BannerService) ListEnabled() ([]BannerResponse, error) {
	banners, err := s.repo.ListEnabled()
	if err != nil {
		return nil, err
	}

	results := make([]BannerResponse, 0, len(banners))
	for _, b := range banners {
		results = append(results, BannerResponse{
			ID:       b.ID,
			Title:    b.Title,
			Image:    b.Image,
			Link:     b.Link,
			LinkType: b.LinkType,
		})
	}

	return results, nil
}

// Create 创建轮播图
func (s *BannerService) Create(banner *model.Banner) error {
	return s.repo.Create(banner)
}

// Update 更新轮播图
func (s *BannerService) Update(banner *model.Banner) error {
	return s.repo.Update(banner)
}

// Delete 删除轮播图
func (s *BannerService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// UpdateStatus 更新状态
func (s *BannerService) UpdateStatus(id uint, status string) error {
	return s.repo.UpdateStatus(id, status)
}

// UpdateSortOrder 更新排序
func (s *BannerService) UpdateSortOrder(id uint, sortOrder int) error {
	return s.repo.UpdateSortOrder(id, sortOrder)
}

// ListAll 获取所有轮播图（管理用）
func (s *BannerService) ListAll() ([]model.Banner, error) {
	return s.repo.ListAll()
}
