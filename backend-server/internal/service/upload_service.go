package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"backend-server/pkg/storage"

	"github.com/redis/go-redis/v9"
)

// UploadService 文件上传服务
type UploadService struct {
	uploadRepo *repository.UploadRepo
	assetRepo  *repository.FileAssetRepo
	fileRepo   *repository.FileRepo
}

func NewUploadService() *UploadService {
	db := database.GetMySQL()
	return &UploadService{
		uploadRepo: repository.NewUploadRepo(db),
		assetRepo:  repository.NewFileAssetRepo(db),
		fileRepo:   repository.NewFileRepo(db),
	}
}

// getStorage 获取当前存储实例
func (s *UploadService) getStorage() storage.Storage {
	return storage.GetStorage()
}

// getUploader 获取分片上传器（如果支持）
func (s *UploadService) getUploader() storage.MultipartUploader {
	uploader, _ := s.getStorage().(storage.MultipartUploader)
	return uploader
}

// CheckResult 秒传检查结果
type CheckResult struct {
	Exists    bool   `json:"exists"`
	ObjectKey string `json:"objectKey,omitempty"`
	URL       string `json:"url,omitempty"`
}

// CheckUpload 秒传检查
func (s *UploadService) CheckUpload(fileHash string, fileSize int64) (*CheckResult, error) {
	asset, err := s.assetRepo.GetByHash(fileHash)
	if err != nil {
		// 未找到，不是秒传
		return &CheckResult{Exists: false}, nil
	}
	// 大小校验
	if asset.FileSize != fileSize {
		return &CheckResult{Exists: false}, nil
	}
	// 秒传命中，增加引用计数
	s.assetRepo.IncrementRefCount(asset.ID)
	return &CheckResult{
		Exists:    true,
		ObjectKey: asset.ObjectKey,
		URL:       s.getStorage().GetURL(asset.ObjectKey),
	}, nil
}

// InitResult 初始化上传结果
type InitResult struct {
	UploadID    string `json:"uploadId"`
	UploadedParts []int `json:"uploadedParts"`
}

// InitUpload 初始化分片上传
func (s *UploadService) InitUpload(userID uint, fileName string, fileSize int64, fileHash string, contentType string, totalParts int) (*InitResult, error) {
	uploader := s.getUploader()
	if uploader == nil {
		return nil, fmt.Errorf("当前存储驱动不支持分片上传")
	}

	// 生成唯一 objectKey（使用时间戳避免中文编码问题）
	ext := filepath.Ext(fileName)
	objectKey := fmt.Sprintf("files/%s/%d%s",
		time.Now().Format("2006/01/02"),
		time.Now().UnixNano(),
		ext)

	// 生成唯一 uploadID（UUID 格式）
	uploadID := fmt.Sprintf("%d-%s", time.Now().UnixNano(), fileHash[:16])

	// 初始化分片上传（通知存储层）
	_, err := uploader.InitiateUpload(context.Background(), objectKey, contentType)
	if err != nil {
		return nil, err
	}

	// 创建上传任务记录
	task := &model.UploadTask{
		FileHash:    fileHash,
		UploadID:    uploadID,
		FileName:    fileName,
		FileSize:    fileSize,
		ContentType: contentType,
		ObjectKey:   objectKey,
		TotalParts:  totalParts,
		Status:      "uploading",
		UserID:      userID,
	}
	if err := s.uploadRepo.CreateTask(task); err != nil {
		return nil, fmt.Errorf("创建上传任务失败: %w", err)
	}

	// Redis 缓存上传状态（24h TTL）
	rdb := database.GetRedis()
	ctx := context.Background()
	rdb.HSet(ctx, "upload:"+uploadID, "objectKey", objectKey, "totalParts", totalParts)
	rdb.Expire(ctx, "upload:"+uploadID, 24*time.Hour)

	return &InitResult{
		UploadID:      uploadID,
		UploadedParts: []int{},
	}, nil
}

// UploadPartResult 分片上传结果
type UploadPartResult struct {
	ETag string `json:"etag"`
}

// UploadPart 上传分片
func (s *UploadService) UploadPart(uploadID string, partNumber int, reader interface{}, size int64) (*UploadPartResult, error) {
	uploader := s.getUploader()
	if uploader == nil {
		return nil, fmt.Errorf("当前存储驱动不支持分片上传")
	}

	task, err := s.uploadRepo.GetTaskByUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("上传任务不存在")
	}
	if task.Status != "uploading" {
		return nil, fmt.Errorf("上传任务已结束")
	}

	// 上传到存储
	rdr, ok := reader.(interface{ Read([]byte) (int, error) })
	if !ok {
		return nil, fmt.Errorf("无效的 reader")
	}
	etag, err := uploader.UploadPart(context.Background(), task.ObjectKey, uploadID, partNumber, rdr, size)
	if err != nil {
		return nil, err
	}

	// 记录分片
	part := &model.UploadedPart{
		UploadTaskID: task.ID,
		PartNumber:   partNumber,
		ETag:         etag,
		Size:         size,
	}
	s.uploadRepo.CreatePart(part)

	// Redis 缓存
	rdb := database.GetRedis()
	ctx := context.Background()
	rdb.HSet(ctx, "upload:"+uploadID, fmt.Sprintf("part:%d", partNumber), etag)

	return &UploadPartResult{ETag: etag}, nil
}

// CompleteResult 合并结果
type CompleteResult struct {
	ObjectKey string `json:"objectKey"`
	URL       string `json:"url"`
}

// CompleteUpload 合并分片
func (s *UploadService) CompleteUpload(uploadID string) (*CompleteResult, error) {
	uploader := s.getUploader()
	if uploader == nil {
		return nil, fmt.Errorf("当前存储驱动不支持分片上传")
	}

	task, err := s.uploadRepo.GetTaskByUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("上传任务不存在")
	}

	// 获取所有已上传分片
	parts, err := s.uploadRepo.GetParts(task.ID)
	if err != nil || len(parts) == 0 {
		return nil, fmt.Errorf("没有已上传的分片")
	}

	// 转换为 CompletedPart
	completedParts := make([]storage.CompletedPart, len(parts))
	for i, p := range parts {
		completedParts[i] = storage.CompletedPart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		}
	}

	// 合并
	url, err := uploader.CompleteUpload(context.Background(), task.ObjectKey, uploadID, completedParts)
	if err != nil {
		return nil, err
	}

	// 更新任务状态
	s.uploadRepo.UpdateTaskStatus(uploadID, "completed")

	// 创建文件资产（秒传映射）
	asset := &model.FileAsset{
		FileHash:    task.FileHash,
		ObjectKey:   task.ObjectKey,
		FileName:    task.FileName,
		FileSize:    task.FileSize,
		ContentType: task.ContentType,
		RefCount:    1,
	}
	s.assetRepo.Create(asset)

	// 创建文件条目（用于文件列表显示）
	entry := &model.FileEntry{
		FolderID:     0, // 根目录
		FileAssetID:  asset.ID,
		Name:         task.FileName,
		Size:         task.FileSize,
		ContentType:  task.ContentType,
		UserID:       task.UserID,
	}
	s.fileRepo.CreateEntry(entry)

	// 清理 Redis
	rdb := database.GetRedis()
	rdb.Del(context.Background(), "upload:"+uploadID)

	return &CompleteResult{
		ObjectKey: task.ObjectKey,
		URL:       url,
	}, nil
}

// AbortUpload 取消上传
func (s *UploadService) AbortUpload(uploadID string) error {
	task, err := s.uploadRepo.GetTaskByUploadID(uploadID)
	if err != nil {
		return nil // 任务不存在，视为成功
	}

	if uploader := s.getUploader(); uploader != nil {
		uploader.AbortUpload(context.Background(), task.ObjectKey, uploadID)
	}

	s.uploadRepo.UpdateTaskStatus(uploadID, "aborted")
	s.uploadRepo.DeletePartsByTaskID(task.ID)

	// 清理 Redis
	rdb := database.GetRedis()
	rdb.Del(context.Background(), "upload:"+uploadID)

	return nil
}

// GetUploadStatus 获取上传状态（断点续传）
func (s *UploadService) GetUploadStatus(uploadID string) (*InitResult, error) {
	task, err := s.uploadRepo.GetTaskByUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("上传任务不存在")
	}
	if task.Status != "uploading" {
		return nil, fmt.Errorf("上传任务已结束: %s", task.Status)
	}

	parts, err := s.uploadRepo.GetParts(task.ID)
	if err != nil {
		return nil, err
	}

	partNumbers := make([]int, len(parts))
	for i, p := range parts {
		partNumbers[i] = p.PartNumber
	}

	return &InitResult{
		UploadID:      uploadID,
		UploadedParts: partNumbers,
	}, nil
}

// CleanupStaleUploads 清理过期上传任务
func (s *UploadService) CleanupStaleUploads(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan).Unix()
	tasks, err := s.uploadRepo.GetStaleTasks(cutoff)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if uploader := s.getUploader(); uploader != nil {
			uploader.AbortUpload(context.Background(), task.ObjectKey, task.UploadID)
		}
		s.uploadRepo.UpdateTaskStatus(task.UploadID, "aborted")
		s.uploadRepo.DeletePartsByTaskID(task.ID)

		rdb := database.GetRedis()
		rdb.Del(context.Background(), "upload:"+task.UploadID)
	}

	_ = redis.Nil // avoid unused import
	return nil
}

// GetAssetByHash 根据哈希获取文件资产
func (s *UploadService) GetAssetByHash(hash string) (*model.FileAsset, error) {
	return s.assetRepo.GetByHash(hash)
}

// generateObjectKey 生成存储路径
func generateObjectKey(fileName string, fileHash string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	return fmt.Sprintf("files/%s/%s%s",
		time.Now().Format("2006/01/02"),
		fileHash[:16],
		ext)
}
