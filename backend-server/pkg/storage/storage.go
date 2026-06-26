package storage

import (
	"context"
	"fmt"
	"io"

	"backend-server/config"
)

// Storage 文件存储接口
type Storage interface {
	// Upload 上传文件，返回文件访问 URL
	Upload(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error)

	// Download 下载文件，返回文件内容读取器
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)

	// DownloadRange 下载文件的指定范围（用于 Range 请求）
	DownloadRange(ctx context.Context, objectKey string, offset int64, length int64) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(ctx context.Context, objectKey string) error

	// Exists 检查文件是否存在
	Exists(ctx context.Context, objectKey string) (bool, error)

	// GetURL 获取文件访问 URL
	GetURL(objectKey string) string

	// GetPresignedURL 获取临时访问 URL（用于私有文件）
	GetPresignedURL(ctx context.Context, objectKey string, expire int64) (string, error)
}

// MultipartUploader 分片上传接口
type MultipartUploader interface {
	// InitiateUpload 初始化分片上传，返回 uploadID
	InitiateUpload(ctx context.Context, objectKey string, contentType string) (string, error)

	// UploadPart 上传单个分片，返回 ETag
	UploadPart(ctx context.Context, objectKey string, uploadID string, partNumber int, reader io.Reader, size int64) (string, error)

	// CompleteUpload 合并所有分片，返回文件访问 URL
	CompleteUpload(ctx context.Context, objectKey string, uploadID string, parts []CompletedPart) (string, error)

	// AbortUpload 取消上传，清理临时分片
	AbortUpload(ctx context.Context, objectKey string, uploadID string) error
}

// CompletedPart 已完成的分片信息
type CompletedPart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// UploadedPart 已上传的分片信息（用于列举已上传分片）
type UploadedPart struct {
	PartNumber int   `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
}

// FileInfo 文件信息
type FileInfo struct {
	ObjectKey   string `json:"object_key"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}

// New 根据配置创建存储实例
func New(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Driver {
	case "minio":
		s, err := NewMinIOStorage(cfg.MinIO)
		if err != nil {
			return nil, fmt.Errorf("MinIO 初始化失败: %w", err)
		}
		return s, nil
	case "ceph":
		s, err := NewCephStorage(cfg.MinIO)
		if err != nil {
			return nil, fmt.Errorf("Ceph RGW 初始化失败: %w", err)
		}
		return s, nil
	case "oss":
		s, err := NewOSSStorage(cfg.OSS)
		if err != nil {
			return nil, fmt.Errorf("OSS 初始化失败: %w", err)
		}
		return s, nil
	case "cos":
		s, err := NewCOStorage(cfg.COS)
		if err != nil {
			return nil, fmt.Errorf("COS 初始化失败: %w", err)
		}
		return s, nil
	default:
		return NewLocalStorage(cfg.Local), nil
	}
}
