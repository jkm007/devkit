package model

import "time"

// QuickMenu 移动端快捷菜单
type QuickMenu struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"type:varchar(100);not null" json:"title"`
	Icon      string    `gorm:"type:varchar(50);not null" json:"icon"`
	Link      string    `gorm:"type:varchar(512)" json:"link"`
	LinkType  string    `gorm:"type:varchar(20);default:page" json:"linkType"` // page / url / function / none
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
	Status    string    `gorm:"type:varchar(20);default:enabled" json:"status"` // enabled / disabled
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (QuickMenu) TableName() string {
	return "mobile_quick_menus"
}

// MyPageMenu 移动端"我的"页面菜单
type MyPageMenu struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"type:varchar(100);not null" json:"title"`
	Icon      string    `gorm:"type:varchar(50);not null" json:"icon"`
	Link      string    `gorm:"type:varchar(512);not null" json:"link"`
	ShowBadge bool      `gorm:"default:false" json:"showBadge"`
	BadgeText string    `gorm:"type:varchar(50)" json:"badgeText"`
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
	Status    string    `gorm:"type:varchar(20);default:enabled" json:"status"` // enabled / disabled
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (MyPageMenu) TableName() string {
	return "mobile_my_page_menus"
}

// MobileSettings 移动端全局设置
type MobileSettings struct {
	ID                 uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	NoticeEnabled      bool   `gorm:"default:false" json:"noticeEnabled"`
	NoticeContent      string `gorm:"type:text" json:"noticeContent"`
	AppDownloadUrl     string `gorm:"type:varchar(512)" json:"appDownloadUrl"`
	CustomerServiceUrl string `gorm:"type:varchar(512)" json:"customerServiceUrl"`
	AboutUs            string `gorm:"type:text" json:"aboutUs"`
	AgreementUrl       string `gorm:"type:varchar(512)" json:"agreementUrl"`
	PrivacyUrl         string `gorm:"type:varchar(512)" json:"privacyUrl"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func (MobileSettings) TableName() string {
	return "mobile_settings"
}

// 状态常量
const (
	MobileConfigEnabled  = "enabled"
	MobileConfigDisabled = "disabled"
)

// 链接类型常量
const (
	QuickMenuLinkPage     = "page"
	QuickMenuLinkURL      = "url"
	QuickMenuLinkFunction = "function"
	QuickMenuLinkNone     = "none"
)
