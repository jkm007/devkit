package repository

import (
	"backend-server/internal/model"

	"gorm.io/gorm"
)

// MenuRepo 菜单仓库
type MenuRepo struct {
	db *gorm.DB
}

// NewMenuRepo 创建菜单仓库
func NewMenuRepo(db *gorm.DB) *MenuRepo {
	return &MenuRepo{db: db}
}

// List 获取所有菜单列表
func (r *MenuRepo) List() ([]model.Menu, error) {
	var menus []model.Menu
	if err := r.db.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

// GetByID 根据 ID 获取菜单
func (r *MenuRepo) GetByID(id uint) (*model.Menu, error) {
	var menu model.Menu
	if err := r.db.Where("id = ?", id).First(&menu).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

// GetByName 根据名称获取菜单
func (r *MenuRepo) GetByName(name string) (*model.Menu, error) {
	var menu model.Menu
	if err := r.db.Where("name = ?", name).First(&menu).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

// GetByPath 根据路径获取菜单
func (r *MenuRepo) GetByPath(path string) (*model.Menu, error) {
	var menu model.Menu
	if err := r.db.Where("path = ?", path).First(&menu).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

// Create 创建菜单
func (r *MenuRepo) Create(menu *model.Menu) error {
	return r.db.Create(menu).Error
}

// Update 更新菜单
func (r *MenuRepo) Update(menu *model.Menu) error {
	return r.db.Save(menu).Error
}

// Delete 删除菜单（软删除）
func (r *MenuRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Menu{}).Error
}

// GetByIDs 根据 ID 列表获取菜单
func (r *MenuRepo) GetByIDs(ids []uint) ([]model.Menu, error) {
	var menus []model.Menu
	if err := r.db.Where("id IN ?", ids).Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

// GetChildren 获取子菜单
func (r *MenuRepo) GetChildren(pid uint) ([]model.Menu, error) {
	var menus []model.Menu
	if err := r.db.Where("pid = ?", pid).Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}
