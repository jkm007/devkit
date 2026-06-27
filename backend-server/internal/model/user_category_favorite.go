package model

import "time"

// UserCategoryFavorite 用户分类收藏
// 支持收藏四级分类中的任意一级：
//   - exam_category: 考试大类(L1)
//   - exam: 具体考试(L2)
//   - subject: 科目(L3)
//   - category: 章节分类(L4)
type UserCategoryFavorite struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"index:idx_user_category_favorite_target,unique;not null;comment:用户ID" json:"userId"`
	TargetID   uint      `gorm:"index:idx_user_category_favorite_target,unique;not null;comment:目标分类ID" json:"targetId"`
	TargetType string    `gorm:"index:idx_user_category_favorite_target,unique;type:varchar(20);not null;comment:目标类型" json:"targetType"`
	TargetName string    `gorm:"type:varchar(150);not null;comment:目标名称" json:"targetName"`
	Path       string    `gorm:"type:varchar(1000);default:;comment:分类路径" json:"path"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:收藏时间" json:"createdAt"`
}

// TableName 表名
func (UserCategoryFavorite) TableName() string {
	return "user_category_favorites"
}
