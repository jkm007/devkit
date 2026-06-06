package storage

import (
	"context"
	"fmt"
	"io"

	"backend-server/config"
)

// OSSStorage 阿里云 OSS 存储
type OSSStorage struct {
	cfg config.OSSConfig
}

// NewOSSStorage 创建 OSS 存储实例
func NewOSSStorage(cfg config.OSSConfig) *OSSStorage {
	return &OSSStorage{cfg: cfg}
}

// Upload 上传文件
func (s *OSSStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error) {
	// TODO: 接入阿里云 OSS SDK
	return "", fmt.Errorf("OSS 存储驱动尚未实现，请切换到 local 或 minio")
}

// Download 下载文件
func (s *OSSStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	// TODO: 接入阿里云 OSS SDK
	return nil, fmt.Errorf("OSS 下载未实现")
}

// Delete 删除文件
func (s *OSSStorage) Delete(ctx context.Context, objectKey string) error {
	return fmt.Errorf("OSS 存储驱动尚未实现")
}

// GetURL 获取文件访问 URL
func (s *OSSStorage) GetURL(objectKey string) string {
	if s.cfg.CDNDomain != "" {
		return fmt.Sprintf("https://%s/%s", s.cfg.CDNDomain, objectKey)
	}
	return fmt.Sprintf("https://%s.%s/%s", s.cfg.Bucket, s.cfg.Endpoint, objectKey)
}

// GetPresignedURL 获取临时访问 URL
func (s *OSSStorage) GetPresignedURL(ctx context.Context, objectKey string, expire int64) (string, error) {
	return "", fmt.Errorf("OSS 存储驱动尚未实现")
}
