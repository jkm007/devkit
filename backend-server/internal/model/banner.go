package model

import "time"

// Banner 轮播图
type Banner struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Image     string    `gorm:"type:varchar(512);not null" json:"image"`
	Link      string    `gorm:"type:varchar(512)" json:"link"`
	LinkType  string    `gorm:"type:varchar(20);default:none" json:"linkType"` // internal / external / none
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
	Status    string    `gorm:"type:varchar(20);default:enabled" json:"status"` // enabled / disabled
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Banner) TableName() string {
	return "banners"
}

// Banner 状态常量
const (
	BannerEnabled  = "enabled"
	BannerDisabled = "disabled"
)

// Banner 链接类型常量
const (
	BannerLinkInternal = "internal"
	BannerLinkExternal = "external"
	BannerLinkNone     = "none"
)
