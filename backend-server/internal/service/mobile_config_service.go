package service

import (
	"fmt"
	"net/url"
	"strings"

	"backend-server/internal/model"
	"backend-server/internal/repository"
)

// 预定义页面路径（App 已注册页面）
var validPagePaths = map[string]bool{
	"/pages/index/index":      true,
	"/pages/question/list":    true,
	"/pages/question/detail":  true,
	"/pages/question/search":  true,
	"/pages/practice/session": true,
	"/pages/practice/result":  true,
	"/pages/profile/favorites": true,
	"/pages/profile/wrong-book": true,
	"/pages/profile/notes":    true,
	"/pages/profile/history":  true,
	"/pages/class/list":       true,
	"/pages/class/detail":     true,
	"/pages/class/create":     true,
	"/pages/settings/index":   true,
	"/pages/settings/about":   true,
	"/pages/webview/index":    true,
}

// 预定义函数名
var validFunctions = map[string]bool{
	"openCustomerService": true,
	"openAbout":           true,
	"openAgreement":       true,
	"openPrivacy":         true,
	"openFeedback":        true,
	"shareApp":            true,
}

type MobileConfigService struct {
	repo *repository.MobileConfigRepo
}

func NewMobileConfigService(repo *repository.MobileConfigRepo) *MobileConfigService {
	return &MobileConfigService{repo: repo}
}

// validateLink 按 linkType 校验 link 字段
func validateLink(linkType, link string) error {
	if linkType == "none" {
		return nil
	}
	if link == "" {
		return fmt.Errorf("linkType=%s 时 link 不能为空", linkType)
	}
	switch linkType {
	case "page":
		if !validPagePaths[link] {
			return fmt.Errorf("无效的页面路径，可用值: %v", listMapKeys(validPagePaths))
		}
	case "url":
		if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			return fmt.Errorf("URL 必须以 http:// 或 https:// 开头")
		}
		if _, err := url.ParseRequestURI(link); err != nil {
			return fmt.Errorf("无效的 URL 格式")
		}
	case "function":
		if !validFunctions[link] {
			return fmt.Errorf("无效的函数名，可用值: %v", listMapKeys(validFunctions))
		}
	}
	return nil
}

func listMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ===== 快捷菜单 =====

type QuickMenuRequest struct {
	Title     string `json:"title" binding:"required"`
	Icon      string `json:"icon"`
	Link      string `json:"link"`
	LinkType  string `json:"linkType" binding:"required,oneof=none page url function"`
	SortOrder int    `json:"sortOrder"`
	Status    string `json:"status"`
}

func (s *MobileConfigService) GetQuickMenus() ([]model.QuickMenu, error) {
	return s.repo.GetAllQuickMenus()
}

func (s *MobileConfigService) GetActiveQuickMenus() ([]model.QuickMenu, error) {
	return s.repo.GetQuickMenus()
}

func (s *MobileConfigService) CreateQuickMenu(menu *model.QuickMenu) error {
	if err := validateLink(menu.LinkType, menu.Link); err != nil {
		return err
	}
	return s.repo.CreateQuickMenu(menu)
}

func (s *MobileConfigService) UpdateQuickMenu(id uint, req QuickMenuRequest) error {
	menu, err := s.repo.GetQuickMenuByID(id)
	if err != nil {
		return err
	}

	linkType := menu.LinkType
	if req.LinkType != "" {
		linkType = req.LinkType
	}
	link := menu.Link
	if req.Link != "" || req.LinkType != "" {
		link = req.Link
	}
	if err := validateLink(linkType, link); err != nil {
		return err
	}

	if req.Title != "" {
		menu.Title = req.Title
	}
	if req.Link != "" || req.LinkType != "" {
		menu.Link = req.Link
		menu.LinkType = req.LinkType
	}
	if req.Icon != "" {
		menu.Icon = req.Icon
	}
	if req.SortOrder > 0 || req.SortOrder == 0 {
		menu.SortOrder = req.SortOrder
	}
	if req.Status != "" {
		menu.Status = req.Status
	}

	return s.repo.UpdateQuickMenu(menu)
}

func (s *MobileConfigService) DeleteQuickMenu(id uint) error {
	return s.repo.DeleteQuickMenu(id)
}

// ===== 我的页面菜单 =====

type MyPageMenuRequest struct {
	Title     string `json:"title" binding:"required"`
	Icon      string `json:"icon"`
	Link      string `json:"link"`
	LinkType  string `json:"linkType" binding:"required,oneof=none page url function"`
	ShowBadge bool   `json:"showBadge"`
	BadgeText string `json:"badgeText"`
	SortOrder int    `json:"sortOrder"`
	Status    string `json:"status"`
}

func (s *MobileConfigService) GetMyPageMenus() ([]model.MyPageMenu, error) {
	return s.repo.GetAllMyPageMenus()
}

func (s *MobileConfigService) GetActiveMyPageMenus() ([]model.MyPageMenu, error) {
	return s.repo.GetMyPageMenus()
}

func (s *MobileConfigService) CreateMyPageMenu(menu *model.MyPageMenu) error {
	if err := validateLink(menu.LinkType, menu.Link); err != nil {
		return err
	}
	return s.repo.CreateMyPageMenu(menu)
}

func (s *MobileConfigService) UpdateMyPageMenu(id uint, req MyPageMenuRequest) error {
	menu, err := s.repo.GetMyPageMenuByID(id)
	if err != nil {
		return err
	}

	linkType := menu.LinkType
	if req.LinkType != "" {
		linkType = req.LinkType
	}
	link := menu.Link
	if req.Link != "" || req.LinkType != "" {
		link = req.Link
	}
	if err := validateLink(linkType, link); err != nil {
		return err
	}

	if req.Title != "" {
		menu.Title = req.Title
	}
	if req.Link != "" || req.LinkType != "" {
		menu.Link = req.Link
		menu.LinkType = req.LinkType
	}
	if req.Icon != "" {
		menu.Icon = req.Icon
	}
	menu.ShowBadge = req.ShowBadge
	if req.BadgeText != "" {
		menu.BadgeText = req.BadgeText
	}
	if req.SortOrder > 0 || req.SortOrder == 0 {
		menu.SortOrder = req.SortOrder
	}
	if req.Status != "" {
		menu.Status = req.Status
	}

	return s.repo.UpdateMyPageMenu(menu)
}

func (s *MobileConfigService) DeleteMyPageMenu(id uint) error {
	return s.repo.DeleteMyPageMenu(id)
}

// ===== 移动端设置 =====

type MobileSettingsRequest struct {
	NoticeEnabled      *bool  `json:"noticeEnabled"`
	NoticeContent      string `json:"noticeContent"`
	AppDownloadUrl     string `json:"appDownloadUrl"`
	CustomerServiceUrl string `json:"customerServiceUrl"`
	AboutUs            string `json:"aboutUs"`
	AgreementUrl       string `json:"agreementUrl"`
	PrivacyUrl         string `json:"privacyUrl"`
}

func (s *MobileConfigService) GetMobileSettings() (*model.MobileSettings, error) {
	return s.repo.GetMobileSettings()
}

func (s *MobileConfigService) UpdateMobileSettings(req MobileSettingsRequest) error {
	settings, err := s.repo.GetMobileSettings()
	if err != nil {
		return err
	}

	// 校验 URL 类字段
	if req.CustomerServiceUrl != "" {
		if !strings.HasPrefix(req.CustomerServiceUrl, "http://") && !strings.HasPrefix(req.CustomerServiceUrl, "https://") {
			return fmt.Errorf("客服URL 必须以 http:// 或 https:// 开头")
		}
	}
	if req.AppDownloadUrl != "" {
		if !strings.HasPrefix(req.AppDownloadUrl, "http://") && !strings.HasPrefix(req.AppDownloadUrl, "https://") {
			return fmt.Errorf("下载URL 必须以 http:// 或 https:// 开头")
		}
	}
	if req.AgreementUrl != "" {
		if !strings.HasPrefix(req.AgreementUrl, "http://") && !strings.HasPrefix(req.AgreementUrl, "https://") {
			return fmt.Errorf("协议URL 必须以 http:// 或 https:// 开头")
		}
	}
	if req.PrivacyUrl != "" {
		if !strings.HasPrefix(req.PrivacyUrl, "http://") && !strings.HasPrefix(req.PrivacyUrl, "https://") {
			return fmt.Errorf("隐私URL 必须以 http:// 或 https:// 开头")
		}
	}

	if req.NoticeEnabled != nil {
		settings.NoticeEnabled = *req.NoticeEnabled
	}
	if req.NoticeContent != "" {
		settings.NoticeContent = req.NoticeContent
	}
	if req.AppDownloadUrl != "" {
		settings.AppDownloadUrl = req.AppDownloadUrl
	}
	if req.CustomerServiceUrl != "" {
		settings.CustomerServiceUrl = req.CustomerServiceUrl
	}
	if req.AboutUs != "" {
		settings.AboutUs = req.AboutUs
	}
	if req.AgreementUrl != "" {
		settings.AgreementUrl = req.AgreementUrl
	}
	if req.PrivacyUrl != "" {
		settings.PrivacyUrl = req.PrivacyUrl
	}

	return s.repo.UpdateMobileSettings(settings)
}
