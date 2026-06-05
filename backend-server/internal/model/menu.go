package model

import (
	"time"

	"gorm.io/gorm"
)

// Menu 菜单模型
type Menu struct {
	ID        uint           `gorm:"primaryKey;autoIncrement;comment:菜单ID" json:"id"`
	PID       uint           `gorm:"column:pid;comment:父菜单ID" json:"pid"`
	Name      string         `gorm:"type:varchar(100);not null;comment:菜单名称" json:"name"`
	Path      string         `gorm:"type:varchar(200);comment:路由地址" json:"path"`
	Component string         `gorm:"type:varchar(200);comment:组件路径" json:"component"`
	Type      string         `gorm:"type:varchar(20);not null;comment:类型 catalog/menu/embedded/link/button" json:"type"`
	Status    int            `gorm:"type:tinyint;default:1;comment:状态 1:启用 0:禁用" json:"status"`
	AuthCode  string         `gorm:"type:varchar(100);comment:权限码" json:"authCode"`
	Icon      string         `gorm:"type:varchar(100);comment:图标" json:"icon"`
	Meta      string         `gorm:"type:text;comment:元数据(JSON)" json:"meta"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"createTime"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 表名
func (Menu) TableName() string {
	return "sys_menus"
}

// MenuMeta 菜单元数据
type MenuMeta struct {
	Icon               string `json:"icon,omitempty"`
	Title              string `json:"title"`
	Order              int    `json:"order,omitempty"`
	HideInMenu         bool   `json:"hideInMenu,omitempty"`
	KeepAlive          bool   `json:"keepAlive,omitempty"`
	AffixTab           bool   `json:"affixTab,omitempty"`
	HideInTab          bool   `json:"hideInTab,omitempty"`
	HideChildrenInMenu bool   `json:"hideChildrenInMenu,omitempty"`
	Badge              string `json:"badge,omitempty"`
	BadgeType          string `json:"badgeType,omitempty"`
	IframeSrc          string `json:"iframeSrc,omitempty"`
	Link               string `json:"link,omitempty"`
}

// MenuTree 菜单树结构
type MenuTree struct {
	Menu
	Children []*MenuTree `json:"children,omitempty"`
}
