package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// FileRepo 文件管理仓库
type FileRepo struct {
	db *gorm.DB
}

func NewFileRepo(db *gorm.DB) *FileRepo {
	return &FileRepo{db: db}
}

// --- 文件夹 ---

// CreateFolder 创建文件夹
func (r *FileRepo) CreateFolder(folder *model.FileFolder) error {
	return r.db.Create(folder).Error
}

// GetFolderByID 根据 ID 获取文件夹
func (r *FileRepo) GetFolderByID(id uint) (*model.FileFolder, error) {
	var folder model.FileFolder
	if err := r.db.Where("id = ?", id).First(&folder).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}

// GetFolderTree 获取用户的目录树
func (r *FileRepo) GetFolderTree(userID uint) ([]model.FileFolder, error) {
	var folders []model.FileFolder
	if err := r.db.Where("user_id = ?", userID).Order("path ASC").Find(&folders).Error; err != nil {
		return nil, err
	}
	return folders, nil
}

// UpdateFolder 更新文件夹
func (r *FileRepo) UpdateFolder(folder *model.FileFolder) error {
	return r.db.Save(folder).Error
}

// DeleteFolder 删除文件夹
func (r *FileRepo) DeleteFolder(id uint) error {
	return r.db.Delete(&model.FileFolder{}, id).Error
}

// DeleteFoldersByPrefix 根据路径前缀删除文件夹（递归删除）
func (r *FileRepo) DeleteFoldersByPrefix(pathPrefix string) error {
	return r.db.Where("path LIKE ?", pathPrefix+"%").Delete(&model.FileFolder{}).Error
}

// GetChildFolders 获取子文件夹
func (r *FileRepo) GetChildFolders(parentID uint) ([]model.FileFolder, error) {
	var folders []model.FileFolder
	if err := r.db.Where("parent_id = ?", parentID).Order("name ASC").Find(&folders).Error; err != nil {
		return nil, err
	}
	return folders, nil
}

// --- 文件条目 ---

// CreateEntry 创建文件条目
func (r *FileRepo) CreateEntry(entry *model.FileEntry) error {
	return r.db.Create(entry).Error
}

// GetEntryByID 根据 ID 获取文件条目
func (r *FileRepo) GetEntryByID(id uint) (*model.FileEntry, error) {
	var entry model.FileEntry
	if err := r.db.Where("id = ?", id).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// ListEntries 文件列表（分页+搜索）
func (r *FileRepo) ListEntries(userID uint, folderID uint, page, pageSize int, filters map[string]interface{}) ([]model.FileEntry, int64, error) {
	var entries []model.FileEntry
	var total int64

	query := r.db.Model(&model.FileEntry{}).Where("user_id = ?", userID)
	if folderID > 0 {
		query = query.Where("folder_id = ?", folderID)
	}
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		query = query.Where("name LIKE ?", "%"+escapeLike(keyword)+"%")
	}
	if contentType, ok := filters["contentType"].(string); ok && contentType != "" {
		query = query.Where("content_type LIKE ?", escapeLike(contentType)+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// UpdateEntry 更新文件条目
func (r *FileRepo) UpdateEntry(entry *model.FileEntry) error {
	return r.db.Save(entry).Error
}

// DeleteEntry 删除文件条目
func (r *FileRepo) DeleteEntry(id uint) error {
	return r.db.Delete(&model.FileEntry{}, id).Error
}

// DeleteEntriesByFolder 删除文件夹下所有文件
func (r *FileRepo) DeleteEntriesByFolder(folderID uint) error {
	return r.db.Where("folder_id = ?", folderID).Delete(&model.FileEntry{}).Error
}

// DeleteEntriesByFolderRecursive 递归删除文件夹下所有文件（通过文件夹ID列表）
func (r *FileRepo) DeleteEntriesByFolderRecursive(folderIDs []uint) error {
	if len(folderIDs) == 0 {
		return nil
	}
	return r.db.Where("folder_id IN ?", folderIDs).Delete(&model.FileEntry{}).Error
}

// ListEntriesByFolder 获取文件夹内所有文件（不分页，用于分享）
func (r *FileRepo) ListEntriesByFolder(folderID uint) ([]model.FileEntry, error) {
	var entries []model.FileEntry
	if err := r.db.Where("folder_id = ?", folderID).Order("created_at DESC").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

// ListEntriesByFolders 获取多个文件夹内所有文件（用于文件夹分享递归）
func (r *FileRepo) ListEntriesByFolders(folderIDs []uint) ([]model.FileEntry, error) {
	if len(folderIDs) == 0 {
		return []model.FileEntry{}, nil
	}
	var entries []model.FileEntry
	if err := r.db.Where("folder_id IN ?", folderIDs).Order("created_at DESC").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}
