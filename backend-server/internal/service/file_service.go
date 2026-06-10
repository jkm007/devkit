package service

import (
	"context"
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

	"gorm.io/gorm"
)

// intersectFileIDs 计算两个文件ID列表的交集
// 当任一输入为 nil 时，交集结果为空（nil 表示"没有匹配"）
func intersectFileIDs(a, b []uint) []uint {
	if a == nil || b == nil {
		return nil
	}
	set := make(map[uint]bool, len(a))
	for _, id := range a {
		set[id] = true
	}
	result := make([]uint, 0)
	for _, id := range b {
		if set[id] {
			result = append(result, id)
		}
	}
	return result
}

// FileService 文件管理服务
type FileService struct {
	fileRepo    *repository.FileRepo
	assetRepo   *repository.FileAssetRepo
	userRepo    *repository.UserRepo
	tagRepo     *repository.TagRepo
	fileTagRepo *repository.FileTagRepo
	shareRepo   *repository.FileShareRepo
}

func NewFileService() *FileService {
	db := database.GetMySQL()
	return &FileService{
		fileRepo:    repository.NewFileRepo(db),
		assetRepo:   repository.NewFileAssetRepo(db),
		userRepo:    repository.NewUserRepo(db),
		tagRepo:     repository.NewTagRepo(db),
		fileTagRepo: repository.NewFileTagRepo(db),
		shareRepo:   repository.NewFileShareRepo(db),
	}
}

// --- 文件夹 ---

// CreateAvatarFolder 创建头像文件夹
func (s *FileService) CreateAvatarFolder(userID uint) (*model.FileFolder, error) {
	// 检查是否已存在
	folders, err := s.fileRepo.GetFolderTree(userID)
	if err == nil {
		for _, f := range folders {
			if f.Name == "Avatars" && f.Type == "avatar" {
				return &f, nil
			}
		}
	}

	// 创建头像文件夹
	folder := &model.FileFolder{
		Name:   "Avatars",
		Path:   "/Avatars",
		Type:   "avatar",
		UserID: userID,
	}
	if err := s.fileRepo.CreateFolder(folder); err != nil {
		return nil, err
	}
	return folder, nil
}

// CreateFolder 创建文件夹
func (s *FileService) CreateFolder(userID uint, name string, parentID *uint) (*model.FileFolder, error) {
	// 过滤非法字符（防止路径遍历）
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "..", "")
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("文件夹名称不能为空")
	}

	path := "/" + name
	if parentID != nil {
		parent, err := s.fileRepo.GetFolderByID(*parentID)
		if err != nil {
			return nil, fmt.Errorf("父文件夹不存在")
		}
		if parent.UserID != userID {
			return nil, fmt.Errorf("无权访问父文件夹")
		}
		path = parent.Path + "/" + name
	}

	folder := &model.FileFolder{
		Name:     name,
		ParentID: parentID,
		Path:     path,
		UserID:   userID,
	}
	if err := s.fileRepo.CreateFolder(folder); err != nil {
		return nil, err
	}
	return folder, nil
}

// FolderTreeNode 目录树节点
type FolderTreeNode struct {
	ID       uint              `json:"id"`
	Name     string            `json:"name"`
	ParentID *uint             `json:"parentId"`
	Type     string            `json:"type"`
	Children []*FolderTreeNode `json:"children,omitempty"`
}

// GetFolderTree 获取目录树
func (s *FileService) GetFolderTree(userID uint) ([]*FolderTreeNode, error) {
	folders, err := s.fileRepo.GetFolderTree(userID)
	if err != nil {
		return nil, err
	}

	// 构建树
	nodeMap := make(map[uint]*FolderTreeNode)
	var roots []*FolderTreeNode

	for _, f := range folders {
		node := &FolderTreeNode{
			ID:       f.ID,
			Name:     f.Name,
			ParentID: f.ParentID,
				Type:     f.Type,
		}
		nodeMap[f.ID] = node
	}

	for _, f := range folders {
		node := nodeMap[f.ID]
		if f.ParentID == nil {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[*f.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	return roots, nil
}

// RenameFolder 重命名文件夹
func (s *FileService) RenameFolder(userID uint, folderID uint, newName string) error {
	folder, err := s.fileRepo.GetFolderByID(folderID)
	if err != nil {
		return fmt.Errorf("文件夹不存在")
	}
	if folder.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	// 过滤非法字符
	newName = strings.ReplaceAll(newName, "/", "")
	newName = strings.ReplaceAll(newName, "..", "")
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("文件夹名称不能为空")
	}

	oldPath := folder.Path
	newPath := filepath.Dir(folder.Path) + "/" + newName

	folder.Name = newName
	folder.Path = newPath
	if err := s.fileRepo.UpdateFolder(folder); err != nil {
		return err
	}

	// 更新子文件夹路径
	if err := s.updateChildPaths(folderID, oldPath, newPath); err != nil {
		return fmt.Errorf("更新子文件夹路径失败: %w", err)
	}

	return nil
}

func (s *FileService) updateChildPaths(parentID uint, oldPrefix, newPrefix string) error {
	// 迭代替代递归，避免深层嵌套栈溢出
	queue := []uint{parentID}
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		children, err := s.fileRepo.GetChildFolders(currentID)
		if err != nil {
			return err
		}
		for i := range children {
			child := &children[i]
			if len(child.Path) > len(oldPrefix) {
				child.Path = newPrefix + child.Path[len(oldPrefix):]
			}
			if err := s.fileRepo.UpdateFolder(child); err != nil {
				return err
			}
			queue = append(queue, child.ID)
		}
	}
	return nil
}

// DeleteFolder 删除文件夹（递归）
func (s *FileService) DeleteFolder(userID uint, folderID uint) error {
	folder, err := s.fileRepo.GetFolderByID(folderID)
	if err != nil {
		return fmt.Errorf("文件夹不存在")
	}
	if folder.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	// 收集所有子文件夹ID
	allIDs := []uint{folderID}
	if err := s.collectChildFolderIDs(folderID, &allIDs); err != nil {
		return fmt.Errorf("获取子文件夹失败: %w", err)
	}

	// 先获取所有待删除文件条目的 FileAssetID，用于递减引用计数
	entries, err := s.fileRepo.ListEntriesByFolders(allIDs)
	if err != nil {
		return fmt.Errorf("获取文件列表失败: %w", err)
	}

	// 处理每个文件：删除分享、标签、存储对象
	for _, entry := range entries {
		// 删除文件的分享记录
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
					if err := s.assetRepo.DecrementRefCount(entry.FileAssetID); err != nil {
						fmt.Printf("递减引用计数失败: fileAssetID=%d, err=%v\n", entry.FileAssetID, err)
					}
				}
			}
		}
	}

	// 删除所有文件条目
	if err := s.fileRepo.DeleteEntriesByFolderRecursive(allIDs); err != nil {
		return fmt.Errorf("删除文件条目失败: %w", err)
	}

	// 删除所有文件夹
	for _, id := range allIDs {
		if err := s.fileRepo.DeleteFolder(id); err != nil {
			return fmt.Errorf("删除文件夹失败: %w", err)
		}
	}

	return nil
}

func (s *FileService) collectChildFolderIDs(parentID uint, ids *[]uint) error {
	children, err := s.fileRepo.GetChildFolders(parentID)
	if err != nil {
		return err
	}
	for _, child := range children {
		*ids = append(*ids, child.ID)
		if err := s.collectChildFolderIDs(child.ID, ids); err != nil {
			return err
		}
	}
	return nil
}

// --- 文件条目 ---

// TagInfo 标签信息
type TagInfo struct {
	ID    int64  `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// FileEntryWithURL 文件条目带预览URL
type FileEntryWithURL struct {
	ID             uint      `json:"id"`
	FolderID       uint      `json:"folderId"`
	Name           string    `json:"name"`
	Size           int64     `json:"size"`
	ContentType    string    `json:"contentType"`
	StorageType    string    `json:"storageType"`
	UserID         uint      `json:"userId"`
	CreatedAt      string    `json:"createdAt"`
	UpdatedAt      string    `json:"updatedAt"`
	PreviewURL     string    `json:"previewUrl,omitempty"`
	UploaderName   string    `json:"uploaderName"`
	UploaderAvatar string    `json:"uploaderAvatar"`
	Tags           []TagInfo `json:"tags,omitempty"`
}

// ListFiles 文件列表
type ListFilesRequest struct {
	FolderID    uint   `form:"folderId"`
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Keyword     string `form:"keyword"`
	ContentType string `form:"contentType"`
	TagKeys     string `form:"tagKeys"` // 标签筛选，格式: "type:image,source:user"
}

func (s *FileService) ListFiles(userID uint, req *ListFilesRequest) ([]FileEntryWithURL, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 处理标签筛选
	var tagFilteredFileIDs []uint
	if req.TagKeys != "" {
		tagPairs := strings.Split(req.TagKeys, ",")
		for _, pair := range tagPairs {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key, value := parts[0], parts[1]
			fileIDs, err := s.fileTagRepo.GetFilesByTagCondition(key, value)
			if err != nil {
				continue
			}
			if tagFilteredFileIDs == nil {
				tagFilteredFileIDs = fileIDs
			} else {
				// 取交集（AND 逻辑）
				tagFilteredFileIDs = intersectFileIDs(tagFilteredFileIDs, fileIDs)
			}
		}
		// 如果筛选后没有结果，直接返回空
		if len(tagFilteredFileIDs) == 0 && req.TagKeys != "" {
			return []FileEntryWithURL{}, 0, nil
		}
	}

	filters := map[string]interface{}{
		"keyword":       req.Keyword,
		"contentType":   req.ContentType,
		"tagFileIDs":    tagFilteredFileIDs,
	}

	entries, total, err := s.fileRepo.ListEntries(userID, req.FolderID, req.Page, req.PageSize, filters)
	if err != nil {
		return nil, 0, err
	}

	// 收集所有用户ID和资产ID
	userIDs := make(map[uint]bool)
	assetIDs := make([]uint, 0)
	for _, entry := range entries {
		userIDs[entry.UserID] = true
		if entry.FileAssetID > 0 {
			assetIDs = append(assetIDs, entry.FileAssetID)
		}
	}

	// 批量查询用户信息
	userMap := make(map[uint]*model.User)
	for uid := range userIDs {
		user, err := s.userRepo.GetByID(uid)
		if err == nil {
			userMap[uid] = user
		}
	}

	// 批量查询资产信息
	assetMap := make(map[uint]*model.FileAsset)
	if len(assetIDs) > 0 {
		var assets []model.FileAsset
		db := database.GetMySQL()
		if err := db.Where("id IN ?", assetIDs).Find(&assets).Error; err == nil {
			for i := range assets {
				assetMap[assets[i].ID] = &assets[i]
			}
		}
	}

	// 批量查询文件标签
	fileIDs := make([]uint, 0, len(entries))
	for _, entry := range entries {
		fileIDs = append(fileIDs, entry.ID)
	}
	fileTagsMap, _ := s.fileTagRepo.GetByFileIDs(fileIDs)

	// 批量查询哪些文件有活跃分享（用于动态添加"公开"标签）
	sharedFileIDs, _ := s.shareRepo.GetSharedFileIDs(fileIDs)

	// 转换为带 URL 的结构
	result := make([]FileEntryWithURL, len(entries))
	for i, entry := range entries {
		result[i] = FileEntryWithURL{
			ID:          entry.ID,
			FolderID:    entry.FolderID,
			Name:        entry.Name,
			Size:        entry.Size,
			ContentType: entry.ContentType,
			UserID:      entry.UserID,
			CreatedAt:   entry.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   entry.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// 获取上传者信息
		if user, ok := userMap[entry.UserID]; ok {
			result[i].UploaderName = user.Nickname
			result[i].UploaderAvatar = user.Avatar
		}

		// 获取预览 URL 和存储类型
		if asset, ok := assetMap[entry.FileAssetID]; ok {
			if asset.StorageType == "local" {
				// 本地存储：走代理（文件在服务器上）
				result[i].PreviewURL = "/files/" + strconv.FormatUint(uint64(entry.ID), 10) + "/view"
			} else {
				// 云存储：使用 direct-url 接口获取 presigned URL（节省带宽）
				result[i].PreviewURL = "/files/" + strconv.FormatUint(uint64(entry.ID), 10) + "/direct-url"
			}
			result[i].StorageType = asset.StorageType
		}

		// 获取标签（过滤掉存储的"公开"标签，改为动态判断）
		if fileTags, ok := fileTagsMap[entry.ID]; ok {
			tags := make([]TagInfo, 0, len(fileTags))
			for _, ft := range fileTags {
				if ft.Tag != nil {
					// 跳过存储的"公开"标签，后面根据分享状态动态添加
					if ft.Tag.TagKey == "sensitivity" && ft.Tag.TagValue == "public" {
						continue
					}
					tags = append(tags, TagInfo{
						ID:    ft.Tag.ID,
						Key:   ft.Tag.TagKey,
						Value: ft.Tag.TagValue,
						Name:  ft.Tag.TagName,
						Icon:  ft.Tag.Icon,
						Color: ft.Tag.Color,
					})
				}
			}
			result[i].Tags = tags
		}

		// 根据分享状态动态添加"公开"标签
		if sharedFileIDs != nil && sharedFileIDs[entry.ID] {
			publicTag := TagInfo{
				ID:    10,
				Key:   "sensitivity",
				Value: "public",
				Name:  "公开",
				Icon:  "🔓",
				Color: "#52c41a",
			}
			if result[i].Tags == nil {
				result[i].Tags = []TagInfo{publicTag}
			} else {
				result[i].Tags = append(result[i].Tags, publicTag)
			}
		}
	}

	return result, total, nil
}

// MoveFile 移动文件到目标文件夹
func (s *FileService) MoveFile(userID uint, fileID uint, targetFolderID uint) error {
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}
	if entry.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	if targetFolderID > 0 {
		folder, err := s.fileRepo.GetFolderByID(targetFolderID)
		if err != nil || folder.UserID != userID {
			return fmt.Errorf("目标文件夹不存在")
		}
	}

	entry.FolderID = targetFolderID
	return s.fileRepo.UpdateEntry(entry)
}

// CheckFileShares 检查文件是否有活跃分享
func (s *FileService) CheckFileShares(fileID uint) (int64, error) {
	return s.shareRepo.CountActiveByFileID(fileID)
}

// DeleteFile 删除文件（移入回收站）
// hasPermission=true 时可以删除任何文件，否则只能删除自己的
func (s *FileService) DeleteFile(userID uint, fileID uint, hasPermission bool) error {
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}
	if !hasPermission && entry.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	// 软删除：移入回收站，7天后过期
	expireAt := time.Now().Add(7 * 24 * time.Hour)

	db := database.GetMySQL()
	return db.Transaction(func(tx *gorm.DB) error {
		// 软删除文件
		now := time.Now()
		if err := tx.Model(&model.FileEntry{}).Where("id = ?", fileID).Updates(map[string]interface{}{
			"deleted_at":        now,
			"recycle_expire_at": expireAt,
		}).Error; err != nil {
			return fmt.Errorf("软删除文件失败: %w", err)
		}

		// 删除该文件的所有分享链接
		if err := tx.Where("file_id = ?", fileID).Delete(&model.FileShare{}).Error; err != nil {
			return fmt.Errorf("删除分享链接失败: %w", err)
		}

		return nil
	})
}

// BatchDeleteFiles 批量删除文件（移入回收站）
func (s *FileService) BatchDeleteFiles(userID uint, fileIDs []uint, hasPermission bool) (int, []string) {
	deleted := 0
	errList := []string{}

	for _, fileID := range fileIDs {
		err := s.DeleteFile(userID, fileID, hasPermission)
		if err != nil {
			errList = append(errList, fmt.Sprintf("文件 %d: %s", fileID, err.Error()))
		} else {
			deleted++
		}
	}

	return deleted, errList
}

// BatchMoveFiles 批量移动文件
func (s *FileService) BatchMoveFiles(userID uint, fileIDs []uint, targetFolderID uint) (int, []string) {
	moved := 0
	errors := []string{}

	// 验证目标文件夹
	if targetFolderID > 0 {
		folder, err := s.fileRepo.GetFolderByID(targetFolderID)
		if err != nil || folder.UserID != userID {
			return 0, []string{"目标文件夹不存在或无权访问"}
		}
	}

	for _, fileID := range fileIDs {
		err := s.MoveFile(userID, fileID, targetFolderID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("文件 %d: %s", fileID, err.Error()))
		} else {
			moved++
		}
	}

	return moved, errors
}

// CreateFileEntry 创建文件条目（上传完成后调用）
func (s *FileService) CreateFileEntry(userID uint, folderID uint, assetID uint, name string, size int64, contentType string) error {
	// 验证文件夹归属
	if folderID > 0 {
		folder, err := s.fileRepo.GetFolderByID(folderID)
		if err != nil {
			return fmt.Errorf("目标文件夹不存在")
		}
		if folder.UserID != userID {
			return fmt.Errorf("无权写入目标文件夹")
		}
	}

	entry := &model.FileEntry{
		FolderID:    folderID,
		FileAssetID: assetID,
		Name:        name,
		Size:        size,
		ContentType: contentType,
		UserID:      userID,
	}
	return s.fileRepo.CreateEntry(entry)
}

// GetFileAsset 获取文件资产（用于下载/播放）
func (s *FileService) GetFileAsset(fileEntryID uint) (*model.FileAsset, *model.FileEntry, error) {
	entry, err := s.fileRepo.GetEntryByID(fileEntryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("文件不存在")
		}
		return nil, nil, err
	}

	asset, err := s.assetRepo.GetByID(entry.FileAssetID)
	if err != nil {
		return nil, nil, fmt.Errorf("文件资产不存在")
	}

	return asset, entry, nil
}

// GetFileEntry 获取文件条目（验证归属）
func (s *FileService) GetFileEntry(userID uint, fileID uint) (*model.FileEntry, error) {
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("文件不存在")
	}
	if entry.UserID != userID {
		return nil, fmt.Errorf("无权访问")
	}
	return entry, nil
}

// GetAssetByID 获取文件资产
func (s *FileService) GetAssetByID(assetID uint) (*model.FileAsset, error) {
	return s.assetRepo.GetByID(assetID)
}

// GetFileTags 获取文件的标签
func (s *FileService) GetFileTags(fileID uint) ([]TagInfo, error) {
	fileTags, err := s.fileTagRepo.GetByFileID(fileID)
	if err != nil {
		return nil, err
	}

	tags := make([]TagInfo, 0, len(fileTags))
	for _, ft := range fileTags {
		if ft.Tag != nil {
			tags = append(tags, TagInfo{
				ID:    ft.Tag.ID,
				Key:   ft.Tag.TagKey,
				Value: ft.Tag.TagValue,
				Name:  ft.Tag.TagName,
				Icon:  ft.Tag.Icon,
				Color: ft.Tag.Color,
			})
		}
	}
	return tags, nil
}

// AddFileTag 添加文件标签
func (s *FileService) AddFileTag(fileID uint, userID uint, tagID int64) error {
	// 验证文件归属
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}
	if entry.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	// 验证标签存在
	_, err = s.tagRepo.GetByID(tagID)
	if err != nil {
		return fmt.Errorf("标签不存在")
	}

	fileTag := &model.FileTag{
		FileID: fileID,
		TagID:  tagID,
		Source: "manual",
	}
	return s.fileTagRepo.Create(fileTag)
}

// RemoveFileTag 移除文件标签
func (s *FileService) RemoveFileTag(fileID uint, userID uint, tagID int64) error {
	// 验证文件归属
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}
	if entry.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	return s.fileTagRepo.Delete(fileID, tagID)
}

// ReplaceFileTags 替换文件的标签
func (s *FileService) ReplaceFileTags(fileID uint, userID uint, tagIDs []int64) error {
	// 验证文件归属
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}
	if entry.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	fileTags := make([]model.FileTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		fileTags = append(fileTags, model.FileTag{
			FileID: fileID,
			TagID:  tagID,
			Source: "manual",
		})
	}
	return s.fileTagRepo.ReplaceFileTags(fileID, fileTags)
}

// GetPresignedURL 获取预签名 URL
// 如果 expires <= 0，使用存储配置的默认过期时间
func (s *FileService) GetPresignedURL(fileID uint, expires int64) (string, error) {
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return "", fmt.Errorf("文件不存在")
	}

	asset, err := s.assetRepo.GetByID(entry.FileAssetID)
	if err != nil {
		return "", fmt.Errorf("文件资产不存在")
	}

	if asset.StorageType == "local" {
		return "", fmt.Errorf("本地存储不支持预签名 URL")
	}

	// 如果未指定过期时间，使用存储配置的默认值
	if expires <= 0 {
		var config model.StorageConfig
		db := database.GetMySQL()
		if db != nil {
			if err := db.Where("driver = ? AND is_default = 1 AND status = 1", asset.StorageType).First(&config).Error; err == nil && config.PresignedURLExpiry > 0 {
				expires = int64(config.PresignedURLExpiry)
			} else {
				expires = 3600
			}
		} else {
			expires = 3600
		}
	}

	st := storage.GetStorageByDriver(asset.StorageType)
	return st.GetPresignedURL(context.Background(), asset.ObjectKey, expires)
}
