package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"backend-server/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSStorage 阿里云 OSS 存储
type OSSStorage struct {
	cfg    config.OSSConfig
	client *oss.Client
	bucket *oss.Bucket
}

// NewOSSStorage 创建 OSS 存储实例
func NewOSSStorage(cfg config.OSSConfig) (*OSSStorage, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建 OSS 客户端失败: %w", err)
	}

	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("获取 OSS Bucket 失败: %w", err)
	}

	return &OSSStorage{cfg: cfg, client: client, bucket: bucket}, nil
}

// Upload 上传文件
func (s *OSSStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error) {
	options := []oss.Option{
		oss.ContentType(contentType),
	}
	if err := s.bucket.PutObject(objectKey, reader, options...); err != nil {
		return "", fmt.Errorf("OSS 上传失败: %w", err)
	}
	return s.GetURL(objectKey), nil
}

// Download 下载文件
func (s *OSSStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	body, err := s.bucket.GetObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("OSS 下载失败: %w", err)
	}
	return body, nil
}

// Delete 删除文件
func (s *OSSStorage) Delete(ctx context.Context, objectKey string) error {
	if err := s.bucket.DeleteObject(objectKey); err != nil {
		return fmt.Errorf("OSS 删除失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问 URL
func (s *OSSStorage) GetURL(objectKey string) string {
	if s.cfg.CDNDomain != "" {
		return fmt.Sprintf("https://%s/%s", s.cfg.CDNDomain, objectKey)
	}
	endpoint := s.cfg.Endpoint
	if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
	} else if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	}
	endpoint = strings.TrimSuffix(endpoint, "/")
	return fmt.Sprintf("https://%s.%s/%s", s.cfg.Bucket, endpoint, objectKey)
}

// GetPresignedURL 获取临时访问 URL
func (s *OSSStorage) GetPresignedURL(ctx context.Context, objectKey string, expire int64) (string, error) {
	signedURL, err := s.bucket.SignURL(objectKey, oss.HTTPGet, expire)
	if err != nil {
		return "", fmt.Errorf("生成 OSS 签名 URL 失败: %w", err)
	}
	return signedURL, nil
}

// batchPrefixDelete 批量删除指定前缀的文件
func (s *OSSStorage) batchPrefixDelete(ctx context.Context, prefix string) error {
	marker := oss.Marker("")
	for {
		result, err := s.bucket.ListObjects(oss.Prefix(prefix), marker, oss.MaxKeys(1000))
		if err != nil {
			return fmt.Errorf("列举 OSS 对象失败: %w", err)
		}
		if len(result.Objects) == 0 {
			break
		}
		keys := make([]string, 0, len(result.Objects))
		for _, obj := range result.Objects {
			keys = append(keys, obj.Key)
		}
		_, err = s.bucket.DeleteObjects(keys, oss.DeleteObjectsQuiet(true))
		if err != nil {
			return fmt.Errorf("批量删除 OSS 对象失败: %w", err)
		}
		if !result.IsTruncated {
			break
		}
		marker = oss.Marker(result.NextMarker)
	}
	return nil
}

// InitiateUpload 初始化分片上传
func (s *OSSStorage) InitiateUpload(ctx context.Context, objectKey string, contentType string) (string, error) {
	imur, err := s.bucket.InitiateMultipartUpload(objectKey, oss.ContentType(contentType))
	if err != nil {
		return "", fmt.Errorf("OSS 初始化分片上传失败: %w", err)
	}
	return imur.UploadID, nil
}

// UploadPart 上传分片
func (s *OSSStorage) UploadPart(ctx context.Context, objectKey string, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	imur := oss.InitiateMultipartUploadResult{
		Key:      objectKey,
		UploadID: uploadID,
	}
	part, err := s.bucket.UploadPart(imur, reader, size, partNumber)
	if err != nil {
		return "", fmt.Errorf("OSS 上传分片失败: %w", err)
	}
	return part.ETag, nil
}

// CompleteUpload 完成分片上传
func (s *OSSStorage) CompleteUpload(ctx context.Context, objectKey string, uploadID string, parts []CompletedPart) (string, error) {
	imur := oss.InitiateMultipartUploadResult{
		Key:      objectKey,
		UploadID: uploadID,
	}
	ossParts := make([]oss.UploadPart, len(parts))
	for i, p := range parts {
		ossParts[i] = oss.UploadPart{
			ETag:       p.ETag,
			PartNumber: p.PartNumber,
		}
	}
	_, err := s.bucket.CompleteMultipartUpload(imur, ossParts)
	if err != nil {
		return "", fmt.Errorf("OSS 完成分片上传失败: %w", err)
	}
	return s.GetURL(objectKey), nil
}

// AbortUpload 取消分片上传
func (s *OSSStorage) AbortUpload(ctx context.Context, objectKey string, uploadID string) error {
	imur := oss.InitiateMultipartUploadResult{
		Key:      objectKey,
		UploadID: uploadID,
	}
	if err := s.bucket.AbortMultipartUpload(imur); err != nil {
		return fmt.Errorf("OSS 取消分片上传失败: %w", err)
	}
	return nil
}

// ListUploadedParts 列举已上传的分片
func (s *OSSStorage) ListUploadedParts(ctx context.Context, objectKey string, uploadID string) ([]UploadedPart, error) {
	imur := oss.InitiateMultipartUploadResult{
		Key:      objectKey,
		UploadID: uploadID,
	}
	result, err := s.bucket.ListUploadedParts(imur)
	if err != nil {
		return nil, fmt.Errorf("OSS 列举已上传分片失败: %w", err)
	}
	parts := make([]UploadedPart, len(result.UploadedParts))
	for i, p := range result.UploadedParts {
		parts[i] = UploadedPart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
			Size:       int64(p.Size),
		}
	}
	return parts, nil
}

// GetFullPath 获取完整存储路径
func (s *OSSStorage) GetFullPath(prefix, fileName string) string {
	datePath := time.Now().Format("20060102")
	encodedName := url.PathEscape(fileName)
	return fmt.Sprintf("%s/%s/%s", prefix, datePath, encodedName)
}
