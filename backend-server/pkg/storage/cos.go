package storage

import (
	"context"
	"fmt"
	"io"

	"backend-server/config"
)

// COStorage 腾讯云 COS 存储
type COStorage struct {
	cfg config.COSConfig
}

// NewCOStorage 创建 COS 存储实例
func NewCOStorage(cfg config.COSConfig) *COStorage {
	return &COStorage{cfg: cfg}
}

// Upload 上传文件
func (s *COStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error) {
	// TODO: 接入腾讯云 COS SDK
	// client := cos.NewClient(...)
	// _, err := client.Object.Put(ctx, objectKey, reader, nil)
	return s.GetURL(objectKey), nil
}

// Download 下载文件
func (s *COStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	// TODO: 接入腾讯云 COS SDK
	return nil, fmt.Errorf("COS 下载未实现")
}

// Delete 删除文件
func (s *COStorage) Delete(ctx context.Context, objectKey string) error {
	// TODO: 接入腾讯云 COS SDK
	return nil
}

// GetURL 获取文件访问 URL
func (s *COStorage) GetURL(objectKey string) string {
	if s.cfg.CDNDomain != "" {
		return fmt.Sprintf("https://%s/%s", s.cfg.CDNDomain, objectKey)
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", s.cfg.Bucket, s.cfg.Region, objectKey)
}

// GetPresignedURL 获取临时访问 URL
func (s *COStorage) GetPresignedURL(ctx context.Context, objectKey string, expire int64) (string, error) {
	// TODO: 接入腾讯云 COS SDK
	return s.GetURL(objectKey), nil
}
