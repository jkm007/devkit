package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

// LoginDeviceService 登录设备服务
type LoginDeviceService struct {
	repo *repository.LoginDeviceRepo
}

// NewLoginDeviceService 创建登录设备服务
func NewLoginDeviceService() *LoginDeviceService {
	return &LoginDeviceService{
		repo: repository.NewLoginDeviceRepo(database.GetMySQL()),
	}
}

// CreateOrUpdateDevice 创建或更新设备记录
func (s *LoginDeviceService) CreateOrUpdateDevice(device *model.LoginDevice) error {
	existing, err := s.repo.GetByUserIDAndDeviceID(device.UserID, device.DeviceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existing != nil {
		// 更新已有设备
		existing.DeviceName = device.DeviceName
		existing.Browser = device.Browser
		existing.OS = device.OS
		existing.IP = device.IP
		existing.Location = device.Location
		existing.TokenJTI = device.TokenJTI
		existing.LastActiveAt = device.LastActiveAt
		existing.IsCurrent = 1
		return s.repo.Update(existing)
	}

	// 先将该用户其他设备标记为非当前
	devices, err := s.repo.ListByUser(device.UserID)
	if err != nil {
		return err
	}
	for _, d := range devices {
		if d.IsCurrent == 1 {
			d.IsCurrent = 0
			if err := s.repo.Update(&d); err != nil {
				return err
			}
		}
	}

	device.IsCurrent = 1
	return s.repo.Create(device)
}

// List 获取用户的设备列表
func (s *LoginDeviceService) List(userID uint) ([]model.LoginDevice, error) {
	return s.repo.ListByUser(userID)
}

// KickDevice 踢出指定设备
// currentDeviceID: 前端传来的当前设备ID，用于防止踢自己
func (s *LoginDeviceService) KickDevice(userID, deviceID uint, currentDeviceID string) error {
	device, err := s.repo.GetByID(deviceID)
	if err != nil {
		return err
	}
	if device.UserID != userID {
		return errors.New("device does not belong to user")
	}
	// 用前端传来的设备ID判断是否是当前设备
	if currentDeviceID != "" && device.DeviceID == currentDeviceID {
		return errors.New("Cannot kick current device, use logout instead.")
	}

	// 将被踢出的设备ID加入Redis黑名单，TTL 24小时（JWT最大有效期）
	ctx := context.Background()
	blacklistKey := fmt.Sprintf("kicked_device:%d:%s", userID, device.DeviceID)
	database.GetRedis().Set(ctx, blacklistKey, 1, 24*time.Hour)

	return s.repo.Delete(deviceID)
}

// KickAllOther 踢出除当前设备外的所有设备
// currentDeviceID: 前端传来的当前设备ID
func (s *LoginDeviceService) KickAllOther(userID uint, currentDeviceID string) (int64, error) {
	devices, err := s.repo.ListByUser(userID)
	if err != nil {
		return 0, err
	}

	// 用前端传来的设备ID找到当前设备
	var currentID uint
	for _, d := range devices {
		if currentDeviceID != "" && d.DeviceID == currentDeviceID {
			currentID = d.ID
			break
		}
	}

	// 兜底：如果前端没传设备ID，用 IsCurrent 标记
	if currentID == 0 {
		for _, d := range devices {
			if d.IsCurrent == 1 {
				currentID = d.ID
				break
			}
		}
	}

	if currentID == 0 {
		return 0, errors.New("no current device found")
	}

	// 将所有其他设备加入Redis黑名单
	ctx := context.Background()
	for _, d := range devices {
		if d.ID != currentID {
			blacklistKey := fmt.Sprintf("kicked_device:%d:%s", userID, d.DeviceID)
			database.GetRedis().Set(ctx, blacklistKey, 1, 24*time.Hour)
		}
	}

	return s.repo.DeleteByUserExcept(userID, currentID)
}

// DeleteAllDevices 删除用户的所有登录设备记录（用于重置密码等场景）
func (s *LoginDeviceService) DeleteAllDevices(userID uint) error {
	// 将所有设备加入Redis黑名单
	devices, err := s.repo.ListByUser(userID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, d := range devices {
		blacklistKey := fmt.Sprintf("kicked_device:%d:%s", userID, d.DeviceID)
		database.GetRedis().Set(ctx, blacklistKey, 1, 24*time.Hour)
	}

	return s.repo.DeleteAllByUser(userID)
}
