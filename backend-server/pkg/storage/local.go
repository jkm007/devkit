package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"backend-server/config"
)

// LocalStorage 本地文件存储
type LocalStorage struct {
	cfg config.LocalConfig
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(cfg config.LocalConfig) *LocalStorage {
	os.MkdirAll(cfg.Path, 0755)
	return &LocalStorage{cfg: cfg}
}

// validateObjectKey 校验 objectKey 防止路径穿越
func validateObjectKey(objectKey string) error {
	if objectKey == "" {
		return fmt.Errorf("objectKey 不能为空")
	}
	cleaned := filepath.Clean(objectKey)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("objectKey 包含非法路径: %s", objectKey)
	}
	if filepath.IsAbs(cleaned) {
		return fmt.Errorf("objectKey 不能是绝对路径: %s", objectKey)
	}
	return nil
}

// fullPath 获取文件完整路径并校验
func (s *LocalStorage) fullPath(objectKey string) (string, error) {
	if err := validateObjectKey(objectKey); err != nil {
		return "", err
	}
	fullPath := filepath.Join(s.cfg.Path, objectKey)
	absPath, _ := filepath.Abs(fullPath)
	absBase, _ := filepath.Abs(s.cfg.Path)
	if !strings.HasPrefix(absPath, absBase) {
		return "", fmt.Errorf("路径越界: %s", objectKey)
	}
	return fullPath, nil
}

// Upload 上传文件
func (s *LocalStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error) {
	fullPath, err := s.fullPath(objectKey)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return s.GetURL(objectKey), nil
}

// Download 下载文件
func (s *LocalStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	fullPath, err := s.fullPath(objectKey)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	return file, nil
}

// DownloadRange 下载文件的指定范围
func (s *LocalStorage) DownloadRange(ctx context.Context, objectKey string, offset int64, length int64) (io.ReadCloser, error) {
	fullPath, err := s.fullPath(objectKey)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}

	// 定位到指定偏移
	actualOffset, err := file.Seek(offset, io.SeekStart)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("定位文件失败: %w", err)
	}
	if actualOffset != offset {
		file.Close()
		return nil, fmt.Errorf("Seek 偏移量不匹配: 期望 %d, 实际 %d", offset, actualOffset)
	}

	// 返回限制长度的读取器
	return &limitedReadCloser{file: file, remaining: length}, nil
}

// limitedReadCloser 限制读取长度的 ReadCloser
type limitedReadCloser struct {
	file      *os.File
	remaining int64
}

func (l *limitedReadCloser) Read(p []byte) (n int, err error) {
	if l.remaining <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.remaining {
		p = p[:l.remaining]
	}
	n, err = l.file.Read(p)
	l.remaining -= int64(n)
	return n, err
}

func (l *limitedReadCloser) Close() error {
	return l.file.Close()
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, objectKey string) error {
	fullPath, err := s.fullPath(objectKey)
	if err != nil {
		return err
	}
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

// uploadTmpDir 分片临时存储目录
func (s *LocalStorage) uploadTmpDir(uploadID string) string {
	return filepath.Join(s.cfg.Path, ".uploads", uploadID)
}

// InitiateUpload 初始化分片上传
func (s *LocalStorage) InitiateUpload(ctx context.Context, objectKey string, contentType string) (string, error) {
	// 生成 uploadID
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成 uploadID 失败: %w", err)
	}
	uploadID := hex.EncodeToString(b)

	// 创建临时目录
	tmpDir := s.uploadTmpDir(uploadID)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 写入元数据
	metaFile := filepath.Join(tmpDir, ".meta")
	if err := os.WriteFile(metaFile, []byte(objectKey), 0644); err != nil {
		return "", fmt.Errorf("写入元数据失败: %w", err)
	}

	return uploadID, nil
}

// UploadPart 上传分片
func (s *LocalStorage) UploadPart(ctx context.Context, objectKey string, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	tmpDir := s.uploadTmpDir(uploadID)
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		return "", fmt.Errorf("上传任务不存在: %s", uploadID)
	}

	partFile := filepath.Join(tmpDir, fmt.Sprintf("part-%06d", partNumber))
	f, err := os.Create(partFile)
	if err != nil {
		return "", fmt.Errorf("创建分片文件失败: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, reader)
	if err != nil {
		return "", fmt.Errorf("写入分片失败: %w", err)
	}

	// 用文件大小的 hex 作为简易 ETag
	etag := fmt.Sprintf("%x", written)
	return etag, nil
}

// CompleteUpload 合并分片完成上传
func (s *LocalStorage) CompleteUpload(ctx context.Context, objectKey string, uploadID string, parts []CompletedPart) (string, error) {
	tmpDir := s.uploadTmpDir(uploadID)

	// 读取元数据中的 objectKey（优先使用参数传入的）
	metaFile := filepath.Join(tmpDir, ".meta")
	if data, err := os.ReadFile(metaFile); err == nil && len(data) > 0 {
		objectKey = string(data)
	}

	// 确保目标目录存在
	destPath, err := s.fullPath(objectKey)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return "", fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 按分片号排序
	sortedParts := make([]CompletedPart, len(parts))
	copy(sortedParts, parts)
	sort.Slice(sortedParts, func(i, j int) bool {
		return sortedParts[i].PartNumber < sortedParts[j].PartNumber
	})

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	// 依次合并分片
	for _, part := range sortedParts {
		partFile := filepath.Join(tmpDir, fmt.Sprintf("part-%06d", part.PartNumber))
		src, err := os.Open(partFile)
		if err != nil {
			return "", fmt.Errorf("打开分片文件失败 (part %d): %w", part.PartNumber, err)
		}
		if _, err := io.Copy(destFile, src); err != nil {
			src.Close()
			return "", fmt.Errorf("合并分片失败 (part %d): %w", part.PartNumber, err)
		}
		src.Close()
	}

	// 清理临时目录
	os.RemoveAll(tmpDir)

	return s.GetURL(objectKey), nil
}

// AbortUpload 取消分片上传
func (s *LocalStorage) AbortUpload(ctx context.Context, objectKey string, uploadID string) error {
	tmpDir := s.uploadTmpDir(uploadID)
	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("清理临时文件失败: %w", err)
	}
	return nil
}
