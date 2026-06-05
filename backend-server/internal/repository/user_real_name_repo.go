package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// UserRealNameRepo 实名认证仓库
type UserRealNameRepo struct {
	db *gorm.DB
}

// NewUserRealNameRepo 创建实名认证仓库
func NewUserRealNameRepo(db *gorm.DB) *UserRealNameRepo {
	return &UserRealNameRepo{db: db}
}

// Create 创建实名认证记录
func (r *UserRealNameRepo) Create(rn *model.UserRealName) error {
	return r.db.Create(rn).Error
}

// Update 更新实名认证记录
func (r *UserRealNameRepo) Update(rn *model.UserRealName) error {
	return r.db.Save(rn).Error
}

// GetByUserID 根据用户ID获取实名认证记录
func (r *UserRealNameRepo) GetByUserID(userID uint) (*model.UserRealName, error) {
	var rn model.UserRealName
	if err := r.db.Where("user_id = ?", userID).First(&rn).Error; err != nil {
		return nil, err
	}
	return &rn, nil
}

// GetByID 根据ID获取实名认证记录
func (r *UserRealNameRepo) GetByID(id uint) (*model.UserRealName, error) {
	var rn model.UserRealName
	if err := r.db.Where("id = ?", id).First(&rn).Error; err != nil {
		return nil, err
	}
	return &rn, nil
}

// GetByIDCardHash 根据身份证号哈希获取实名认证记录（用于查重）
func (r *UserRealNameRepo) GetByIDCardHash(hash string) (*model.UserRealName, error) {
	var rn model.UserRealName
	if err := r.db.Where("id_card_hash = ? AND status != 2", hash).First(&rn).Error; err != nil {
		return nil, err
	}
	return &rn, nil
}

// List 查询实名认证列表（管理员，分页）
func (r *UserRealNameRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.UserRealName, int64, error) {
	var list []model.UserRealName
	var total int64

	query := r.db.Model(&model.UserRealName{})

	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if userID, ok := filters["userId"]; ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if realName, ok := filters["realName"]; ok && realName != "" {
		query = query.Where("real_name LIKE ?", "%"+escapeLike(realName.(string))+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
