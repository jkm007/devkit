package model

import "time"

// LoginDevice 登录设备模型
type LoginDevice struct {
	ID            uint       `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID        uint       `gorm:"not null;comment:用户ID" json:"userId"`
	DeviceID      string     `gorm:"type:varchar(100);not null;comment:设备唯一标识" json:"deviceId"`
	DeviceType    string     `gorm:"type:varchar(20);not null;comment:设备类型 web/h5/app/miniapp" json:"deviceType"`
	DeviceName    string     `gorm:"type:varchar(100);default:;comment:设备名称" json:"deviceName"`
	Browser       string     `gorm:"type:varchar(50);default:;comment:浏览器" json:"browser"`
	OS            string     `gorm:"type:varchar(50);default:;comment:操作系统" json:"os"`
	IP            string     `gorm:"type:varchar(50);default:;comment:IP地址" json:"ip"`
	Location      string     `gorm:"type:varchar(100);default:;comment:登录地点" json:"location"`
	AppVersion    string     `gorm:"type:varchar(50);default:;comment:App版本" json:"appVersion"`
	SystemVersion string     `gorm:"type:varchar(100);default:;comment:系统版本" json:"systemVersion"`
	DeviceModel   string     `gorm:"type:varchar(100);default:;comment:设备型号" json:"deviceModel"`
	Platform      string     `gorm:"type:varchar(50);default:;comment:平台 ios/android/h5/web/miniapp" json:"platform"`
	Channel       string     `gorm:"type:varchar(50);default:;comment:渠道" json:"channel"`
	TokenJTI      string     `gorm:"type:varchar(100);not null;comment:JWT ID" json:"-"`
	LastActiveAt  *time.Time `gorm:"comment:最后活跃时间" json:"lastActiveAt"`
	IsCurrent     int        `gorm:"type:tinyint;default:0;comment:是否当前设备" json:"isCurrent"`
	CreatedAt     time.Time  `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"comment:更新时间" json:"-"`
}

// TableName 表名
func (LoginDevice) TableName() string {
	return "sys_login_devices"
}
