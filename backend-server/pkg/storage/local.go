package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"backend-server/config"
)

// LocalStorage 本地文件存储
type LocalStorage struct {
	cfg config.LocalConfig
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(cfg config.LocalConfig) *LocalStorage {
	// 确保目录存在
	os.MkdirAll(cfg.Path, 0755)
	return &LocalStorage{cfg: cfg}
}

// Upload 上传文件
func (s *LocalStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error) {
	fullPath := filepath.Join(s.cfg.Path, objectKey)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入内容
	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return s.GetURL(objectKey), nil
}

// Download 下载文件
func (s *LocalStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.cfg.Path, objectKey)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	return file, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, objectKey string) error {
	fullPath := filepath.Join(s.cfg.Path, objectKey)
	return os.Remove(fullPath)
}

// GetURL 获取文件访问 URL
func (s *LocalStorage) GetURL(objectKey string) string {
	return s.cfg.URLPrefix + "/" + objectKey
}

// GetPresignedURL 获取临时访问 URL（本地存储直接返回普通 URL）
func (s *LocalStorage) GetPresignedURL(ctx context.Context, objectKey string, expire int64) (string, error) {
	return s.GetURL(objectKey), nil
}
