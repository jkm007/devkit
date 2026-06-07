package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
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
	// 清理 endpoint 格式：移除 http:// 或 https:// 前缀
	endpoint := cfg.Endpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimSpace(endpoint)

	// 移除可能的 JSON 引号
	endpoint = strings.Trim(endpoint, "\"")

	log.Printf("[DEBUG] MinIO Endpoint 原始值: %q, 清理后: %q", cfg.Endpoint, endpoint)

	if endpoint == "" {
		return nil, fmt.Errorf("MinIO endpoint 不能为空")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:    cfg.UseSSL,
		Transport: nil, // 使用默认 transport，后续可自定义超时
	})
	if err != nil {
		return nil, fmt.Errorf("连接 MinIO 失败 (endpoint=%s, useSSL=%v): %w", endpoint, cfg.UseSSL, err)
	}

	// 检查并创建 bucket
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("检查 bucket 失败: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("创建 bucket %s 失败: %w", cfg.Bucket, err)
		}
		log.Printf("[INFO] MinIO bucket %s 已创建", cfg.Bucket)
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

// chunkPrefix 返回分片临时存储路径前缀
func chunkPrefix(uploadID string) string {
	return fmt.Sprintf("uploads/%s", uploadID)
}

// chunkKey 返回单个分片的存储路径
func chunkKey(uploadID string, partNumber int) string {
	return fmt.Sprintf("uploads/%s/part-%06d", uploadID, partNumber)
}

// InitiateUpload 初始化分片上传（uploadID 由调用方生成并传入）
func (s *MinIOStorage) InitiateUpload(ctx context.Context, objectKey string, contentType string) (string, error) {
	// uploadID 由 service 层生成，这里仅做占位
	// 实际的 MinIO 分片由 UploadPart 逐个上传临时对象
	return "", nil
}

// UploadPart 上传单个分片为临时对象
func (s *MinIOStorage) UploadPart(ctx context.Context, objectKey string, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	key := chunkKey(uploadID, partNumber)
	log.Printf("[DEBUG] MinIO UploadPart: bucket=%s, key=%s, size=%d", s.bucket, key, size)

	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{})
	if err != nil {
		log.Printf("[ERROR] MinIO UploadPart 失败: %v", err)
		return "", fmt.Errorf("上传分片 %d 失败: %w", partNumber, err)
	}
	return key, nil // 返回临时对象路径作为 ETag
}

// CompleteUpload 使用 ComposeObject 合并所有分片到最终路径
func (s *MinIOStorage) CompleteUpload(ctx context.Context, objectKey string, uploadID string, parts []CompletedPart) (string, error) {
	sort.Slice(parts, func(i, j int) bool { return parts[i].PartNumber < parts[j].PartNumber })

	// 构建 ComposeObject 源
	srcs := make([]minio.CopySrcOptions, len(parts))
	for i, p := range parts {
		srcs[i] = minio.CopySrcOptions{
			Bucket: s.bucket,
			Object: p.ETag, // ETag 存的是临时对象路径
		}
	}

	dst := minio.CopyDestOptions{
		Bucket: s.bucket,
		Object: objectKey,
	}

	_, err := s.client.ComposeObject(ctx, dst, srcs...)
	if err != nil {
		return "", fmt.Errorf("合并分片失败: %w", err)
	}

	// 清理临时分片对象
	go func() {
		delCtx := context.Background()
		for _, p := range parts {
			s.client.RemoveObject(delCtx, s.bucket, p.ETag, minio.RemoveObjectOptions{})
		}
	}()

	return s.GetURL(objectKey), nil
}

// AbortUpload 取消上传，清理所有临时分片
func (s *MinIOStorage) AbortUpload(ctx context.Context, objectKey string, uploadID string) error {
	// 列出所有临时分片并删除
	prefix := chunkPrefix(uploadID)
	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})
	for obj := range objectCh {
		if obj.Err != nil {
			continue
		}
		s.client.RemoveObject(ctx, s.bucket, obj.Key, minio.RemoveObjectOptions{})
	}
	return nil
}
