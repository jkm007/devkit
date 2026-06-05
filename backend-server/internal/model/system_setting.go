package model

// SystemSetting 系统配置
type SystemSetting struct {
	BaseModel
	GroupKey    string `gorm:"type:varchar(50);not null;uniqueIndex:uk_group_key" json:"group_key"`
	Key         string `gorm:"type:varchar(100);not null;uniqueIndex:uk_group_key" json:"key"`
	Value       string `gorm:"type:text" json:"value"`
	Label       string `gorm:"type:varchar(100);not null" json:"label"`
	Type        string `gorm:"type:varchar(20);not null;default:string" json:"type"`
	Options     string `gorm:"type:text" json:"options"`
	Tip         string `gorm:"type:varchar(500);default:" json:"tip"`
	Sort        int    `gorm:"not null;default:0" json:"sort"`
	IsPublic    int8   `gorm:"type:tinyint;not null;default:0" json:"is_public"`
	IsSensitive int8   `gorm:"type:tinyint;not null;default:0" json:"is_sensitive"`
	UpdatedBy   *uint  `gorm:"type:bigint unsigned" json:"updated_by"`
}

// TableName 表名
func (SystemSetting) TableName() string {
	return "sys_system_settings"
}
