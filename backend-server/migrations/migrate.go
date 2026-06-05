package migrations

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// Run 执行数据库迁移
func Run(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Menu{},
		&model.Group{},
		&model.UserRole{},
		&model.SecurityLog{},
		&model.LoginDevice{},
		&model.OAuthUser{},
		&model.UserPrivacy{},
		&model.PasswordHistory{},
		&model.UserRealName{},
		&model.RoleApplication{},
		&model.SystemSetting{},
	)
}
