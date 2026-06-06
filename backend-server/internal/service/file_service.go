package service

import (
	"errors"
	"fmt"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

// FileService 文件管理服务
type FileService struct {
	fileRepo  *repository.FileRepo
	assetRepo *repository.FileAssetRepo
}

func NewFileService() *FileService {
	db := database.GetMySQL()
	return &FileService{
		fileRepo:  repository.NewFileRepo(db),
		assetRepo: repository.NewFileAssetRepo(db),
	}
}

// --- 文件夹 ---

// CreateFolder 创建文件夹
func (s *FileService) CreateFolder(userID uint, name string, parentID *uint) (*model.FileFolder, error) {
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

	oldPath := folder.Path
	newPath := folder.Path[:len(folder.Path)-len(folder.Name)] + newName

	folder.Name = newName
	folder.Path = newPath
	if err := s.fileRepo.UpdateFolder(folder); err != nil {
		return err
	}

	// 更新子文件夹路径
	s.updateChildPaths(folderID, oldPath, newPath)

	return nil
}

func (s *FileService) updateChildPaths(parentID uint, oldPrefix, newPrefix string) {
	children, _ := s.fileRepo.GetChildFolders(parentID)
	for _, child := range children {
		child.Path = newPrefix + child.Path[len(oldPrefix):]
		s.fileRepo.UpdateFolder(&child)
		s.updateChildPaths(child.ID, oldPrefix, newPrefix)
	}
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
	s.collectChildFolderIDs(folderID, &allIDs)

	// 删除所有文件条目
	s.fileRepo.DeleteEntriesByFolderRecursive(allIDs)

	// 删除所有文件夹
	for _, id := range allIDs {
		s.fileRepo.DeleteFolder(id)
	}

	return nil
}

func (s *FileService) collectChildFolderIDs(parentID uint, ids *[]uint) {
	children, _ := s.fileRepo.GetChildFolders(parentID)
	for _, child := range children {
		*ids = append(*ids, child.ID)
		s.collectChildFolderIDs(child.ID, ids)
	}
}

// --- 文件条目 ---

// ListFiles 文件列表
type ListFilesRequest struct {
	FolderID    uint   `form:"folderId"`
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Keyword     string `form:"keyword"`
	ContentType string `form:"contentType"`
}

func (s *FileService) ListFiles(userID uint, req *ListFilesRequest) ([]model.FileEntry, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filters := map[string]interface{}{
		"keyword":     req.Keyword,
		"contentType": req.ContentType,
	}

	return s.fileRepo.ListEntries(userID, req.FolderID, req.Page, req.PageSize, filters)
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

// DeleteFile 删除文件
func (s *FileService) DeleteFile(userID uint, fileID uint) error {
	entry, err := s.fileRepo.GetEntryByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}
	if entry.UserID != userID {
		return fmt.Errorf("无权操作")
	}

	// 减少文件资产引用计数
	if entry.FileAssetID > 0 {
		s.assetRepo.DecrementRefCount(entry.FileAssetID)
	}

	return s.fileRepo.DeleteEntry(fileID)
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
