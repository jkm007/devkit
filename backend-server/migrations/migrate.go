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
		&model.GroupRole{},
		&model.UserRole{},
		&model.SecurityLog{},
		&model.LoginDevice{},
		&model.OAuthUser{},
		&model.UserPrivacy{},
		&model.PasswordHistory{},
		&model.UserRealName{},
		&model.RoleApplication{},
		&model.SystemSetting{},
		&model.UploadTask{},
		&model.UploadedPart{},
		&model.FileAsset{},
		&model.FileFolder{},
		&model.FileEntry{},
		&model.MediaAsset{},
		&model.FileShare{},
		&model.Tag{},
		&model.TagRouting{},
		&model.FileTag{},
		&model.StorageBucket{},
		&model.StorageConfig{},
		&model.FileTypeRule{},
		&model.RateLimitRule{},
		&model.ScheduledTask{},
		&model.Notification{},
		&model.NotificationRead{},
		&model.ExamCategory{},
		&model.Exam{},
		&model.Subject{},
		&model.Category{},
		&model.KnowledgePoint{},
		&model.QuestionSource{},
		&model.Question{},
	)
}
