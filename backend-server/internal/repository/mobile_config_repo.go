package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

type MobileConfigRepo struct {
	db *gorm.DB
}

func NewMobileConfigRepo(db *gorm.DB) *MobileConfigRepo {
	return &MobileConfigRepo{db: db}
}

// ===== 快捷菜单 =====

func (r *MobileConfigRepo) GetQuickMenus() ([]model.QuickMenu, error) {
	var menus []model.QuickMenu
	err := r.db.Where("status = ?", model.MobileConfigEnabled).
		Order("sort_order ASC, id ASC").
		Find(&menus).Error
	return menus, err
}

func (r *MobileConfigRepo) GetAllQuickMenus() ([]model.QuickMenu, error) {
	var menus []model.QuickMenu
	err := r.db.Order("sort_order ASC, id ASC").Find(&menus).Error
	return menus, err
}

func (r *MobileConfigRepo) GetQuickMenuByID(id uint) (*model.QuickMenu, error) {
	var menu model.QuickMenu
	err := r.db.First(&menu, id).Error
	return &menu, err
}

func (r *MobileConfigRepo) CreateQuickMenu(menu *model.QuickMenu) error {
	return r.db.Create(menu).Error
}

func (r *MobileConfigRepo) UpdateQuickMenu(menu *model.QuickMenu) error {
	return r.db.Save(menu).Error
}

func (r *MobileConfigRepo) DeleteQuickMenu(id uint) error {
	return r.db.Delete(&model.QuickMenu{}, id).Error
}

// ===== 我的页面菜单 =====

func (r *MobileConfigRepo) GetMyPageMenus() ([]model.MyPageMenu, error) {
	var menus []model.MyPageMenu
	err := r.db.Where("status = ?", model.MobileConfigEnabled).
		Order("sort_order ASC, id ASC").
		Find(&menus).Error
	return menus, err
}

func (r *MobileConfigRepo) GetAllMyPageMenus() ([]model.MyPageMenu, error) {
	var menus []model.MyPageMenu
	err := r.db.Order("sort_order ASC, id ASC").Find(&menus).Error
	return menus, err
}

func (r *MobileConfigRepo) GetMyPageMenuByID(id uint) (*model.MyPageMenu, error) {
	var menu model.MyPageMenu
	err := r.db.First(&menu, id).Error
	return &menu, err
}

func (r *MobileConfigRepo) CreateMyPageMenu(menu *model.MyPageMenu) error {
	return r.db.Create(menu).Error
}

func (r *MobileConfigRepo) UpdateMyPageMenu(menu *model.MyPageMenu) error {
	return r.db.Save(menu).Error
}

func (r *MobileConfigRepo) DeleteMyPageMenu(id uint) error {
	return r.db.Delete(&model.MyPageMenu{}, id).Error
}

// ===== 移动端设置 =====

func (r *MobileConfigRepo) GetMobileSettings() (*model.MobileSettings, error) {
	var settings model.MobileSettings
	err := r.db.First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		// 如果没有记录，创建默认设置
		settings = model.MobileSettings{}
		err = r.db.Create(&settings).Error
	}
	return &settings, err
}

func (r *MobileConfigRepo) UpdateMobileSettings(settings *model.MobileSettings) error {
	return r.db.Save(settings).Error
}
