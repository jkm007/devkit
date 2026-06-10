package service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"backend-server/pkg/storage"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// UploadService 文件上传服务
type UploadService struct {
	uploadRepo  *repository.UploadRepo
	assetRepo   *repository.FileAssetRepo
	fileRepo    *repository.FileRepo
	tagRepo     *repository.TagRepo
	fileTagRepo *repository.FileTagRepo
}

func NewUploadService() *UploadService {
	db := database.GetMySQL()
	return &UploadService{
		uploadRepo:  repository.NewUploadRepo(db),
		assetRepo:   repository.NewFileAssetRepo(db),
		fileRepo:    repository.NewFileRepo(db),
		tagRepo:     repository.NewTagRepo(db),
		fileTagRepo: repository.NewFileTagRepo(db),
	}
}

// getStorage 获取当前存储实例
func (s *UploadService) getStorage() storage.Storage {
	return storage.GetStorage()
}

// getUploader 获取分片上传器（如果支持）
func (s *UploadService) getUploader() storage.MultipartUploader {
	st := s.getStorage()
	if st == nil {
		return nil
	}
	uploader, _ := st.(storage.MultipartUploader)
	return uploader
}

// CheckResult 秒传检查结果
type CheckResult struct {
	Exists    bool   `json:"exists"`
	FileID    uint   `json:"fileId,omitempty"`
	ObjectKey string `json:"objectKey,omitempty"`
	URL       string `json:"url,omitempty"`
}

// CheckUpload 秒传检查
func (s *UploadService) CheckUpload(fileHash string, fileSize int64) (*CheckResult, error) {
	asset, err := s.assetRepo.GetByHash(fileHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 未找到，不是秒传
			return &CheckResult{Exists: false}, nil
		}
		return nil, fmt.Errorf("查询文件资产失败: %w", err)
	}
	// 大小校验
	if asset.FileSize != fileSize {
		return &CheckResult{Exists: false}, nil
	}
	// 存储类型校验：必须在当前存储中存在
	currentDriver := storage.GetStorageDriver()
	if asset.StorageType != currentDriver {
		return &CheckResult{Exists: false}, nil
	}
	// 秒传命中，增加引用计数
	s.assetRepo.IncrementRefCount(asset.ID)
	return &CheckResult{
		Exists:    true,
		ObjectKey: asset.ObjectKey,
	}, nil
}

// InitResult 初始化上传结果
type InitResult struct {
	UploadID    string `json:"uploadId"`
	UploadedParts []int `json:"uploadedParts"`
}

// InitUpload 初始化分片上传
func (s *UploadService) InitUpload(userID uint, fileName string, fileSize int64, fileHash string, contentType string, totalParts int, folderID uint) (*InitResult, error) {
	// 使用路由引擎生成 objectKey 并确定存储驱动
	routingResult, _, err := storage.Route(fileName, contentType, "user")
	if err != nil {
		// 路由失败，使用默认路径
		routingResult = &storage.RoutingResult{
			Driver:     "local",
			PathPrefix: "files/",
		}
	}

	// 根据路由结果获取对应的存储驱动
	storageDriver := storage.GetStorageByDriver(routingResult.Driver)
	if storageDriver == nil {
		return nil, fmt.Errorf("存储驱动 %s 不可用", routingResult.Driver)
	}
	uploader, ok := storageDriver.(storage.MultipartUploader)
	if !ok {
		return nil, fmt.Errorf("存储驱动 %s 不支持分片上传", routingResult.Driver)
	}

	// 生成唯一 objectKey
	tagger := storage.GetAutoTagger()
	var objectKey string
	if tagger != nil {
		objectKey = tagger.GenerateObjectKey(routingResult.PathPrefix, fileName)
	} else {
		ext := filepath.Ext(fileName)
		objectKey = fmt.Sprintf("files/%s/%d%s",
			time.Now().Format("2006/01/02"),
			time.Now().UnixNano(),
			ext)
	}

	// 生成唯一 uploadID（只使用ASCII十六进制字符，避免中文编码问题）
	hashInput := fmt.Sprintf("%s-%d", fileHash, time.Now().UnixNano())
	hashBytes := md5.Sum([]byte(hashInput))
	uploadID := fmt.Sprintf("%d-%x", time.Now().UnixNano(), hashBytes[:8])

	// 初始化分片上传（通知存储层）
	// 注意：本地存储会返回自己的 uploadID，MinIO 返回空字符串
	storageUploadID, err := uploader.InitiateUpload(context.Background(), objectKey, contentType)
	if err != nil {
		return nil, err
	}
	// 如果存储层返回了 uploadID（本地存储），使用它
	if storageUploadID != "" {
		uploadID = storageUploadID
	}

	// 创建上传任务记录
	task := &model.UploadTask{
		FileHash:      fileHash,
		UploadID:      uploadID,
		FileName:      fileName,
		FileSize:      fileSize,
		ContentType:   contentType,
		ObjectKey:     objectKey,
		StorageDriver: routingResult.Driver,
		TotalParts:    totalParts,
		Status:        "uploading",
		UserID:        userID,
		FolderID:      folderID,
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
	task, err := s.uploadRepo.GetTaskByUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("上传任务不存在")
	}
	if task.Status != "uploading" {
		return nil, fmt.Errorf("上传任务已结束")
	}

	// 根据任务的存储驱动获取对应的存储实例
	storageDriver := storage.GetStorageByDriver(task.StorageDriver)
	if storageDriver == nil {
		return nil, fmt.Errorf("存储驱动 %s 不可用", task.StorageDriver)
	}
	uploader, ok := storageDriver.(storage.MultipartUploader)
	if !ok {
		return nil, fmt.Errorf("存储驱动 %s 不支持分片上传", task.StorageDriver)
	}

	// 上传到存储
	rdr, ok := reader.(interface{ Read([]byte) (int, error) })
	if !ok {
		return nil, fmt.Errorf("无效的 reader")
	}
	etag, err := uploader.UploadPart(context.Background(), task.ObjectKey, uploadID, partNumber, rdr, size)
	if err != nil {
		// 更新任务失败状态
		s.uploadRepo.UpdateTaskFailed(uploadID, fmt.Sprintf("上传分片 %d 失败: %v", partNumber, err))
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

	// 更新进度
	uploadedParts := partNumber
	progress := int(float64(uploadedParts) / float64(task.TotalParts) * 100)
	s.uploadRepo.UpdateTaskProgress(uploadID, uploadedParts, progress)

	// Redis 缓存
	rdb := database.GetRedis()
	ctx := context.Background()
	rdb.HSet(ctx, "upload:"+uploadID, fmt.Sprintf("part:%d", partNumber), etag)

	return &UploadPartResult{ETag: etag}, nil
}

// RoutingInfo 路由信息
type RoutingInfo struct {
	Driver     string `json:"driver"`
	Bucket     string `json:"bucket,omitempty"`
	PathPrefix string `json:"pathPrefix,omitempty"`
	RuleName   string `json:"ruleName,omitempty"`
}

// CompleteResult 合并结果
type CompleteResult struct {
	FileID    uint         `json:"fileId"`
	ObjectKey string       `json:"objectKey"`
	URL       string       `json:"url"`
	Routing   *RoutingInfo `json:"routing,omitempty"`
}

// CompleteUpload 合并分片
func (s *UploadService) CompleteUpload(uploadID string) (*CompleteResult, error) {
	task, err := s.uploadRepo.GetTaskByUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("上传任务不存在")
	}

	// 根据任务的存储驱动获取对应的存储实例
	storageDriver := storage.GetStorageByDriver(task.StorageDriver)
	if storageDriver == nil {
		return nil, fmt.Errorf("存储驱动 %s 不可用", task.StorageDriver)
	}
	uploader, ok := storageDriver.(storage.MultipartUploader)
	if !ok {
		return nil, fmt.Errorf("存储驱动 %s 不支持分片上传", task.StorageDriver)
	}

	// 更新状态为处理中
	s.uploadRepo.UpdateTaskStatus(uploadID, "processing")

	// 获取所有已上传分片
	parts, err := s.uploadRepo.GetParts(task.ID)
	if err != nil || len(parts) == 0 {
		s.uploadRepo.UpdateTaskFailed(uploadID, "没有已上传的分片")
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
	_, err = uploader.CompleteUpload(context.Background(), task.ObjectKey, uploadID, completedParts)
	if err != nil {
		s.uploadRepo.UpdateTaskFailed(uploadID, fmt.Sprintf("合并分片失败: %v", err))
		return nil, err
	}

	// 事务：更新任务状态 + 创建文件资产 + 创建文件条目 + 创建标签
	db := database.GetMySQL()
	var entryID uint
	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新任务完成状态
		now := time.Now()
		if err := tx.Model(&model.UploadTask{}).Where("upload_id = ?", uploadID).
			Updates(map[string]interface{}{
				"status":       "completed",
				"progress":     100,
				"completed_at": now,
			}).Error; err != nil {
			return fmt.Errorf("更新任务状态失败: %w", err)
		}

		// 获取路由结果（用于标签）
		_, tags, routeErr := storage.Route(task.FileName, task.ContentType, "user")
		if routeErr != nil {
			// 路由失败不影响上传
		}

		// 查找或创建文件资产（秒传映射）
		// 使用任务中记录的存储驱动
		var asset model.FileAsset
		err = tx.Where("file_hash = ?", task.FileHash).First(&asset).Error
		if err == nil {
			// 文件已存在，增加引用计数并更新存储信息（如果驱动变化）
			updates := map[string]interface{}{
				"ref_count": gorm.Expr("ref_count + 1"),
			}
			// 如果存储驱动发生变化，更新存储类型和对象键
			if asset.StorageType != task.StorageDriver {
				updates["storage_type"] = task.StorageDriver
				updates["object_key"] = task.ObjectKey
			}
			if err := tx.Model(&asset).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新文件资产失败: %w", err)
			}
		} else {
			// 文件不存在，创建新资产
			asset = model.FileAsset{
				FileHash:    task.FileHash,
				ObjectKey:   task.ObjectKey,
				FileName:    task.FileName,
				FileSize:    task.FileSize,
				ContentType: task.ContentType,
				StorageType: task.StorageDriver,
				RefCount:    1,
			}
			if err := tx.Create(&asset).Error; err != nil {
				return fmt.Errorf("创建文件资产失败: %w", err)
			}
		}

		// 创建文件条目（用于文件列表显示）
		entry := &model.FileEntry{
			FolderID:     task.FolderID,
			FileAssetID:  asset.ID,
			Name:         task.FileName,
			Size:         task.FileSize,
			ContentType:  task.ContentType,
			UserID:       task.UserID,
		}
		if err := tx.Create(entry).Error; err != nil {
			return fmt.Errorf("创建文件条目失败: %w", err)
		}
		entryID = entry.ID

		// 自动打标签
		if len(tags) > 0 {
			for _, tag := range tags {
				// 查找标签
				tagModel, err := s.tagRepo.GetByKeyValue(tag.Key, tag.Value)
				if err != nil {
					continue // 标签不存在，跳过
				}
				// 创建文件标签关联
				fileTag := &model.FileTag{
					FileID: entry.ID,
					TagID:  tagModel.ID,
					Source: "auto",
				}
				tx.Create(fileTag)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 清理 Redis
	rdb := database.GetRedis()
	rdb.Del(context.Background(), "upload:"+uploadID)

	// 获取路由信息
	routingResult, _, _ := storage.Route(task.FileName, task.ContentType, "user")
	var routingInfo *RoutingInfo
	if routingResult != nil {
		routingInfo = &RoutingInfo{
			Driver:     routingResult.Driver,
			Bucket:     routingResult.Bucket,
			PathPrefix: routingResult.PathPrefix,
			RuleName:   routingResult.RuleName,
		}
	}

	// 根据存储驱动返回合适的 URL
	var fileURL string
	if task.StorageDriver == "local" {
		fileURL = "/files/" + strconv.FormatUint(uint64(entryID), 10) + "/view"
	} else {
		// 云存储：返回 direct-url 接口（前端获取 presigned URL）
		fileURL = "/files/" + strconv.FormatUint(uint64(entryID), 10) + "/direct-url"
	}
	return &CompleteResult{
		FileID:    entryID,
		ObjectKey: task.ObjectKey,
		URL:       fileURL,
		Routing:   routingInfo,
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

// GetUserUploadTasks 获取用户的上传任务列表
func (s *UploadService) GetUserUploadTasks(userID uint, limit int) ([]model.UploadTask, error) {
	return s.uploadRepo.GetUserTasks(userID, limit)
}

// GetUploadTaskByID 根据ID获取上传任务
func (s *UploadService) GetUploadTaskByID(id uint) (*model.UploadTask, error) {
	return s.uploadRepo.GetTaskByID(id)
}

// GetTaskByUploadID 根据 uploadID 获取上传任务
func (s *UploadService) GetTaskByUploadID(uploadID string) (*model.UploadTask, error) {
	return s.uploadRepo.GetTaskByUploadID(uploadID)
}

// TaskStatusResponse 任务状态响应
type TaskStatusResponse struct {
	ID            uint       `json:"id"`
	UploadID      string     `json:"uploadId"`
	FileName      string     `json:"fileName"`
	FileSize      int64      `json:"fileSize"`
	ContentType   string     `json:"contentType"`
	TotalParts    int        `json:"totalParts"`
	UploadedParts int        `json:"uploadedParts"`
	Progress      int        `json:"progress"`
	Status        string     `json:"status"`
	ErrorMessage  string     `json:"errorMessage,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// GetTaskStatusResponse 获取任务状态响应
func (s *UploadService) GetTaskStatusResponse(task *model.UploadTask) *TaskStatusResponse {
	return &TaskStatusResponse{
		ID:            task.ID,
		UploadID:      task.UploadID,
		FileName:      task.FileName,
		FileSize:      task.FileSize,
		ContentType:   task.ContentType,
		TotalParts:    task.TotalParts,
		UploadedParts: task.UploadedParts,
		Progress:      task.Progress,
		Status:        task.Status,
		ErrorMessage:  task.ErrorMessage,
		CompletedAt:   task.CompletedAt,
		CreatedAt:     task.CreatedAt,
	}
}

// generateObjectKey 生成存储路径
func generateObjectKey(fileName string, fileHash string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	// 使用完整 hash 或至少 32 字符前缀以降低碰撞风险
	hashLen := 32
	if len(fileHash) < hashLen {
		hashLen = len(fileHash)
	}
	if hashLen == 0 {
		// 极端情况：无 hash，使用时间戳
		fileHash = time.Now().Format("20060102150405")
	} else {
		fileHash = fileHash[:hashLen]
	}
	return fmt.Sprintf("files/%s/%s%s",
		time.Now().Format("2006/01/02"),
		fileHash,
		ext)
}
