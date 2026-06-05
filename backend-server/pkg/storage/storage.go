package storage

import (
	"context"
	"io"
	"log"

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
		s, err := NewMinIOStorage(cfg.MinIO)
		if err != nil {
			log.Printf("[WARN] MinIO 初始化失败，回退到本地存储: %v", err)
			return NewLocalStorage(cfg.Local)
		}
		return s
	case "oss":
		log.Printf("[WARN] OSS 存储驱动尚未实现，回退到本地存储")
		return NewLocalStorage(cfg.Local)
	case "cos":
		log.Printf("[WARN] COS 存储驱动尚未实现，回退到本地存储")
		return NewLocalStorage(cfg.Local)
	default:
		return NewLocalStorage(cfg.Local)
	}
}
