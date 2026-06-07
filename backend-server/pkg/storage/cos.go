package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"backend-server/config"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

// COStorage 腾讯云 COS 存储
type COStorage struct {
	cfg    config.COSConfig
	client *cos.Client
}

// NewCOStorage 创建 COS 存储实例
func NewCOStorage(cfg config.COSConfig) (*COStorage, error) {
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("解析 COS Bucket URL 失败: %w", err)
	}

	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})

	return &COStorage{cfg: cfg, client: client}, nil
}

// Upload 上传文件
func (s *COStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error) {
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}
	_, err := s.client.Object.Put(ctx, objectKey, reader, opt)
	if err != nil {
		return "", fmt.Errorf("COS 上传失败: %w", err)
	}
	return s.GetURL(objectKey), nil
}

// Download 下载文件
func (s *COStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	resp, err := s.client.Object.Get(ctx, objectKey, nil)
	if err != nil {
		return nil, fmt.Errorf("COS 下载失败: %w", err)
	}
	return resp.Body, nil
}

// DownloadRange 下载文件的指定范围
func (s *COStorage) DownloadRange(ctx context.Context, objectKey string, offset int64, length int64) (io.ReadCloser, error) {
	end := offset + length - 1
	opt := &cos.ObjectGetOptions{
		Range: fmt.Sprintf("bytes=%d-%d", offset, end),
	}
	resp, err := s.client.Object.Get(ctx, objectKey, opt)
	if err != nil {
		return nil, fmt.Errorf("COS 下载失败: %w", err)
	}
	return resp.Body, nil
}

// Delete 删除文件
func (s *COStorage) Delete(ctx context.Context, objectKey string) error {
	_, err := s.client.Object.Delete(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("COS 删除失败: %w", err)
	}
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
	presignedURL, err := s.client.Object.GetPresignedURL(ctx, http.MethodGet, objectKey, s.cfg.SecretID, s.cfg.SecretKey, time.Duration(expire)*time.Second, nil)
	if err != nil {
		return "", fmt.Errorf("生成 COS 签名 URL 失败: %w", err)
	}
	return presignedURL.String(), nil
}

// batchPrefixDelete 批量删除指定前缀的文件
func (s *COStorage) batchPrefixDelete(ctx context.Context, prefix string) error {
	isTruncated := true
	marker := ""
	for isTruncated {
		opt := &cos.BucketGetOptions{
			Prefix:  prefix,
			Marker:  marker,
			MaxKeys: 1000,
		}
		result, _, err := s.client.Bucket.Get(ctx, opt)
		if err != nil {
			return fmt.Errorf("列举 COS 对象失败: %w", err)
		}

		if len(result.Contents) == 0 {
			break
		}

		objects := make([]cos.Object, len(result.Contents))
		for i, obj := range result.Contents {
			objects[i] = cos.Object{Key: obj.Key}
		}

		_, _, err = s.client.Object.DeleteMulti(ctx, &cos.ObjectDeleteMultiOptions{
			Objects: objects,
			Quiet:   true,
		})
		if err != nil {
			return fmt.Errorf("批量删除 COS 对象失败: %w", err)
		}

		isTruncated = result.IsTruncated
		marker = result.NextMarker
	}
	return nil
}

// InitiateUpload 初始化分片上传
func (s *COStorage) InitiateUpload(ctx context.Context, objectKey string, contentType string) (string, error) {
	opt := &cos.InitiateMultipartUploadOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}
	resp, _, err := s.client.Object.InitiateMultipartUpload(ctx, objectKey, opt)
	if err != nil {
		return "", fmt.Errorf("COS 初始化分片上传失败: %w", err)
	}
	return resp.UploadID, nil
}

// UploadPart 上传分片
func (s *COStorage) UploadPart(ctx context.Context, objectKey string, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	opt := &cos.ObjectUploadPartOptions{
		ContentLength: size,
	}
	resp, err := s.client.Object.UploadPart(ctx, objectKey, uploadID, partNumber, reader, opt)
	if err != nil {
		return "", fmt.Errorf("COS 上传分片失败: %w", err)
	}
	return resp.Header.Get("ETag"), nil
}

// CompleteUpload 完成分片上传
func (s *COStorage) CompleteUpload(ctx context.Context, objectKey string, uploadID string, parts []CompletedPart) (string, error) {
	opt := &cos.CompleteMultipartUploadOptions{
		Parts: make([]cos.Object, len(parts)),
	}
	for i, p := range parts {
		opt.Parts[i] = cos.Object{
			Key:  fmt.Sprintf("%d", p.PartNumber),
			ETag: p.ETag,
		}
	}
	_, _, err := s.client.Object.CompleteMultipartUpload(ctx, objectKey, uploadID, opt)
	if err != nil {
		return "", fmt.Errorf("COS 完成分片上传失败: %w", err)
	}
	return s.GetURL(objectKey), nil
}

// AbortUpload 取消分片上传
func (s *COStorage) AbortUpload(ctx context.Context, objectKey string, uploadID string) error {
	_, err := s.client.Object.AbortMultipartUpload(ctx, objectKey, uploadID)
	if err != nil {
		return fmt.Errorf("COS 取消分片上传失败: %w", err)
	}
	return nil
}
