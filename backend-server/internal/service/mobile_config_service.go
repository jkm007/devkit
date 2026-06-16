package service

import (
	"backend-server/internal/model"
	"backend-server/internal/repository"
)

type MobileConfigService struct {
	repo *repository.MobileConfigRepo
}

func NewMobileConfigService(repo *repository.MobileConfigRepo) *MobileConfigService {
	return &MobileConfigService{repo: repo}
}

// ===== 快捷菜单 =====

func (s *MobileConfigService) GetQuickMenus() ([]model.QuickMenu, error) {
	return s.repo.GetAllQuickMenus()
}

func (s *MobileConfigService) GetActiveQuickMenus() ([]model.QuickMenu, error) {
	return s.repo.GetQuickMenus()
}

func (s *MobileConfigService) CreateQuickMenu(menu *model.QuickMenu) error {
	return s.repo.CreateQuickMenu(menu)
}

func (s *MobileConfigService) UpdateQuickMenu(id uint, data map[string]interface{}) error {
	menu, err := s.repo.GetQuickMenuByID(id)
	if err != nil {
		return err
	}

	if v, ok := data["title"].(string); ok {
		menu.Title = v
	}
	if v, ok := data["icon"].(string); ok {
		menu.Icon = v
	}
	if v, ok := data["link"].(string); ok {
		menu.Link = v
	}
	if v, ok := data["linkType"].(string); ok {
		menu.LinkType = v
	}
	if v, ok := data["sortOrder"].(float64); ok {
		menu.SortOrder = int(v)
	}
	if v, ok := data["status"].(string); ok {
		menu.Status = v
	}

	return s.repo.UpdateQuickMenu(menu)
}

func (s *MobileConfigService) DeleteQuickMenu(id uint) error {
	return s.repo.DeleteQuickMenu(id)
}

// ===== 我的页面菜单 =====

func (s *MobileConfigService) GetMyPageMenus() ([]model.MyPageMenu, error) {
	return s.repo.GetAllMyPageMenus()
}

func (s *MobileConfigService) GetActiveMyPageMenus() ([]model.MyPageMenu, error) {
	return s.repo.GetMyPageMenus()
}

func (s *MobileConfigService) CreateMyPageMenu(menu *model.MyPageMenu) error {
	return s.repo.CreateMyPageMenu(menu)
}

func (s *MobileConfigService) UpdateMyPageMenu(id uint, data map[string]interface{}) error {
	menu, err := s.repo.GetMyPageMenuByID(id)
	if err != nil {
		return err
	}

	if v, ok := data["title"].(string); ok {
		menu.Title = v
	}
	if v, ok := data["icon"].(string); ok {
		menu.Icon = v
	}
	if v, ok := data["link"].(string); ok {
		menu.Link = v
	}
	if v, ok := data["showBadge"].(bool); ok {
		menu.ShowBadge = v
	}
	if v, ok := data["badgeText"].(string); ok {
		menu.BadgeText = v
	}
	if v, ok := data["sortOrder"].(float64); ok {
		menu.SortOrder = int(v)
	}
	if v, ok := data["status"].(string); ok {
		menu.Status = v
	}

	return s.repo.UpdateMyPageMenu(menu)
}

func (s *MobileConfigService) DeleteMyPageMenu(id uint) error {
	return s.repo.DeleteMyPageMenu(id)
}

// ===== 移动端设置 =====

func (s *MobileConfigService) GetMobileSettings() (*model.MobileSettings, error) {
	return s.repo.GetMobileSettings()
}

func (s *MobileConfigService) UpdateMobileSettings(data map[string]interface{}) error {
	settings, err := s.repo.GetMobileSettings()
	if err != nil {
		return err
	}

	if v, ok := data["noticeEnabled"].(bool); ok {
		settings.NoticeEnabled = v
	}
	if v, ok := data["noticeContent"].(string); ok {
		settings.NoticeContent = v
	}
	if v, ok := data["appDownloadUrl"].(string); ok {
		settings.AppDownloadUrl = v
	}
	if v, ok := data["customerServiceUrl"].(string); ok {
		settings.CustomerServiceUrl = v
	}
	if v, ok := data["aboutUs"].(string); ok {
		settings.AboutUs = v
	}
	if v, ok := data["agreementUrl"].(string); ok {
		settings.AgreementUrl = v
	}
	if v, ok := data["privacyUrl"].(string); ok {
		settings.PrivacyUrl = v
	}

	return s.repo.UpdateMobileSettings(settings)
}
