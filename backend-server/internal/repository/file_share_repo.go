package repository

import (
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
	return r.db.Model(&model.FileShare{}).Where("share_code = ?", code).UpdateColumn("access_count", gorm.Expr("access_count + 1")).Error
}

func (r *FileShareRepo) Delete(id uint) error {
	return r.db.Delete(&model.FileShare{}, id).Error
}