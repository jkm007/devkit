package repository

import (
	"claude-manager/internal/model"
	"gorm.io/gorm"
)

// FileTagRepo 文件标签关联仓库
type FileTagRepo struct {
	db *gorm.DB
}

// NewFileTagRepo 创建文件标签关联仓库
func NewFileTagRepo(db *gorm.DB) *FileTagRepo {
	return &FileTagRepo{db: db}
}

// GetByFileID 获取文件的所有标签
func (r *FileTagRepo) GetByFileID(fileID uint) ([]model.FileTag, error) {
	var fileTags []model.FileTag
	err := r.db.Preload("Tag").Where("file_id = ?", fileID).Find(&fileTags).Error
	return fileTags, err
}

// GetByFileIDs 批量获取文件标签
func (r *FileTagRepo) GetByFileIDs(fileIDs []uint) (map[uint][]model.FileTag, error) {
	if len(fileIDs) == 0 {
		return nil, nil
	}
	var fileTags []model.FileTag
	if err := r.db.Preload("Tag").Where("file_id IN ?", fileIDs).Find(&fileTags).Error; err != nil {
		return nil, err
	}

	result := make(map[uint][]model.FileTag, len(fileIDs))
	for _, ft := range fileTags {
		result[ft.FileID] = append(result[ft.FileID], ft)
	}
	return result, nil
}

// GetByTagID 获取标签关联的所有文件ID
func (r *FileTagRepo) GetByTagID(tagID int64) ([]uint, error) {
	var fileIDs []uint
	err := r.db.Model(&model.FileTag{}).Where("tag_id = ?", tagID).Pluck("file_id", &fileIDs).Error
	return fileIDs, err
}

// Create 创建文件标签关联
func (r *FileTagRepo) Create(fileTag *model.FileTag) error {
	return r.db.Create(fileTag).Error
}

// BatchCreate 批量创建文件标签关联
func (r *FileTagRepo) BatchCreate(fileTags []model.FileTag) error {
	if len(fileTags) == 0 {
		return nil
	}
	return r.db.CreateInBatches(fileTags, 100).Error
}

// Delete 删除文件标签关联
func (r *FileTagRepo) Delete(fileID uint, tagID int64) error {
	return r.db.Where("file_id = ? AND tag_id = ?", fileID, tagID).Delete(&model.FileTag{}).Error
}

// DeleteByFileID 删除文件的所有标签
func (r *FileTagRepo) DeleteByFileID(fileID uint) error {
	return r.db.Where("file_id = ?", fileID).Delete(&model.FileTag{}).Error
}

// DeleteByFileIDAndSource 删除文件指定来源的标签
func (r *FileTagRepo) DeleteByFileIDAndSource(fileID uint, source string) error {
	return r.db.Where("file_id = ? AND source = ?", fileID, source).Delete(&model.FileTag{}).Error
}

// ReplaceFileTags 替换文件的标签（删除旧的，插入新的）
func (r *FileTagRepo) ReplaceFileTags(fileID uint, fileTags []model.FileTag) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧标签
		if err := tx.Where("file_id = ?", fileID).Delete(&model.FileTag{}).Error; err != nil {
			return err
		}
		// 插入新标签
		if len(fileTags) > 0 {
			if err := tx.CreateInBatches(fileTags, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CountByTagID 统计标签关联的文件数量
func (r *FileTagRepo) CountByTagID(tagID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.FileTag{}).Where("tag_id = ?", tagID).Count(&count).Error
	return count, err
}

// GetFilesByTagCondition 根据标签条件获取文件ID
func (r *FileTagRepo) GetFilesByTagCondition(tagKey, tagValue string) ([]uint, error) {
	var fileIDs []uint
	err := r.db.Raw(`
		SELECT DISTINCT ft.file_id
		FROM sys_file_tag ft
		JOIN sys_tag t ON ft.tag_id = t.id
		WHERE t.tag_key = ? AND t.tag_value = ? AND t.status = 1
	`, tagKey, tagValue).Scan(&fileIDs).Error
	return fileIDs, err
}
