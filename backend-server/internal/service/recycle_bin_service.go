package service

import (
	"context"
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"backend-server/pkg/storage"
)

// RecycleBinService 回收站服务
type RecycleBinService struct {
	fileRepo  *repository.FileRepo
	assetRepo *repository.FileAssetRepo
	shareRepo *repository.FileShareRepo
	fileTagRepo *repository.FileTagRepo
	taskRepo  *repository.ScheduledTaskRepo
	userRepo  *repository.UserRepo
}

func NewRecycleBinService() *RecycleBinService {
	db := database.GetMySQL()
	return &RecycleBinService{
		fileRepo:    repository.NewFileRepo(db),
		assetRepo:   repository.NewFileAssetRepo(db),
		shareRepo:   repository.NewFileShareRepo(db),
		fileTagRepo: repository.NewFileTagRepo(db),
		taskRepo:    repository.NewScheduledTaskRepo(db),
		userRepo:    repository.NewUserRepo(db),
	}
}

// GetRecycleBinList 获取回收站列表
func (s *RecycleBinService) GetRecycleBinList(userID uint, page, pageSize int) ([]model.RecycleBinItem, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	entries, total, err := s.fileRepo.ListRecycleBin(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 批量查询用户信息
	userIDs := make(map[uint]bool)
	for _, entry := range entries {
		userIDs[entry.UserID] = true
	}
	userMap := make(map[uint]*model.User)
	for uid := range userIDs {
		user, err := s.userRepo.GetByID(uid)
		if err == nil {
			userMap[uid] = user
		}
	}

	items := make([]model.RecycleBinItem, 0, len(entries))
	for _, entry := range entries {
		item := model.RecycleBinItem{
			ID:              entry.ID,
			Name:            entry.Name,
			Size:            entry.Size,
			ContentType:     entry.ContentType,
			FolderID:        entry.FolderID,
			UserID:          entry.UserID,
			DeletedAt:       entry.DeletedAt,
			RecycleExpireAt: entry.RecycleExpireAt,
		}

		// 计算剩余天数
		if entry.RecycleExpireAt != nil {
			days := int(time.Until(*entry.RecycleExpireAt).Hours() / 24)
			if days < 0 {
				days = 0
			}
			item.DaysRemaining = days
		} else {
			item.DaysRemaining = 7 // 默认 7 天
		}

		// 添加用户名
		if user, ok := userMap[entry.UserID]; ok {
			item.UserName = user.Nickname
		}

		items = append(items, item)
	}

	return items, total, nil
}

// RestoreFile 从回收站恢复文件
func (s *RecycleBinService) RestoreFile(userID uint, fileID uint, hasPermission bool) error {
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}
	if entry.DeletedAt == nil {
		return fmt.Errorf("文件不在回收站中")
	}
	if !hasPermission && entry.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	return s.fileRepo.RestoreEntry(fileID)
}

// BatchRestoreFiles 批量恢复文件
func (s *RecycleBinService) BatchRestoreFiles(userID uint, fileIDs []uint, hasPermission bool) (int, []string) {
	restored := 0
	errList := []string{}

	for _, fileID := range fileIDs {
		err := s.RestoreFile(userID, fileID, hasPermission)
		if err != nil {
			errList = append(errList, fmt.Sprintf("文件 %d: %s", fileID, err.Error()))
		} else {
			restored++
		}
	}

	return restored, errList
}

// PermanentDeleteFile 永久删除文件
func (s *RecycleBinService) PermanentDeleteFile(userID uint, fileID uint, hasPermission bool) error {
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}
	if !hasPermission && entry.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	return s.permanentDeleteEntry(entry)
}

// BatchPermanentDelete 批量永久删除
func (s *RecycleBinService) BatchPermanentDelete(userID uint, fileIDs []uint, hasPermission bool) (int, []string) {
	deleted := 0
	errList := []string{}

	for _, fileID := range fileIDs {
		err := s.PermanentDeleteFile(userID, fileID, hasPermission)
		if err != nil {
			errList = append(errList, fmt.Sprintf("文件 %d: %s", fileID, err.Error()))
		} else {
			deleted++
		}
	}

	return deleted, errList
}

// EmptyRecycleBin 清空回收站
func (s *RecycleBinService) EmptyRecycleBin(userID uint) error {
	// 先获取所有回收站文件，删除存储对象
	entries, _, err := s.fileRepo.ListRecycleBin(userID, 1, 10000)
	if err != nil {
		return fmt.Errorf("获取回收站文件失败: %w", err)
	}

	// 删除每个文件的存储对象和关联数据
	for _, entry := range entries {
		s.cleanupFileData(&entry)
	}

	// 永久删除所有回收站文件
	return s.fileRepo.EmptyRecycleBin(userID)
}

// CleanupExpiredFiles 清理过期文件（定时任务调用）
func (s *RecycleBinService) CleanupExpiredFiles() (int, error) {
	entries, err := s.fileRepo.GetExpiredRecycleBinEntries()
	if err != nil {
		return 0, fmt.Errorf("获取过期文件失败: %w", err)
	}

	deleted := 0
	for _, entry := range entries {
		if err := s.permanentDeleteEntry(&entry); err != nil {
			fmt.Printf("永久删除文件失败: id=%d, err=%v\n", entry.ID, err)
			continue
		}
		deleted++
	}

	return deleted, nil
}

// permanentDeleteEntry 永久删除单个文件条目及其关联数据
func (s *RecycleBinService) permanentDeleteEntry(entry *model.FileEntry) error {
	// 清理文件关联数据
	s.cleanupFileData(entry)

	// 永久删除文件条目
	return s.fileRepo.HardDeleteEntry(entry.ID)
}

// cleanupFileData 清理文件的关联数据（分享、标签、存储对象）
func (s *RecycleBinService) cleanupFileData(entry *model.FileEntry) {
	// 删除分享记录
	if err := s.shareRepo.DeleteByFileID(entry.ID); err != nil {
		fmt.Printf("删除分享记录失败: fileID=%d, err=%v\n", entry.ID, err)
	}

	// 删除文件标签
	if err := s.fileTagRepo.DeleteByFileID(entry.ID); err != nil {
		fmt.Printf("删除文件标签失败: fileID=%d, err=%v\n", entry.ID, err)
	}

	// 处理文件资产
	if entry.FileAssetID > 0 {
		asset, err := s.assetRepo.GetByID(entry.FileAssetID)
		if err == nil {
			if asset.RefCount <= 1 {
				// 引用计数减到 0，删除存储对象和资产记录
				if asset.ObjectKey != "" {
					st := storage.GetStorageByDriver(asset.StorageType)
					if err := st.Delete(context.Background(), asset.ObjectKey); err != nil {
						fmt.Printf("删除存储对象失败: objectKey=%s, err=%v\n", asset.ObjectKey, err)
					}
				}
				if err := s.assetRepo.DeleteByID(entry.FileAssetID); err != nil {
					fmt.Printf("删除资产记录失败: assetID=%d, err=%v\n", entry.FileAssetID, err)
				}
			} else {
				// 引用计数减 1
				s.assetRepo.DecrementRefCount(entry.FileAssetID)
			}
		}
	}
}

// GetRecycleBinCount 获取回收站文件数量
func (s *RecycleBinService) GetRecycleBinCount(userID uint) (int64, error) {
	return s.fileRepo.CountRecycleBin(userID)
}
