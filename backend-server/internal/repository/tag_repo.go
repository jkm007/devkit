package repository

import (
	"backend-server/internal/model"
	"gorm.io/gorm"
)

// TagRepo 标签仓库
type TagRepo struct {
	db *gorm.DB
}

// NewTagRepo 创建标签仓库
func NewTagRepo(db *gorm.DB) *TagRepo {
	return &TagRepo{db: db}
}

// GetAll 获取所有标签
func (r *TagRepo) GetAll() ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Where("status = ?", 1).Order("sort_order ASC, id ASC").Find(&tags).Error
	return tags, err
}

// GetByKey 获取指定键的所有标签值
func (r *TagRepo) GetByKey(tagKey string) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Where("tag_key = ? AND status = ?", tagKey, 1).Order("sort_order ASC").Find(&tags).Error
	return tags, err
}

// GetByKeyValue 根据键值获取标签
func (r *TagRepo) GetByKeyValue(tagKey, tagValue string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("tag_key = ? AND tag_value = ?", tagKey, tagValue).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetByID 根据ID获取标签
func (r *TagRepo) GetByID(id int64) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("id = ?", id).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetByIDs 根据ID列表获取标签
func (r *TagRepo) GetByIDs(ids []int64) (map[int64]*model.Tag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var tags []model.Tag
	if err := r.db.Where("id IN ?", ids).Find(&tags).Error; err != nil {
		return nil, err
	}
	result := make(map[int64]*model.Tag, len(tags))
	for i := range tags {
		result[tags[i].ID] = &tags[i]
	}
	return result, nil
}

// Create 创建标签
func (r *TagRepo) Create(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

// Update 更新标签
func (r *TagRepo) Update(tag *model.Tag) error {
	return r.db.Save(tag).Error
}

// Delete 删除标签（软删除）
func (r *TagRepo) Delete(id int64) error {
	return r.db.Delete(&model.Tag{}, id).Error
}

// GetGroupedTags 获取按 key 分组的标签
func (r *TagRepo) GetGroupedTags() (map[string][]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Where("status = ?", 1).Order("sort_order ASC").Find(&tags).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string][]model.Tag)
	for _, tag := range tags {
		result[tag.TagKey] = append(result[tag.TagKey], tag)
	}
	return result, nil
}

// GetUsageStats 获取标签使用统计
func (r *TagRepo) GetUsageStats() ([]TagUsageStat, error) {
	var stats []TagUsageStat
	err := r.db.Raw(`
		SELECT t.id, t.tag_key, t.tag_value, t.tag_name, t.icon, t.color,
			COUNT(ft.id) as file_count
		FROM sys_tag t
		LEFT JOIN sys_file_tag ft ON t.id = ft.tag_id
		WHERE t.status = 1
		GROUP BY t.id, t.tag_key, t.tag_value, t.tag_name, t.icon, t.color
		ORDER BY file_count DESC
	`).Scan(&stats).Error
	return stats, err
}

// TagUsageStat 标签使用统计
type TagUsageStat struct {
	ID        int64  `json:"id"`
	TagKey    string `json:"tagKey"`
	TagValue  string `json:"tagValue"`
	TagName   string `json:"tagName"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	FileCount int64  `json:"fileCount"`
}
