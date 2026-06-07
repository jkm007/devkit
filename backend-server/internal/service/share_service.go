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

	// 增加访问次数
	s.shareRepo.IncrementAccessCount(code)

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

// isFolderInFolder 递归检查文件夹是否是目标文件夹的子文件夹
func (s *ShareService) isFolderInFolder(folder *model.FileFolder, targetFolderID uint) bool {
	if folder.ID == targetFolderID {
		return true
	}

	// 递归检查父文件夹
	if folder.ParentID != nil {
		parent, err := s.fileRepo.GetFolderByID(*folder.ParentID)
		if err != nil {
			return false
		}
		return s.isFolderInFolder(parent, targetFolderID)
	}

	return false
}

// GetMyShares 获取我的分享列表
func (s *ShareService) GetMyShares(userID uint) ([]model.FileShare, error) {
	return s.shareRepo.GetByUserID(userID)
}

// DeleteShare 删除分享
func (s *ShareService) DeleteShare(userID, shareID uint) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return err
	}
	if share.UserID != userID {
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
func (s *ShareService) GetUserShares(userID uint, page, pageSize int) ([]ShareListItem, int64, error) {
	shares, total, err := s.shareRepo.GetUserSharesWithFile(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 检查并更新过期状态
	s.shareRepo.CheckExpiredShares()

	// 收集所有需要查询的用户ID
	userIDs := make(map[uint]bool)
	for _, share := range shares {
		userIDs[share.UserID] = true
	}

	// 批量查询用户信息
	userMap := make(map[uint]*model.User)
	for uid := range userIDs {
		user, err := s.userRepo.GetByID(uid)
		if err == nil && user != nil {
			userMap[uid] = user
		}
	}

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
		if user, ok := userMap[share.UserID]; ok {
			item.UserName = user.Name
			item.UserAvatar = user.Avatar
		}

		// 获取文件信息
		if share.FileID > 0 {
			entry, err := s.fileRepo.GetEntryByID(share.FileID)
			if err == nil && entry != nil {
				item.Type = "file"
				item.FileID = share.FileID
				item.FileName = entry.Name
				item.FileSize = entry.Size
				item.ContentType = entry.ContentType
			}
		}

		// 获取文件夹信息
		if share.FolderID > 0 {
			folder, err := s.fileRepo.GetFolderByID(share.FolderID)
			if err == nil && folder != nil {
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
func (s *ShareService) RenewShare(userID, shareID uint, expireHours int) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}
	if share.UserID != userID {
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
func (s *ShareService) ExpireShare(userID, shareID uint) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}
	if share.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	return s.shareRepo.UpdateStatus(shareID, 2) // 2 = 已过期
}

// DisableShare 禁用分享
func (s *ShareService) DisableShare(userID, shareID uint) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}
	if share.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	return s.shareRepo.UpdateStatus(shareID, 3) // 3 = 已禁用
}

// EnableShare 启用分享
func (s *ShareService) EnableShare(userID, shareID uint) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}
	if share.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	// 检查是否已过期
	if share.ExpireAt != nil && time.Now().After(*share.ExpireAt) {
		return fmt.Errorf("分享已过期，请先续签")
	}

	return s.shareRepo.UpdateStatus(shareID, 1) // 1 = 有效
}

// UpdateShareExpiry 修改分享到期时间
func (s *ShareService) UpdateShareExpiry(userID, shareID uint, expireAt *time.Time) error {
	share, err := s.shareRepo.GetByID(shareID)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}
	if share.UserID != userID {
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