package model

// UserRole 用户角色关联模型
type UserRole struct {
	ID     uint `gorm:"primarykey;comment:主键ID" json:"-"`
	UserID uint `gorm:"not null;index;comment:用户ID" json:"userId"`
	RoleID uint `gorm:"not null;index;comment:角色ID" json:"roleId"`
}

// TableName 表名
func (UserRole) TableName() string {
	return "sys_user_roles"
}
