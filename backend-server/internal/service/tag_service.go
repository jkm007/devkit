package service

import (
	"claude-manager/internal/model"
	"claude-manager/internal/repository"
	"claude-manager/pkg/storage"
	"fmt"
)

// TagService 标签服务
type TagService struct {
	tagRepo     *repository.TagRepo
	fileTagRepo *repository.FileTagRepo
}

// NewTagService 创建标签服务
func NewTagService(tagRepo *repository.TagRepo, fileTagRepo *repository.FileTagRepo) *TagService {
	return &TagService{
		tagRepo:     tagRepo,
		fileTagRepo: fileTagRepo,
	}
}

// GetAllTags 获取所有标签
func (s *TagService) GetAllTags() ([]model.Tag, error) {
	return s.tagRepo.GetAll()
}

// GetGroupedTags 获取按 key 分组的标签
func (s *TagService) GetGroupedTags() (map[string][]model.Tag, error) {
	return s.tagRepo.GetGroupedTags()
}

// GetTagsByKey 获取指定键的所有标签值
func (s *TagService) GetTagsByKey(tagKey string) ([]model.Tag, error) {
	return s.tagRepo.GetByKey(tagKey)
}

// GetTagByID 根据ID获取标签
func (s *TagService) GetTagByID(id int64) (*model.Tag, error) {
	return s.tagRepo.GetByID(id)
}

// CreateTag 创建标签
func (s *TagService) CreateTag(tag *model.Tag) error {
	// 检查是否已存在
	existing, _ := s.tagRepo.GetByKeyValue(tag.TagKey, tag.TagValue)
	if existing != nil {
		return fmt.Errorf("标签 %s:%s 已存在", tag.TagKey, tag.TagValue)
	}
	return s.tagRepo.Create(tag)
}

// UpdateTag 更新标签
func (s *TagService) UpdateTag(tag *model.Tag) error {
	// 检查是否系统内置标签
	existing, err := s.tagRepo.GetByID(tag.ID)
	if err != nil {
		return err
	}
	if existing.IsSystem {
		return fmt.Errorf("系统内置标签不允许修改")
	}
	return s.tagRepo.Update(tag)
}

// DeleteTag 删除标签
func (s *TagService) DeleteTag(id int64) error {
	// 检查是否系统内置标签
	existing, err := s.tagRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing.IsSystem {
		return fmt.Errorf("系统内置标签不允许删除")
	}

	// 检查是否有文件关联
	count, err := s.fileTagRepo.CountByTagID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该标签已关联 %d 个文件，无法删除", count)
	}

	return s.tagRepo.Delete(id)
}

// GetUsageStats 获取标签使用统计
func (s *TagService) GetUsageStats() ([]repository.TagUsageStat, error) {
	return s.tagRepo.GetUsageStats()
}

// GetFileTags 获取文件的标签
func (s *TagService) GetFileTags(fileID uint) ([]model.FileTag, error) {
	return s.fileTagRepo.GetByFileID(fileID)
}

// GetFileTagsMap 批量获取文件标签
func (s *TagService) GetFileTagsMap(fileIDs []uint) (map[uint][]model.FileTag, error) {
	return s.fileTagRepo.GetByFileIDs(fileIDs)
}

// AddFileTag 添加文件标签
func (s *TagService) AddFileTag(fileID uint, tagID int64, source string) error {
	// 检查标签是否存在
	_, err := s.tagRepo.GetByID(tagID)
	if err != nil {
		return fmt.Errorf("标签不存在: %d", tagID)
	}

	fileTag := &model.FileTag{
		FileID: fileID,
		TagID:  tagID,
		Source: source,
	}
	return s.fileTagRepo.Create(fileTag)
}

// RemoveFileTag 移除文件标签
func (s *TagService) RemoveFileTag(fileID uint, tagID int64) error {
	return s.fileTagRepo.Delete(fileID, tagID)
}

// ReplaceFileTags 替换文件的标签
func (s *TagService) ReplaceFileTags(fileID uint, tagIDs []int64, source string) error {
	fileTags := make([]model.FileTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		fileTags = append(fileTags, model.FileTag{
			FileID: fileID,
			TagID:  tagID,
			Source: source,
		})
	}
	return s.fileTagRepo.ReplaceFileTags(fileID, fileTags)
}

// AutoGenerateTags 自动生成文件标签
func (s *TagService) AutoGenerateTags(filename, contentType, source string) []storage.RoutingTag {
	tagger := storage.GetAutoTagger()
	if tagger == nil {
		return nil
	}
	return tagger.GenerateTags(filename, contentType, source)
}
