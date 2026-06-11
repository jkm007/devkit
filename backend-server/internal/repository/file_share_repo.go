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

// ShareFilterOptions 分享列表筛选参数
type ShareFilterOptions struct {
	Status  *int   // 状态筛选（nil=全部, 1=有效, 2=已过期, 3=已禁用）
	Keyword string // 搜索关键词（匹配文件名、文件夹名、分享码）
}

// GetUserSharesWithFile 获取用户的分享列表（包含文件信息）
// viewAll=true 时查看所有分享，否则只查看自己的
func (r *FileShareRepo) GetUserSharesWithFile(userID uint, page, pageSize int, viewAll bool, filter *ShareFilterOptions) ([]model.FileShare, int64, error) {
	var shares []model.FileShare
	var total int64

	query := r.db
	if !viewAll {
		query = query.Where("user_id = ?", userID)
	}

	// 应用状态筛选
	if filter != nil && filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// 应用关键词搜索（需要 JOIN 文件表和文件夹表）
	if filter != nil && filter.Keyword != "" {
		likeKeyword := "%" + filter.Keyword + "%"
		query = query.
			Joins("LEFT JOIN sys_file_entries ON sys_file_shares.file_id = sys_file_entries.id").
			Joins("LEFT JOIN sys_file_folders ON sys_file_shares.folder_id = sys_file_folders.id").
			Where("sys_file_entries.name LIKE ? OR sys_file_folders.name LIKE ? OR sys_file_shares.share_code LIKE ?",
				likeKeyword, likeKeyword, likeKeyword)
	}

	// 统计总数
	if err := query.Model(&model.FileShare{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("sys_file_shares.created_at DESC").Offset(offset).Limit(pageSize).Find(&shares).Error; err != nil {
		return nil, 0, err
	}

	return shares, total, nil
}

// ShareStatusCounts 分享状态统计
type ShareStatusCounts struct {
	Total   int64 `json:"total"`
	Active  int64 `json:"active"`
	Expired int64 `json:"expired"`
	Disabled int64 `json:"disabled"`
}

// GetShareStatusCounts 获取各状态的分享数量
func (r *FileShareRepo) GetShareStatusCounts(userID uint, viewAll bool, keyword string) (*ShareStatusCounts, error) {
	counts := &ShareStatusCounts{}

	query := r.db.Model(&model.FileShare{})
	if !viewAll {
		query = query.Where("user_id = ?", userID)
	}

	// 应用关键词搜索
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.
			Joins("LEFT JOIN sys_file_entries ON sys_file_shares.file_id = sys_file_entries.id").
			Joins("LEFT JOIN sys_file_folders ON sys_file_shares.folder_id = sys_file_folders.id").
			Where("sys_file_entries.name LIKE ? OR sys_file_folders.name LIKE ? OR sys_file_shares.share_code LIKE ?",
				likeKeyword, likeKeyword, likeKeyword)
	}

	// 总数
	if err := query.Count(&counts.Total).Error; err != nil {
		return nil, err
	}

	// 各状态数量
	if err := query.Where("status = 1").Count(&counts.Active).Error; err != nil {
		return nil, err
	}
	if err := query.Where("status = 2").Count(&counts.Expired).Error; err != nil {
		return nil, err
	}
	if err := query.Where("status = 3").Count(&counts.Disabled).Error; err != nil {
		return nil, err
	}

	return counts, nil
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

// GetSharedFileIDs 批量获取有活跃分享的文件 ID 集合
func (r *FileShareRepo) GetSharedFileIDs(fileIDs []uint) (map[uint]bool, error) {
	if len(fileIDs) == 0 {
		return nil, nil
	}
	var ids []uint
	err := r.db.Model(&model.FileShare{}).
		Where("file_id IN ? AND status = 1", fileIDs).
		Distinct("file_id").
		Pluck("file_id", &ids).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]bool, len(ids))
	for _, id := range ids {
		result[id] = true
	}
	return result, nil
}