package service

import (
	"time"

	"backend-server/internal/model"
	"backend-server/pkg/database"
)

// DashboardService 仪表盘服务
type DashboardService struct{}

func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

// DashboardStats 仪表盘统计
type DashboardStats struct {
	Overview           OverviewStats     `json:"overview"`
	EventsTrend        []DayEvents       `json:"eventsTrend"`
	EventsByType       []TypeCount       `json:"eventsByType"`
	DeviceByType       []TypeCount       `json:"deviceByType"`
	DeviceByPlatform   []TypeCount       `json:"deviceByPlatform"`
	RecentLogins       []RecentLogin     `json:"recentLogins"`
}

// OverviewStats 概览统计
type OverviewStats struct {
	UserCount    int64   `json:"userCount"`
	ActiveUsers  int64   `json:"activeUsers"`
	FileCount    int64   `json:"fileCount"`
	TotalStorage int64   `json:"totalStorage"`
	TodayLogins  int64   `json:"todayLogins"`
	TodayEvents  int64   `json:"todayEvents"`
	OnlineDevices int64  `json:"onlineDevices"`
}

// DayEvents 单日事件统计
type DayEvents struct {
	Date    string `json:"date"`
	Success int64  `json:"success"`
	Fail    int64  `json:"fail"`
}

// TypeCount 按类型统计
type TypeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// RecentLogin 最近登录记录
type RecentLogin struct {
	Username  string    `json:"username"`
	IP        string    `json:"ip"`
	Device    string    `json:"device"`
	Location  string    `json:"location"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetStats 获取仪表盘统计数据
func (s *DashboardService) GetStats() (*DashboardStats, error) {
	db := database.GetMySQL()
	stats := &DashboardStats{}

	// 概览统计
	db.Model(&model.User{}).Count(&stats.Overview.UserCount)
	db.Model(&model.User{}).Where("status = 1").Count(&stats.Overview.ActiveUsers)
	db.Model(&model.FileEntry{}).Where("deleted_at IS NULL").Count(&stats.Overview.FileCount)
	db.Model(&model.FileEntry{}).Where("deleted_at IS NULL").Select("COALESCE(SUM(size), 0)").Scan(&stats.Overview.TotalStorage)

	today := time.Now().Format("2006-01-02")
	db.Model(&model.SecurityLog{}).Where("event_type = 'login' AND status = 1 AND DATE(created_at) = ?", today).Count(&stats.Overview.TodayLogins)
	db.Model(&model.SecurityLog{}).Where("DATE(created_at) = ?", today).Count(&stats.Overview.TodayEvents)
	db.Model(&model.LoginDevice{}).Count(&stats.Overview.OnlineDevices)

	// 近7天事件趋势
	var eventsTrend []DayEvents
	for i := 6; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var success, fail int64
		db.Model(&model.SecurityLog{}).Where("DATE(created_at) = ? AND status = 1", d).Count(&success)
		db.Model(&model.SecurityLog{}).Where("DATE(created_at) = ? AND status = 0", d).Count(&fail)
		eventsTrend = append(eventsTrend, DayEvents{Date: d, Success: success, Fail: fail})
	}
	if eventsTrend == nil {
		eventsTrend = []DayEvents{}
	}
	stats.EventsTrend = eventsTrend

	// 事件类型分布（近30天）
	var eventsByType []TypeCount
	db.Model(&model.SecurityLog{}).
		Select("event_type as type, COUNT(*) as count").
		Where("created_at >= ?", time.Now().AddDate(0, 0, -30)).
		Group("event_type").
		Order("count DESC").
		Find(&eventsByType)
	if eventsByType == nil {
		eventsByType = []TypeCount{}
	}
	stats.EventsByType = eventsByType

	// 设备类型分布
	var deviceByType []TypeCount
	db.Model(&model.LoginDevice{}).
		Select("device_type as type, COUNT(*) as count").
		Group("device_type").
		Order("count DESC").
		Find(&deviceByType)
	if deviceByType == nil {
		deviceByType = []TypeCount{}
	}
	stats.DeviceByType = deviceByType

	// 平台分布
	var deviceByPlatform []TypeCount
	db.Model(&model.LoginDevice{}).
		Select("platform as type, COUNT(*) as count").
		Where("platform != ''").
		Group("platform").
		Order("count DESC").
		Find(&deviceByPlatform)
	if deviceByPlatform == nil {
		deviceByPlatform = []TypeCount{}
	}
	stats.DeviceByPlatform = deviceByPlatform

	// 最近登录记录
	type loginRow struct {
		Username  string
		IP        string
		UserAgent string
		Status    int
		CreatedAt time.Time
	}
	var logins []loginRow
	db.Model(&model.SecurityLog{}).
		Select("username, ip, user_agent, status, created_at").
		Where("event_type = 'login'").
		Order("created_at DESC").
		Limit(10).
		Find(&logins)
	var recentLogins []RecentLogin
	for _, l := range logins {
		recentLogins = append(recentLogins, RecentLogin{
			Username:  l.Username,
			IP:        l.IP,
			Device:    l.UserAgent,
			Status:    l.Status,
			CreatedAt: l.CreatedAt,
		})
	}
	if recentLogins == nil {
		recentLogins = []RecentLogin{}
	}
	stats.RecentLogins = recentLogins

	return stats, nil
}
