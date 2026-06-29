package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// ClassRepo 班级数据访问
type ClassRepo struct {
	db *gorm.DB
}

// NewClassRepo 创建班级仓库
func NewClassRepo(db *gorm.DB) *ClassRepo {
	return &ClassRepo{db: db}
}

// DB 返回底层数据库连接
func (r *ClassRepo) DB() *gorm.DB {
	return r.db
}

// ==================== 班级 CRUD ====================

// List 分页获取班级列表
func (r *ClassRepo) List(page, pageSize int, keyword string) ([]model.Class, int64, error) {
	var items []model.Class
	var total int64

	query := r.db.Model(&model.Class{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+escapeLike(keyword)+"%", "%"+escapeLike(keyword)+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetByID 根据ID获取班级
func (r *ClassRepo) GetByID(id uint) (*model.Class, error) {
	var item model.Class
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// GetByCode 根据邀请码获取班级
func (r *ClassRepo) GetByCode(code string) (*model.Class, error) {
	var item model.Class
	if err := r.db.Where("code = ?", code).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// Create 创建班级
func (r *ClassRepo) Create(item *model.Class) error {
	return r.db.Create(item).Error
}

// Update 更新班级
func (r *ClassRepo) Update(item *model.Class) error {
	return r.db.Save(item).Error
}

// Delete 删除班级（软删除）
func (r *ClassRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Class{}).Error
}

// CodeExists 判断邀请码是否已存在
func (r *ClassRepo) CodeExists(code string) bool {
	var count int64
	r.db.Model(&model.Class{}).Where("code = ?", code).Count(&count)
	return count > 0
}

// ==================== 班级成员 ====================

// AddMember 添加班级成员
func (r *ClassRepo) AddMember(member *model.ClassMember) error {
	return r.db.Create(member).Error
}

// GetMemberByUserID 获取指定用户在班级中的成员记录
func (r *ClassRepo) GetMemberByUserID(classID, userID uint) (*model.ClassMember, error) {
	var member model.ClassMember
	if err := r.db.Where("class_id = ? AND user_id = ?", classID, userID).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// ListMembersByClassID 获取班级成员列表
func (r *ClassRepo) ListMembersByClassID(classID uint, page, pageSize int) ([]model.ClassMember, int64, error) {
	var items []model.ClassMember
	var total int64

	query := r.db.Model(&model.ClassMember{}).Where("class_id = ?", classID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UpdateMemberRole 更新成员角色
func (r *ClassRepo) UpdateMemberRole(memberID uint, role model.ClassMemberRole) error {
	return r.db.Model(&model.ClassMember{}).Where("id = ?", memberID).Update("role", role).Error
}

// RemoveMember 移除成员
func (r *ClassRepo) RemoveMember(memberID uint) error {
	return r.db.Where("id = ?", memberID).Delete(&model.ClassMember{}).Error
}

// ListClassIDsByUserID 获取用户加入的所有班级ID
func (r *ClassRepo) ListClassIDsByUserID(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&model.ClassMember{}).
		Where("user_id = ? AND status = ?", userID, 1).
		Pluck("class_id", &ids).Error
	return ids, err
}

// CountMembers 统计班级成员数量
func (r *ClassRepo) CountMembers(classID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ClassMember{}).Where("class_id = ? AND status = ?", classID, 1).Count(&count).Error
	return count, err
}

// CountMembersByRole 统计班级中指定角色的成员数
func (r *ClassRepo) CountMembersByRole(classID uint, role model.ClassMemberRole) (int64, error) {
	var count int64
	err := r.db.Model(&model.ClassMember{}).Where("class_id = ? AND role = ? AND status = ?", classID, role, 1).Count(&count).Error
	return count, err
}

// DeleteMembersByClassID 删除班级所有成员（软删除班级时使用）
func (r *ClassRepo) DeleteMembersByClassID(classID uint) error {
	return r.db.Where("class_id = ?", classID).Delete(&model.ClassMember{}).Error
}

// ==================== 班级邀请码 ====================

// CreateInvitation 创建邀请码
func (r *ClassRepo) CreateInvitation(invitation *model.ClassInvitation) error {
	return r.db.Create(invitation).Error
}

// GetInvitationByCode 根据邀请码获取邀请记录
func (r *ClassRepo) GetInvitationByCode(code string) (*model.ClassInvitation, error) {
	var item model.ClassInvitation
	if err := r.db.Where("code = ?", code).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// ListInvitationsByClassID 获取班级邀请码列表
func (r *ClassRepo) ListInvitationsByClassID(classID uint) ([]model.ClassInvitation, error) {
	var items []model.ClassInvitation
	if err := r.db.Where("class_id = ?", classID).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// DisableInvitation 禁用邀请码
func (r *ClassRepo) DisableInvitation(id uint) error {
	return r.db.Model(&model.ClassInvitation{}).Where("id = ?", id).Update("status", 0).Error
}

// IncrementInvitationUsedCount 增加邀请码使用次数
func (r *ClassRepo) IncrementInvitationUsedCount(id uint) error {
	return r.db.Model(&model.ClassInvitation{}).Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}

// DeleteInvitationsByClassID 删除班级所有邀请码
func (r *ClassRepo) DeleteInvitationsByClassID(classID uint) error {
	return r.db.Where("class_id = ?", classID).Delete(&model.ClassInvitation{}).Error
}

// ==================== 事务辅助 ====================

// DeleteClassWithRelations 删除班级及其成员、邀请码（事务）
func (r *ClassRepo) DeleteClassWithRelations(classID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("class_id = ?", classID).Delete(&model.ClassMember{}).Error; err != nil {
			return err
		}
		if err := tx.Where("class_id = ?", classID).Delete(&model.ClassInvitation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", classID).Delete(&model.Class{}).Error; err != nil {
			return err
		}
		return nil
	})
}
