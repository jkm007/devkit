package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// LoginDeviceRepo 登录设备仓库
type LoginDeviceRepo struct {
	db *gorm.DB
}

// NewLoginDeviceRepo 创建登录设备仓库
func NewLoginDeviceRepo(db *gorm.DB) *LoginDeviceRepo {
	return &LoginDeviceRepo{db: db}
}

// Create 创建设备记录
func (r *LoginDeviceRepo) Create(device *model.LoginDevice) error {
	return r.db.Create(device).Error
}

// Update 更新设备记录
func (r *LoginDeviceRepo) Update(device *model.LoginDevice) error {
	return r.db.Save(device).Error
}

// ListByUser 获取用户的设备列表
func (r *LoginDeviceRepo) ListByUser(userID uint) ([]model.LoginDevice, error) {
	return r.ListByUserAndType(userID, "")
}

// ListByUserAndType 获取用户的设备列表，支持按设备类型过滤
func (r *LoginDeviceRepo) ListByUserAndType(userID uint, deviceType string) ([]model.LoginDevice, error) {
	var devices []model.LoginDevice
	query := r.db.Where("user_id = ?", userID)
	if deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}
	err := query.Order("last_active_at DESC").Find(&devices).Error
	return devices, err
}

// GetByID 根据 ID 获取设备
func (r *LoginDeviceRepo) GetByID(id uint) (*model.LoginDevice, error) {
	var device model.LoginDevice
	if err := r.db.Where("id = ?", id).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

// GetByUserIDAndDeviceID 根据用户ID和设备ID获取设备
func (r *LoginDeviceRepo) GetByUserIDAndDeviceID(userID uint, deviceID string) (*model.LoginDevice, error) {
	var device model.LoginDevice
	if err := r.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

// Delete 删除设备
func (r *LoginDeviceRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.LoginDevice{}).Error
}

// DeleteByUserExcept 删除用户除指定设备外的所有设备
func (r *LoginDeviceRepo) DeleteByUserExcept(userID uint, exceptID uint) (int64, error) {
	result := r.db.Where("user_id = ? AND id != ?", userID, exceptID).Delete(&model.LoginDevice{})
	return result.RowsAffected, result.Error
}

// DeleteAllByUser 删除用户的所有设备记录
func (r *LoginDeviceRepo) DeleteAllByUser(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.LoginDevice{}).Error
}
