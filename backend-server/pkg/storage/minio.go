package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"backend-server/config"
)

// MinIOStorage MinIO 对象存储
type MinIOStorage struct {
	client *minio.Client
	bucket string
}

// NewMinIOStorage 创建 MinIO 存储实例
func NewMinIOStorage(cfg config.MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("连接 MinIO 失败: %w", err)
	}

	return &MinIOStorage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload 上传文件
func (s *MinIOStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("上传文件到 MinIO 失败: %w", err)
	}

	return s.GetURL(objectKey), nil
}

// Download 下载文件
func (s *MinIOStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("从 MinIO 下载文件失败: %w", err)
	}
	return object, nil
}

// Delete 删除文件
func (s *MinIOStorage) Delete(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
}

// GetURL 获取文件访问 URL
func (s *MinIOStorage) GetURL(objectKey string) string {
	return fmt.Sprintf("/%s/%s", s.bucket, objectKey)
}

// GetPresignedURL 获取临时访问 URL
func (s *MinIOStorage) GetPresignedURL(ctx context.Context, objectKey string, expire int64) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, time.Duration(expire)*time.Second, nil)
	if err != nil {
		return "", fmt.Errorf("生成 MinIO 临时 URL 失败: %w", err)
	}
	return url.String(), nil
}
