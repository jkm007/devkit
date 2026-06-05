package model

import "time"

// OAuthUser 第三方登录绑定模型
type OAuthUser struct {
	ID              uint      `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID          uint      `gorm:"not null;comment:用户ID" json:"userId"`
	Provider        string    `gorm:"type:varchar(20);not null;comment:提供商" json:"provider"`
	ProviderType    string    `gorm:"type:varchar(20);default:;comment:提供商类型" json:"providerType"`
	ProviderUserID  string    `gorm:"type:varchar(100);not null;comment:第三方用户ID" json:"-"`
	ProviderUsername string   `gorm:"type:varchar(100);default:;comment:第三方用户名" json:"providerUsername"`
	ProviderAvatar  string    `gorm:"type:varchar(255);default:;comment:第三方头像" json:"providerAvatar"`
	AccessToken     string    `gorm:"type:varchar(500);default:;comment:Access Token" json:"-"`
	RefreshToken    string    `gorm:"type:varchar(500);default:;comment:Refresh Token" json:"-"`
	ExpiresAt       *time.Time `gorm:"comment:Token过期时间" json:"-"`
	CreatedAt       time.Time `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"comment:更新时间" json:"-"`
}

// TableName 表名
func (OAuthUser) TableName() string {
	return "sys_oauth_users"
}
