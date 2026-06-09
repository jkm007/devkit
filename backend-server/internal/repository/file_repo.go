package repository

import (
	"time"

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

// GetFoldersByIDs 批量获取文件夹（N+1 优化）
func (r *FileRepo) GetFoldersByIDs(ids []uint) (map[uint]*model.FileFolder, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var folders []model.FileFolder
	if err := r.db.Where("id IN ?", ids).Find(&folders).Error; err != nil {
		return nil, err
	}
	result := make(map[uint]*model.FileFolder, len(folders))
	for i := range folders {
		result[folders[i].ID] = &folders[i]
	}
	return result, nil
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

// GetEntriesByIDs 批量获取文件条目（N+1 优化）
func (r *FileRepo) GetEntriesByIDs(ids []uint) (map[uint]*model.FileEntry, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var entries []model.FileEntry
	if err := r.db.Where("id IN ?", ids).Find(&entries).Error; err != nil {
		return nil, err
	}
	result := make(map[uint]*model.FileEntry, len(entries))
	for i := range entries {
		result[entries[i].ID] = &entries[i]
	}
	return result, nil
}

// ListEntries 文件列表（分页+搜索）
// userID 为 0 时查看所有文件
func (r *FileRepo) ListEntries(userID uint, folderID uint, page, pageSize int, filters map[string]interface{}) ([]model.FileEntry, int64, error) {
	var entries []model.FileEntry
	var total int64

	query := r.db.Model(&model.FileEntry{}).Where("deleted_at IS NULL")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if folderID > 0 {
		query = query.Where("folder_id = ?", folderID)
	}
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		query = query.Where("name LIKE ?", "%"+escapeLike(keyword)+"%")
	}
	if contentType, ok := filters["contentType"].(string); ok && contentType != "" {
		query = query.Where("content_type LIKE ?", escapeLike(contentType)+"%")
	}
	// 标签筛选
	if tagFileIDs, ok := filters["tagFileIDs"].([]uint); ok && len(tagFileIDs) > 0 {
		query = query.Where("id IN ?", tagFileIDs)
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

// --- 回收站 ---

// SoftDeleteEntry 软删除文件（移入回收站）
func (r *FileRepo) SoftDeleteEntry(id uint, expireAt time.Time) error {
	now := time.Now()
	return r.db.Model(&model.FileEntry{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at":        now,
		"recycle_expire_at": expireAt,
	}).Error
}

// SoftDeleteEntries 批量软删除文件
func (r *FileRepo) SoftDeleteEntries(ids []uint, expireAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	return r.db.Model(&model.FileEntry{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"deleted_at":        now,
		"recycle_expire_at": expireAt,
	}).Error
}

// RestoreEntry 从回收站恢复文件
func (r *FileRepo) RestoreEntry(id uint) error {
	return r.db.Model(&model.FileEntry{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at":        nil,
		"recycle_expire_at": nil,
	}).Error
}

// RestoreEntries 批量恢复文件
func (r *FileRepo) RestoreEntries(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&model.FileEntry{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"deleted_at":        nil,
		"recycle_expire_at": nil,
	}).Error
}

// ListRecycleBin 获取回收站文件列表
func (r *FileRepo) ListRecycleBin(userID uint, page, pageSize int) ([]model.FileEntry, int64, error) {
	var entries []model.FileEntry
	var total int64

	query := r.db.Model(&model.FileEntry{}).Where("deleted_at IS NOT NULL")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// GetExpiredRecycleBinEntries 获取已过期的回收站文件
func (r *FileRepo) GetExpiredRecycleBinEntries() ([]model.FileEntry, error) {
	var entries []model.FileEntry
	now := time.Now()
	err := r.db.Where("deleted_at IS NOT NULL AND recycle_expire_at IS NOT NULL AND recycle_expire_at <= ?", now).Find(&entries).Error
	return entries, err
}

// HardDeleteEntry 永久删除文件条目
func (r *FileRepo) HardDeleteEntry(id uint) error {
	return r.db.Unscoped().Delete(&model.FileEntry{}, id).Error
}

// HardDeleteEntries 批量永久删除文件条目
func (r *FileRepo) HardDeleteEntries(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Unscoped().Where("id IN ?", ids).Delete(&model.FileEntry{}).Error
}

// EmptyRecycleBin 清空用户的回收站
func (r *FileRepo) EmptyRecycleBin(userID uint) error {
	return r.db.Unscoped().Where("deleted_at IS NOT NULL AND user_id = ?", userID).Delete(&model.FileEntry{}).Error
}

// CountRecycleBin 统计回收站文件数量
func (r *FileRepo) CountRecycleBin(userID uint) (int64, error) {
	var count int64
	query := r.db.Model(&model.FileEntry{}).Where("deleted_at IS NOT NULL")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Count(&count).Error
	return count, err
}
