package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// SystemSettingRepo 系统配置仓库
type SystemSettingRepo struct {
	db *gorm.DB
}

// NewSystemSettingRepo 创建系统配置仓库
func NewSystemSettingRepo(db *gorm.DB) *SystemSettingRepo {
	return &SystemSettingRepo{db: db}
}

// GetAll 获取所有配置
func (r *SystemSettingRepo) GetAll() ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := r.db.Order("`group_key` ASC, `sort` ASC").Find(&settings).Error
	return settings, err
}

// GetByGroup 获取指定分组配置
func (r *SystemSettingRepo) GetByGroup(groupKey string) ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := r.db.Where("group_key = ?", groupKey).Order("`sort` ASC").Find(&settings).Error
	return settings, err
}

// GetPublic 获取公开配置
func (r *SystemSettingRepo) GetPublic() ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := r.db.Where("is_public = 1").Order("`group_key` ASC, `sort` ASC").Find(&settings).Error
	return settings, err
}

// GetByKey 根据分组和键获取单个配置
func (r *SystemSettingRepo) GetByKey(groupKey, key string) (*model.SystemSetting, error) {
	var setting model.SystemSetting
	err := r.db.Where("group_key = ? AND `key` = ?", groupKey, key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// BatchUpdate 批量更新配置（事务）
func (r *SystemSettingRepo) BatchUpdate(updates []model.SystemSetting, userID uint) (int, error) {
	updated := 0
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for i := range updates {
			u := updates[i]
			result := tx.Model(&model.SystemSetting{}).
				Where("group_key = ? AND `key` = ?", u.GroupKey, u.Key).
				Updates(map[string]interface{}{
					"value":      u.Value,
					"updated_by": userID,
				})
			if result.Error != nil {
				return result.Error
			}
			updated += int(result.RowsAffected)
		}
		return nil
	})
	return updated, err
}

// GetKeysByGroup 获取指定分组的所有 key
func (r *SystemSettingRepo) GetKeysByGroup(groupKey string) (map[string]bool, error) {
	var settings []model.SystemSetting
	err := r.db.Where("group_key = ?", groupKey).Select("`key`").Find(&settings).Error
	if err != nil {
		return nil, err
	}
	keys := make(map[string]bool, len(settings))
	for _, s := range settings {
		keys[s.Key] = true
	}
	return keys, nil
}
