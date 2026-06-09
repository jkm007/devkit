package repository

import (
	"time"

	"backend-server/internal/model"

	"gorm.io/gorm"
)

type FileShareRepo struct {
	db *gorm.DB
}

func NewFileShareRepo(db *gorm.DB) *FileShareRepo {
	return &FileShareRepo{db: db}
}

func (r *FileShareRepo) Create(share *model.FileShare) error {
	return r.db.Create(share).Error
}

func (r *FileShareRepo) GetByShareCode(code string) (*model.FileShare, error) {
	var share model.FileShare
	err := r.db.Where("share_code = ?", code).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *FileShareRepo) GetByID(id uint) (*model.FileShare, error) {
	var share model.FileShare
	err := r.db.First(&share, id).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *FileShareRepo) GetByUserID(userID uint) ([]model.FileShare, error) {
	var shares []model.FileShare
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&shares).Error
	return shares, err
}

func (r *FileShareRepo) IncrementAccessCount(code string) error {
	now := time.Now()
	return r.db.Model(&model.FileShare{}).Where("share_code = ?", code).
		Updates(map[string]interface{}{
			"access_count": gorm.Expr("access_count + 1"),
			"accessed_at":  now,
		}).Error
}

func (r *FileShareRepo) Delete(id uint) error {
	return r.db.Delete(&model.FileShare{}, id).Error
}

// UpdateStatus 更新分享状态
func (r *FileShareRepo) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.FileShare{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateExpireAt 更新过期时间
func (r *FileShareRepo) UpdateExpireAt(id uint, expireAt *time.Time) error {
	return r.db.Model(&model.FileShare{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"expire_at": expireAt,
			"status":    1, // 重置为有效状态
		}).Error
}

// GetUserSharesWithFile 获取用户的分享列表（包含文件信息）
// viewAll=true 时查看所有分享，否则只查看自己的
func (r *FileShareRepo) GetUserSharesWithFile(userID uint, page, pageSize int, viewAll bool) ([]model.FileShare, int64, error) {
	var shares []model.FileShare
	var total int64

	query := r.db
	if !viewAll {
		query = query.Where("user_id = ?", userID)
	}

	// 统计总数
	if err := query.Model(&model.FileShare{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&shares).Error; err != nil {
		return nil, 0, err
	}

	return shares, total, nil
}

// GetActiveShares 获取有效的分享（未过期、未禁用）
func (r *FileShareRepo) GetActiveShares(userID uint) ([]model.FileShare, error) {
	var shares []model.FileShare
	now := time.Now()
	err := r.db.Where("user_id = ? AND status = 1 AND (expire_at IS NULL OR expire_at > ?)", userID, now).
		Order("created_at desc").Find(&shares).Error
	return shares, err
}

// CheckExpiredShares 检查并更新过期的分享
func (r *FileShareRepo) CheckExpiredShares() error {
	now := time.Now()
	return r.db.Model(&model.FileShare{}).
		Where("status = 1 AND expire_at IS NOT NULL AND expire_at <= ?", now).
		Update("status", 2).Error
}

// GetActiveByFileID 获取文件的有效分享记录
func (r *FileShareRepo) GetActiveByFileID(fileID uint) ([]model.FileShare, error) {
	var shares []model.FileShare
	err := r.db.Where("file_id = ? AND status = 1", fileID).Find(&shares).Error
	return shares, err
}

// CountActiveByFileID 统计文件的有效分享数量
func (r *FileShareRepo) CountActiveByFileID(fileID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.FileShare{}).Where("file_id = ? AND status = 1", fileID).Count(&count).Error
	return count, err
}

// DeleteByFileID 删除文件的所有分享记录
func (r *FileShareRepo) DeleteByFileID(fileID uint) error {
	return r.db.Where("file_id = ?", fileID).Delete(&model.FileShare{}).Error
}