package service

import (
	"context"
	"encoding/json"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"backend-server/pkg/logger"

	"go.uber.org/zap"
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

// ListEnabled 获取启用的轮播图列表（带缓存）
func (s *BannerService) ListEnabled() ([]BannerResponse, error) {
	// 1. 尝试从 Redis 缓存读取
	cacheKey := "banners:enabled"
	ctx := context.Background()

	cached, err := database.GetRedis().Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var results []BannerResponse
		if json.Unmarshal([]byte(cached), &results) == nil {
			logger.Info("从缓存获取Banner数据", zap.Int("count", len(results)))
			return results, nil
		}
	}

	// 2. 缓存不存在或失效，查询数据库
	banners, err := s.repo.ListEnabled()
	if err != nil {
		logger.Error("查询Banner数据失败", zap.Error(err))
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

	// 3. 写入缓存（5分钟过期）
	if data, err := json.Marshal(results); err == nil {
		if err := database.GetRedis().Set(ctx, cacheKey, data, 5*time.Minute).Err(); err != nil {
			logger.Error("写入Banner缓存失败", zap.Error(err))
		} else {
			logger.Info("Banner数据已缓存", zap.Int("count", len(results)), zap.Duration("ttl", 5*time.Minute))
		}
	}

	return results, nil
}

// Create 创建轮播图（清理缓存）
func (s *BannerService) Create(banner *model.Banner) error {
	err := s.repo.Create(banner)
	if err != nil {
		return err
	}
	s.clearCache()
	return nil
}

// Update 更新轮播图（清理缓存）
func (s *BannerService) Update(banner *model.Banner) error {
	err := s.repo.Update(banner)
	if err != nil {
		return err
	}
	s.clearCache()
	return nil
}

// Delete 删除轮播图（清理缓存）
func (s *BannerService) Delete(id uint) error {
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	s.clearCache()
	return nil
}

// UpdateStatus 更新状态（清理缓存）
func (s *BannerService) UpdateStatus(id uint, status string) error {
	err := s.repo.UpdateStatus(id, status)
	if err != nil {
		return err
	}
	s.clearCache()
	return nil
}

// UpdateSortOrder 更新排序（清理缓存）
func (s *BannerService) UpdateSortOrder(id uint, sortOrder int) error {
	err := s.repo.UpdateSortOrder(id, sortOrder)
	if err != nil {
		return err
	}
	s.clearCache()
	return nil
}

// clearCache 清理Banner缓存
func (s *BannerService) clearCache() {
	ctx := context.Background()
	cacheKey := "banners:enabled"
	if err := database.GetRedis().Del(ctx, cacheKey).Err(); err != nil {
		logger.Error("清理Banner缓存失败", zap.Error(err))
	} else {
		logger.Info("Banner缓存已清理")
	}
}

// ListAll 获取所有轮播图（管理用）
func (s *BannerService) ListAll() ([]model.Banner, error) {
	return s.repo.ListAll()
}
