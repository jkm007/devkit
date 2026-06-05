package cdn

import (
	"fmt"

	"backend-server/config"
)

// CDN CDN 工具
type CDN struct {
	cfg config.CDNConfig
}

// New 创建 CDN 实例
func New(cfg config.CDNConfig) *CDN {
	return &CDN{cfg: cfg}
}

// GetURL 获取 CDN URL
func (c *CDN) GetURL(objectKey string) string {
	return fmt.Sprintf("https://%s/%s", c.cfg.Domain, objectKey)
}

// GetImageURL 获取图片 URL（带处理参数）
func (c *CDN) GetImageURL(objectKey string, width, height int) string {
	baseURL := c.GetURL(objectKey)
	if width > 0 && height > 0 {
		return fmt.Sprintf("%s%sresize,w_%d,h_%d", baseURL, c.cfg.ImageStyleSeparator, width, height)
	}
	if width > 0 {
		return fmt.Sprintf("%s%sw_%d", baseURL, c.cfg.ImageStyleSeparator, width)
	}
	if height > 0 {
		return fmt.Sprintf("%s%sh_%d", baseURL, c.cfg.ImageStyleSeparator, height)
	}
	return baseURL
}

// GetThumbnailURL 获取缩略图 URL
func (c *CDN) GetThumbnailURL(objectKey string, size int) string {
	return c.GetImageURL(objectKey, size, size)
}

// GetCropURL 获取裁剪后的图片 URL
func (c *CDN) GetCropURL(objectKey string, width, height int) string {
	baseURL := c.GetURL(objectKey)
	return fmt.Sprintf("%s%scrop,w_%d,h_%d", baseURL, c.cfg.ImageStyleSeparator, width, height)
}
