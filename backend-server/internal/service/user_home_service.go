package service

import (
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

// UserHomeService 用户首页服务
type UserHomeService struct {
	userRepo *repository.UserRepo
}

func NewUserHomeService() *UserHomeService {
	db := database.GetMySQL()
	return &UserHomeService{
		userRepo: repository.NewUserRepo(db),
	}
}

// HomeData 用户首页数据
type HomeData struct {
	Storage    StorageInfo    `json:"storage"`
	Roles      []RoleInfo     `json:"roles"`
	Devices    []DeviceInfo   `json:"devices"`
}

// StorageInfo 存储信息
type StorageInfo struct {
	Used  int64   `json:"used"`
	Quota int64   `json:"quota"` // 0 = 不限制
	UsedPercent float64 `json:"usedPercent"`
}

// RoleInfo 角色信息
type RoleInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID           uint   `json:"id"`
	DeviceType   string `json:"deviceType"`
	DeviceName   string `json:"deviceName"`
	Browser      string `json:"browser"`
	OS           string `json:"os"`
	IP           string `json:"ip"`
	Platform     string `json:"platform"`
	IsCurrent    bool   `json:"isCurrent"`
	LastActiveAt string `json:"lastActiveAt"`
}

// GetHomeData 获取用户首页数据
func (s *UserHomeService) GetHomeData(userID uint) (*HomeData, error) {
	db := database.GetMySQL()
	data := &HomeData{}

	// 1. 存储信息
	used, _ := s.userRepo.GetStorageUsedByUserID(userID)

	// 获取用户有效配额
	userService := NewUserService()
	quota, _ := userService.GetEffectiveQuota(userID)

	var usedPercent float64
	if quota > 0 {
		usedPercent = float64(used) / float64(quota) * 100
		if usedPercent > 100 {
			usedPercent = 100
		}
	}

	data.Storage = StorageInfo{
		Used:        used,
		Quota:       quota,
		UsedPercent: usedPercent,
	}

	// 2. 用户角色
	var roles []model.Role
	db.Raw(`
		SELECT r.id, r.name, r.status
		FROM sys_roles r
		INNER JOIN sys_user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = ?
	`, userID).Scan(&roles)

	for _, r := range roles {
		data.Roles = append(data.Roles, RoleInfo{
			ID:   r.ID,
			Name: r.Name,
		})
	}

	// 3. 登录设备（最近10个）
	var devices []model.LoginDevice
	db.Where("user_id = ?", userID).
		Order("last_active_at DESC").
		Limit(10).
		Find(&devices)

	for _, d := range devices {
		lastActive := ""
		if d.LastActiveAt != nil {
			lastActive = d.LastActiveAt.Format("2006-01-02 15:04:05")
		}
		data.Devices = append(data.Devices, DeviceInfo{
			ID:           d.ID,
			DeviceType:   d.DeviceType,
			DeviceName:   d.DeviceName,
			Browser:      d.Browser,
			OS:           d.OS,
			IP:           d.IP,
			Platform:     d.Platform,
			IsCurrent:    d.IsCurrent == 1,
			LastActiveAt: lastActive,
		})
	}

	return data, nil
}
