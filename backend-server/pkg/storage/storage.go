package storage

import (
	"context"
	"io"

	"backend-server/config"
)

// Storage 文件存储接口
type Storage interface {
	// Upload 上传文件，返回文件访问 URL
	Upload(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error)

	// Download 下载文件，返回文件内容读取器
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(ctx context.Context, objectKey string) error

	// GetURL 获取文件访问 URL
	GetURL(objectKey string) string

	// GetPresignedURL 获取临时访问 URL（用于私有文件）
	GetPresignedURL(ctx context.Context, objectKey string, expire int64) (string, error)
}

// FileInfo 文件信息
type FileInfo struct {
	ObjectKey   string `json:"object_key"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}

// New 根据配置创建存储实例
func New(cfg config.StorageConfig) Storage {
	switch cfg.Driver {
	case "minio":
		return NewMinIOStorage(cfg.MinIO)
	case "oss":
		return NewOSSStorage(cfg.OSS)
	case "cos":
		return NewCOStorage(cfg.COS)
	default:
		return NewLocalStorage(cfg.Local)
	}
}
