package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

type ShareService struct {
	shareRepo   *repository.FileShareRepo
	fileRepo    *repository.FileRepo
	assetRepo   *repository.FileAssetRepo
	userRepo    *repository.UserRepo
}

func NewShareService() *ShareService {
	db := database.GetMySQL()
	return &ShareService{
		shareRepo:   repository.NewFileShareRepo(db),
		fileRepo:    repository.NewFileRepo(db),
		assetRepo:   repository.NewFileAssetRepo(db),
		userRepo:    repository.NewUserRepo(db),
	}
}

// generateShareCode 生成分享码（128 位，32 字符 hex）
func generateShareCode() string {
	b := make([]byte, 16) // 128 bits of entropy
	if _, err := rand.Read(b); err != nil {
		// 极小概率失败，fallback 到时间戳
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// CreateFileShare 创建文件分享
func (s *ShareService) CreateFileShare(userID, fileID uint, expireHours int, maxAccess int) (*model.FileShare, error) {
	// 验证文件归属
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil || entry.UserID != userID {
		return nil, fmt.Errorf("文件不存在或无权分享")
	}

	// 创建分享
	share := &model.FileShare{
		FileID:    fileID,
		ShareCode: generateShareCode(),
		UserID:    userID,
		MaxAccess: maxAccess,
		IsPublic:  true,
	}

	if expireHours > 0 {
		expireAt := time.Now().Add(time.Duration(expireHours) * time.Hour)
		share.ExpireAt = &expireAt
	}

	if err := s.shareRepo.Create(share); err != nil {
		return nil, err
	}

	return share, nil
}

// CreateFolderShare 创建文件夹分享
func (s *ShareService) CreateFolderShare(userID, folderID uint, expireHours int, maxAccess int) (*model.FileShare, error) {
	// 验证文件夹归属
	folder, err := s.fileRepo.GetFolderByID(folderID)
	if err != nil || folder.UserID != userID {
		return nil, fmt.Errorf("文件夹不存在或无权分享")
	}

	share := &model.FileShare{
		FolderID:  folderID,
		ShareCode: generateShareCode(),
		UserID:    userID,
		MaxAccess: maxAccess,
		IsPublic:  true,
	}

	if expireHours > 0 {
		expireAt := time.Now().Add(time.Duration(expireHours) * time.Hour)
		share.ExpireAt = &expireAt
	}

	if err := s.shareRepo.Create(share); err != nil {
		return nil, err
	}

	return share, nil
}

// GetShareInfo 获取分享信息
func (s *ShareService) GetShareInfo(code string) (map[string]interface{}, error) {
	share, err := s.shareRepo.GetByShareCode(code)
	if err != nil {
		return nil, fmt.Errorf("分享不存在")
	}

	// 检查状态
	if share.Status == 2 {
		return nil, fmt.Errorf("分享已过期")
	}
	if share.Status == 3 {
		return nil, fmt.Errorf("分享已被禁用")
	}

	// 检查过期
	if share.ExpireAt != nil && time.Now().After(*share.ExpireAt) {
		return nil, fmt.Errorf("分享已过期")
	}

	// 检查访问次数
	if share.MaxAccess > 0 && share.AccessCount >= share.MaxAccess {
		return nil, fmt.Errorf("分享已达到最大访问次数")
	}

	// 获取分享者信息
	user, err := s.userRepo.GetByID(share.UserID)
	if err != nil {
		return nil, fmt.Errorf("分享者信息获取失败")
	}

	result := map[string]interface{}{
		"shareCode":    share.ShareCode,
		"createdAt":    share.CreatedAt,
		"expireAt":     share.ExpireAt,
		"accessCount":  share.AccessCount,
		"maxAccess":    share.MaxAccess,
		"sharerId":     share.UserID,
		"sharerName":   user.Nickname,
		"sharerAvatar": user.Avatar,
	}

	// 文件分享
	if share.FileID > 0 {
		entry, err := s.fileRepo.GetEntryByID(share.FileID)
		if err != nil || entry == nil {
			return nil, fmt.Errorf("文件不存在")
		}
		asset, _ := s.assetRepo.GetByID(entry.FileAssetID)
		result["type"] = "file"
		result["fileName"] = entry.Name
		result["fileSize"] = entry.Size
		result["contentType"] = entry.ContentType
		result["fileId"] = entry.ID
		if asset != nil {
			result["objectKey"] = asset.ObjectKey
			result["storageType"] = asset.StorageType
		} else {
			result["storageType"] = "local"
		}
	}

	// 文件夹分享
	if share.FolderID > 0 {
		folder, err := s.fileRepo.GetFolderByID(share.FolderID)
		if err != nil || folder == nil {
			return nil, fmt.Errorf("文件夹不存在")
		}
		result["type"] = "folder"
		result["folderName"] = folder.Name
		result["folderId"] = folder.ID
	}

	return result, nil
}

// GetShareFolderFiles 获取分享文件夹内的文件列表
func (s *ShareService) GetShareFolderFiles(folderID uint) ([]map[string]interface{}, error) {
	// 获取文件夹内所有文件
	entries, err := s.fileRepo.ListEntriesByFolder(folderID)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败")
	}

	files := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		asset, _ := s.assetRepo.GetByID(entry.FileAssetID)
		file := map[string]interface{}{
			"fileId":      entry.ID,
			"fileName":    entry.Name,
			"fileSize":    entry.Size,
			"contentType": entry.ContentType,
			"createdAt":   entry.CreatedAt,
		}
		if asset != nil {
			file["objectKey"] = asset.ObjectKey
		}
		files = append(files, file)
	}

	return files, nil
}

// GetFileInFolder 获取文件夹内指定文件信息（验证文件归属）
func (s *ShareService) GetFileInFolder(folderID uint, fileID uint) (map[string]interface{}, error) {
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("文件不存在")
	}

	// 验证文件属于该文件夹（或其子文件夹）
	if !s.isFileInFolder(entry, folderID) {
		return nil, fmt.Errorf("文件不属于该文件夹")
	}

	asset, _ := s.assetRepo.GetByID(entry.FileAssetID)
	result := map[string]interface{}{
		"fileId":      entry.ID,
		"fileName":    entry.Name,
		"fileSize":    entry.Size,
		"contentType": entry.ContentType,
	}
	if asset != nil {
		result["objectKey"] = asset.ObjectKey
		result["storageType"] = asset.StorageType
	} else {
		result["storageType"] = "local"
	}

	return result, nil
}

// isFileInFolder 递归检查文件是否属于文件夹（包括子文件夹）
func (s *ShareService) isFileInFolder(entry *model.FileEntry, folderID uint) bool {
	// 文件直接在该文件夹下
	if entry.FolderID == folderID {
		return true
	}

	// 文件在根目录（folderID=0 表示根目录）
	if folderID == 0 && entry.FolderID == 0 {
		return true
	}

	// 检查文件是否在子文件夹下
	if entry.FolderID > 0 {
		folder, err := s.fileRepo.GetFolderByID(entry.FolderID)
		if err != nil {
			return false
		}
		// 递归检查父文件夹
		return s.isFolderInFolder(folder, folderID)
	}

	return false
}

// isFolderInFolder 递归检查文件夹是否是目标文件夹的子文件夹（带循环引用保护）
func (s *ShareService) isFolderInFolder(folder *model.FileFolder, targetFolderID uint) bool {
	visited := make(map[uint]bool)
	return s.isFolderInFolderHelper(folder, targetFolderID, visited)
}

func (s *ShareService) isFolderInFolderHelper(folder *model.FileFolder, targetFolderID uint, visited map[uint]bool) bool {
	if folder.ID == targetFolderID {
		return true
	}

	// 防止循环引用导致无限递归
	if visited[folder.ID] {
		return false
	}
	visited[folder.ID] = true

	// 递归检查父文件夹
	if folder.ParentID != nil {
		parent, err := s.fileRepo.GetFolderByID(*folder.ParentID)
		if err != nil {
			return false
		}
		return s.isFolderInFolderHelper(parent, targetFolderID, visited)
	}

	return false
}

// IncrementAccessCount 增加分享访问次数
func (s *ShareService) IncrementAccessCount(code string) {
	s.shareRepo.IncrementAccessCount(code)
}

// GetMyShares 获取我的分享列表
func (s *ShareService) GetMyShares(userID uint) ([]model.FileShare, error) {
	return s.shareRepo.GetByUserID(userID)
}

// DeleteShare 删除分享
// hasPermission=true 时可以删除任何分享，否则只能删除自己的
func (s *ShareService) DeleteShare(userID, shareID uint, hasPermission bool) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return err
	}
	if !hasPermission && share.UserID != userID {
		return fmt.Errorf("无权删除")
	}
	return s.shareRepo.Delete(shareID)
}

// ShareListItem 分享列表项
type ShareListItem struct {
	ID          uint       `json:"id"`
	ShareCode   string     `json:"shareCode"`
	ShareUrl    string     `json:"shareUrl"`
	Type        string     `json:"type"` // file or folder
	FileID      uint       `json:"fileId,omitempty"`
	FolderID    uint       `json:"folderId,omitempty"`
	FileName    string     `json:"fileName,omitempty"`
	FolderName  string     `json:"folderName,omitempty"`
	FileSize    int64      `json:"fileSize,omitempty"`
	ContentType string     `json:"contentType,omitempty"`
	Status      int        `json:"status"`
	ExpireAt    *time.Time `json:"expireAt,omitempty"`
	AccessCount int        `json:"accessCount"`
	MaxAccess   int        `json:"maxAccess"`
	AccessedAt  *time.Time `json:"accessedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	// 分享人信息
	UserID     uint   `json:"userId"`
	UserName   string `json:"userName,omitempty"`
	UserAvatar string `json:"userAvatar,omitempty"`
}

// GetUserShares 获取用户的分享列表（带文件信息）
func (s *ShareService) GetUserShares(userID uint, page, pageSize int, viewAll bool) ([]ShareListItem, int64, error) {
	shares, total, err := s.shareRepo.GetUserSharesWithFile(userID, page, pageSize, viewAll)
	if err != nil {
		return nil, 0, err
	}

	// 检查并更新过期状态
	s.shareRepo.CheckExpiredShares()

	// 收集所有需要查询的 ID（N+1 优化：批量查询）
	userIDSet := make(map[uint]bool)
	fileIDs := make([]uint, 0)
	folderIDs := make([]uint, 0)
	for _, share := range shares {
		userIDSet[share.UserID] = true
		if share.FileID > 0 {
			fileIDs = append(fileIDs, share.FileID)
		}
		if share.FolderID > 0 {
			folderIDs = append(folderIDs, share.FolderID)
		}
	}

	uidList := make([]uint, 0, len(userIDSet))
	for uid := range userIDSet {
		uidList = append(uidList, uid)
	}

	// 批量查询用户、文件、文件夹信息
	userMap, _ := s.userRepo.GetByIDs(uidList)
	entryMap, _ := s.fileRepo.GetEntriesByIDs(fileIDs)
	folderMap, _ := s.fileRepo.GetFoldersByIDs(folderIDs)

	items := make([]ShareListItem, 0, len(shares))
	for _, share := range shares {
		item := ShareListItem{
			ID:          share.ID,
			ShareCode:   share.ShareCode,
			ShareUrl:    fmt.Sprintf("/share/%s", share.ShareCode),
			Status:      share.Status,
			ExpireAt:    share.ExpireAt,
			AccessCount: share.AccessCount,
			MaxAccess:   share.MaxAccess,
			AccessedAt:  share.AccessedAt,
			CreatedAt:   share.CreatedAt,
			UserID:      share.UserID,
		}

		// 添加分享人信息
		if userMap != nil {
			if user, ok := userMap[share.UserID]; ok {
				item.UserName = user.Name
				item.UserAvatar = user.Avatar
			}
		}

		// 获取文件信息（从批量查询结果）
		if share.FileID > 0 && entryMap != nil {
			if entry, ok := entryMap[share.FileID]; ok {
				item.Type = "file"
				item.FileID = share.FileID
				item.FileName = entry.Name
				item.FileSize = entry.Size
				item.ContentType = entry.ContentType
			}
		}

		// 获取文件夹信息（从批量查询结果）
		if share.FolderID > 0 && folderMap != nil {
			if folder, ok := folderMap[share.FolderID]; ok {
				item.Type = "folder"
				item.FolderID = share.FolderID
				item.FolderName = folder.Name
			}
		}

		items = append(items, item)
	}

	return items, total, nil
}

// RenewShare 续签分享（延长过期时间）
func (s *ShareService) RenewShare(userID, shareID uint, expireHours int, hasPermission bool) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}
	if !hasPermission && share.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	var expireAt *time.Time
	if expireHours > 0 {
		t := time.Now().Add(time.Duration(expireHours) * time.Hour)
		expireAt = &t
	}

	return s.shareRepo.UpdateExpireAt(shareID, expireAt)
}

// ExpireShare 立即过期分享
func (s *ShareService) ExpireShare(userID, shareID uint, hasPermission bool) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}
	if !hasPermission && share.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	return s.shareRepo.UpdateStatus(shareID, 2) // 2 = 已过期
}

// DisableShare 禁用分享
func (s *ShareService) DisableShare(userID, shareID uint, hasPermission bool) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}
	if !hasPermission && share.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	return s.shareRepo.UpdateStatus(shareID, 3) // 3 = 已禁用
}

// EnableShare 启用分享
func (s *ShareService) EnableShare(userID, shareID uint, hasPermission bool) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}
	if !hasPermission && share.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	// 检查是否已过期
	if share.ExpireAt != nil && time.Now().After(*share.ExpireAt) {
		return fmt.Errorf("分享已过期，请先续签")
	}

	return s.shareRepo.UpdateStatus(shareID, 1) // 1 = 有效
}

// UpdateShareExpiry 修改分享到期时间
func (s *ShareService) UpdateShareExpiry(userID, shareID uint, expireAt *time.Time, hasPermission bool) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}
	if !hasPermission && share.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	// 如果新时间在未来，自动启用
	status := share.Status
	if expireAt != nil && expireAt.After(time.Now()) {
		status = 1
	}

	// 更新过期时间和状态
	err = s.shareRepo.UpdateExpireAt(shareID, expireAt)
	if err != nil {
		return err
	}

	if status != share.Status {
		return s.shareRepo.UpdateStatus(shareID, status)
	}

	return nil
}