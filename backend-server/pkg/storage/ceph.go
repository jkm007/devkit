package storage

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	"backend-server/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/signer"
)

// CephStorage Ceph RGW 存储
// 使用标准 S3 Multipart Upload API（Ceph RGW 不支持 ComposeObject）
type CephStorage struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
	region   string
	accessKey string
	secretKey string
}

// NewCephStorage 创建 Ceph RGW 存储实例
func NewCephStorage(cfg config.MinIOConfig) (*CephStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("Ceph RGW 客户端初始化失败: %w", err)
	}

	return &CephStorage{
		client:    client,
		bucket:    cfg.Bucket,
		endpoint:  cfg.Endpoint,
		useSSL:    cfg.UseSSL,
		region:    cfg.Region,
		accessKey: cfg.AccessKey,
		secretKey: cfg.SecretKey,
	}, nil
}

// getS3Endpoint 返回 S3 endpoint URL
func (s *CephStorage) getS3Endpoint() string {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, s.endpoint)
}

// InitiateUpload 初始化分片上传
func (s *CephStorage) InitiateUpload(ctx context.Context, objectKey string, contentType string) (string, error) {
	url := fmt.Sprintf("%s/%s/%s?uploads", s.getS3Endpoint(), s.bucket, objectKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// 使用 minio signer 签名请求
	signer.SignV4(*req, s.accessKey, s.secretKey, "", s.region)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ceph 初始化分片上传失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ceph 初始化分片上传失败: %s %s", resp.Status, string(body))
	}

	var result struct {
		UploadID string `xml:"UploadId"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析 InitiateMultipartUpload 响应失败: %w", err)
	}

	log.Printf("[INFO] Ceph 初始化分片上传: bucket=%s, key=%s, uploadID=%s", s.bucket, objectKey, result.UploadID)
	return result.UploadID, nil
}

// UploadPart 上传单个分片
func (s *CephStorage) UploadPart(ctx context.Context, objectKey string, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	url := fmt.Sprintf("%s/%s/%s?partNumber=%d&uploadId=%s", s.getS3Endpoint(), s.bucket, objectKey, partNumber, uploadID)

	// 读取数据到 buffer（需要计算 Content-Length）
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取分片数据失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.ContentLength = int64(len(data))

	// 使用 minio signer 签名请求
	signer.SignV4(*req, s.accessKey, s.secretKey, "", s.region)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ceph 上传分片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ceph 上传分片 %d 失败: %s %s", partNumber, resp.Status, string(body))
	}

	etag := resp.Header.Get("ETag")
	log.Printf("[DEBUG] Ceph UploadPart: bucket=%s, key=%s, part=%d, etag=%s", s.bucket, objectKey, partNumber, etag)
	return etag, nil
}

// CompleteUpload 合并分片
func (s *CephStorage) CompleteUpload(ctx context.Context, objectKey string, uploadID string, parts []CompletedPart) (string, error) {
	sort.Slice(parts, func(i, j int) bool { return parts[i].PartNumber < parts[j].PartNumber })

	// 构建 CompleteMultipartUpload 请求体
	type Part struct {
		PartNumber int    `xml:"PartNumber"`
		ETag       string `xml:"ETag"`
	}
	type CompleteMultipartUpload struct {
		XMLName xml.Name `xml:"CompleteMultipartUpload"`
		Parts   []Part   `xml:"Part"`
	}

	completeReq := CompleteMultipartUpload{}
	for _, p := range parts {
		completeReq.Parts = append(completeReq.Parts, Part{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		})
	}

	body, err := xml.Marshal(completeReq)
	if err != nil {
		return "", fmt.Errorf("序列化 CompleteMultipartUpload 请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s?uploadId=%s", s.getS3Endpoint(), s.bucket, objectKey, uploadID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.ContentLength = int64(len(body))
	req.Header.Set("Content-Type", "application/xml")

	// 使用 minio signer 签名请求
	signer.SignV4(*req, s.accessKey, s.secretKey, "", s.region)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ceph 合并分片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ceph 合并分片失败: %s %s", resp.Status, string(respBody))
	}

	log.Printf("[INFO] Ceph 合并分片成功: %s", objectKey)
	return s.GetURL(objectKey), nil
}

// AbortUpload 取消上传
func (s *CephStorage) AbortUpload(ctx context.Context, objectKey string, uploadID string) error {
	url := fmt.Sprintf("%s/%s/%s?uploadId=%s", s.getS3Endpoint(), s.bucket, objectKey, uploadID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	signer.SignV4(*req, s.accessKey, s.secretKey, "", s.region)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Upload 上传文件
func (s *CephStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取文件数据失败: %w", err)
	}
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}
	_, err = s.client.PutObject(ctx, s.bucket, objectKey, bytes.NewReader(buf), int64(len(buf)), opts)
	if err != nil {
		return "", fmt.Errorf("上传失败: %w", err)
	}
	return s.GetURL(objectKey), nil
}

// Download 下载文件（使用 HTTP 签名请求）
func (s *CephStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s/%s", s.getS3Endpoint(), s.bucket, objectKey)
	log.Printf("[DEBUG] Ceph Download: %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	signer.SignV4(*req, s.accessKey, s.secretKey, "", s.region)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] Ceph Download 请求失败: %v", err)
		return nil, fmt.Errorf("下载失败: %w", err)
	}

	log.Printf("[DEBUG] Ceph Download 响应: status=%d, content-length=%d", resp.StatusCode, resp.ContentLength)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("[ERROR] Ceph Download 失败: %s, body=%s", resp.Status, string(body))
		return nil, fmt.Errorf("下载失败: %s", resp.Status)
	}

	return resp.Body, nil
}

// DownloadRange 下载文件的指定范围（使用 HTTP 签名请求）
func (s *CephStorage) DownloadRange(ctx context.Context, objectKey string, offset int64, length int64) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s/%s", s.getS3Endpoint(), s.bucket, objectKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+length-1))
	signer.SignV4(*req, s.accessKey, s.secretKey, "", s.region)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("范围下载失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return nil, fmt.Errorf("范围下载失败: %s", resp.Status)
	}

	return resp.Body, nil
}

// Delete 删除文件
func (s *CephStorage) Delete(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
}

// GetURL 获取文件访问 URL
func (s *CephStorage) GetURL(objectKey string) string {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, objectKey)
}

// GetPresignedURL 获取临时访问 URL
func (s *CephStorage) GetPresignedURL(ctx context.Context, objectKey string, expire int64) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, time.Duration(expire)*time.Second, nil)
	if err != nil {
		return "", fmt.Errorf("生成临时 URL 失败: %w", err)
	}
	return url.String(), nil
}
