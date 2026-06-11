package service

import (
	"encoding/json"
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

// MenuService 菜单服务
type MenuService struct {
	menuRepo *repository.MenuRepo
}

// NewMenuService 创建菜单服务
func NewMenuService() *MenuService {
	return &MenuService{
		menuRepo: repository.NewMenuRepo(database.GetMySQL()),
	}
}

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	Name      string      `json:"name" binding:"required"`
	Path      string      `json:"path"`
	Component string      `json:"component"`
	Type      string      `json:"type" binding:"required"`
	Status    int         `json:"status" binding:"required"`
	PID       uint        `json:"pid"`
	AuthCode  string      `json:"authCode"`
	Sort      int         `json:"sort"`
	Meta      interface{} `json:"meta"`
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Component string      `json:"component"`
	Type      string      `json:"type"`
	Status    int         `json:"status"`
	PID       uint        `json:"pid"`
	AuthCode  string      `json:"authCode"`
	Sort      *int        `json:"sort"`
	Meta      interface{} `json:"meta"`
}

// List 获取菜单列表（树形结构）
func (s *MenuService) List() ([]model.MenuTree, error) {
	menus, err := s.menuRepo.List()
	if err != nil {
		return nil, err
	}

	return s.buildTree(menus, 0), nil
}

// GetAll 获取所有菜单（树形，含按钮，用于 CRUD 管理）
func (s *MenuService) GetAll() ([]*MenuItem, error) {
	menus, err := s.menuRepo.List()
	if err != nil {
		return nil, err
	}
	return s.buildMenuTree(menus, 0, true), nil
}

// GetUserMenus 获取用户菜单（树形，过滤按钮，用于侧边栏渲染）
func (s *MenuService) GetUserMenus(userID uint, userPermissionCodes []string) ([]*MenuItem, error) {
	menus, err := s.menuRepo.List()
	if err != nil {
		return nil, err
	}

	// 构建权限码集合，用于快速查找
	codeSet := make(map[string]bool, len(userPermissionCodes))
	for _, code := range userPermissionCodes {
		codeSet[code] = true
	}

	// 过滤菜单：跳过禁用菜单（Status=0）和无权限菜单
	// 没有 authCode 的菜单（如 Dashboard）对所有用户可见
	var filteredMenus []model.Menu
	for _, menu := range menus {
		if menu.Status == 0 {
			continue
		}
		if menu.AuthCode == "" || codeSet[menu.AuthCode] {
			filteredMenus = append(filteredMenus, menu)
		}
	}

	return s.buildMenuTree(filteredMenus, 0, false), nil
}

// buildMenuTree 构建菜单树
// includeButton: true 包含按钮类型（菜单管理用），false 过滤按钮（侧边栏用）
func (s *MenuService) buildMenuTree(menus []model.Menu, pid uint, includeButton bool) []*MenuItem {
	var trees []*MenuItem
	for _, menu := range menus {
		if menu.PID == pid {
			// 侧边栏模式下，按钮类型不参与路由注册，跳过
			if !includeButton && menu.Type == "button" {
				continue
			}
			item := s.menuToMap(menu)
			children := s.buildMenuTree(menus, menu.ID, includeButton)
			if len(children) > 0 {
				item.Children = children
			} else if menu.Type == "catalog" {
				// 目录类型如果没有可见子菜单，不显示（避免空目录）
				continue
			}
			trees = append(trees, item)
		}
	}
	return trees
}

// i18n 翻译映射（与前端 locale 文件保持一致）
var i18nMap = map[string]map[string]string{
	"zh-CN": {
		"page.dashboard.title":      "概览",
		"page.dashboard.analytics":  "分析页",
		"page.dashboard.workspace":  "工作台",
		"system.title":              "系统管理",
		"system.user.title":         "用户管理",
		"system.role.title":         "角色管理",
		"system.menu.title":         "菜单管理",
		"system.group.title":        "分组管理",
		"common.create":             "新增",
		"common.edit":               "编辑",
		"common.delete":             "删除",
	},
	"en-US": {
		"page.dashboard.title":      "Dashboard",
		"page.dashboard.analytics":  "Analytics",
		"page.dashboard.workspace":  "Workspace",
		"system.title":              "System",
		"system.user.title":         "User Management",
		"system.role.title":         "Role Management",
		"system.menu.title":         "Menu Management",
		"system.group.title":        "Group Management",
		"common.create":             "Create",
		"common.edit":               "Edit",
		"common.delete":             "Delete",
	},
}

// translateTitle 翻译 i18n key
func translateTitle(key string) string {
	// 默认使用中文
	lang := "zh-CN"
	if translations, ok := i18nMap[lang]; ok {
		if translated, ok := translations[key]; ok {
			return translated
		}
	}
	// 如果没有找到翻译，返回原 key
	return key
}

// MenuItem 菜单项（有序结构体，控制 JSON 字段顺序）
type MenuItem struct {
	ID         uint                   `json:"id"`
	PID        uint                   `json:"pid"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Path       string                 `json:"path"`
	Redirect   string                 `json:"redirect,omitempty"`
	Component  interface{}            `json:"component,omitempty"`
	Icon       string                 `json:"icon"`
	AuthCode   string                 `json:"authCode"`
	Sort       int                    `json:"sort"`
	Meta       interface{}            `json:"meta"`
	Status     int                    `json:"status"`
	CreateTime time.Time              `json:"createTime"`
	Children   []*MenuItem            `json:"children,omitempty"`
}

// menuToMap 将单个菜单转换为 MenuItem，meta 从字符串解析为对象
func (s *MenuService) menuToMap(menu model.Menu) *MenuItem {
	item := &MenuItem{
		ID:         menu.ID,
		PID:        menu.PID,
		Name:       menu.Name,
		Type:       menu.Type,
		Path:       menu.Path,
		Icon:       menu.Icon,
		AuthCode:   menu.AuthCode,
		Sort:       menu.Sort,
		Status:     menu.Status,
		CreateTime: menu.CreatedAt,
	}
	// 只在 component 非空时设置（前端需要字符串类型，空时不返回该字段）
	if menu.Component != "" {
		item.Component = menu.Component
	}
	// 解析 meta JSON 字符串为对象
	if menu.Meta != "" {
		var metaObj map[string]interface{}
		if err := json.Unmarshal([]byte(menu.Meta), &metaObj); err == nil {
			if title, ok := metaObj["title"].(string); ok {
				metaObj["title"] = translateTitle(title)
			}
			// 从 meta 中提取 redirect 到顶层
			if redirect, ok := metaObj["redirect"].(string); ok {
				item.Redirect = redirect
				delete(metaObj, "redirect")
			}
			item.Meta = metaObj
		} else {
			item.Meta = menu.Meta
		}
	} else {
		item.Meta = map[string]interface{}{}
	}
	// 将 icon 字段合并到 meta 中（前端期望 meta.icon）
	if menu.Icon != "" {
		if metaMap, ok := item.Meta.(map[string]interface{}); ok {
			metaMap["icon"] = menu.Icon
		}
	}
	return item
}

// GetByID 根据 ID 获取菜单
func (s *MenuService) GetByID(id uint) (*model.Menu, error) {
	return s.menuRepo.GetByID(id)
}

// 有效的菜单类型
var validMenuTypes = map[string]bool{
	"catalog":  true,
	"menu":     true,
	"embedded": true,
	"link":     true,
	"button":   true,
}

// Create 创建菜单
func (s *MenuService) Create(req *CreateMenuRequest) error {
	if !validMenuTypes[req.Type] {
		return fmt.Errorf("无效的菜单类型: %s", req.Type)
	}

	menu := &model.Menu{
		PID:       req.PID,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Type:      req.Type,
		Status:    req.Status,
		AuthCode:  req.AuthCode,
		Sort:      req.Sort,
	}

	// 从 Meta 中提取 icon
	if req.Meta != nil {
		if metaMap, ok := req.Meta.(map[string]interface{}); ok {
			if icon, ok := metaMap["icon"].(string); ok {
				menu.Icon = icon
			}
		}
	}

	menu.Meta = s.metaToString(req.Meta)

	return s.menuRepo.Create(menu)
}

// Update 更新菜单
func (s *MenuService) Update(id uint, req *UpdateMenuRequest) error {
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 使用 map 来更新，这样可以处理零值
	updates := make(map[string]interface{})

	if req.Name != "" {
		menu.Name = req.Name
		updates["name"] = req.Name
	}
	if req.Path != "" {
		menu.Path = req.Path
		updates["path"] = req.Path
	}
	if req.Component != "" {
		menu.Component = req.Component
		updates["component"] = req.Component
	}
	if req.Type != "" {
		menu.Type = req.Type
		updates["type"] = req.Type
	}
	// Status 始终更新（因为 0 也是有效状态）
	updates["status"] = req.Status
	menu.Status = req.Status
	// PID 始终更新（因为 0 也是有效状态）
	updates["pid"] = req.PID
	menu.PID = req.PID
	// AuthCode 允许清空
	updates["auth_code"] = req.AuthCode
	menu.AuthCode = req.AuthCode
	if req.Sort != nil {
		updates["sort"] = *req.Sort
		menu.Sort = *req.Sort
	}
	if req.Meta != nil {
		// 从 Meta 中提取 icon
		if metaMap, ok := req.Meta.(map[string]interface{}); ok {
			if icon, ok := metaMap["icon"].(string); ok {
				menu.Icon = icon
				updates["icon"] = icon
			}
		}
		metaStr := s.metaToString(req.Meta)
		updates["meta"] = metaStr
		menu.Meta = metaStr
	}

	// 直接更新数据库（使用 updates map）
	return s.menuRepo.UpdateFields(id, updates)
}

// metaToString 将 meta 对象转换为 JSON 字符串
func (s *MenuService) metaToString(meta interface{}) string {
	if meta == nil {
		return ""
	}
	switch v := meta.(type) {
	case string:
		return v
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(bytes)
	}
}

// Delete 删除菜单（级联删除子菜单）
func (s *MenuService) Delete(id uint) error {
	// 检查菜单是否存在
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("菜单不存在: %w", err)
	}

	// 查找所有子菜单
	children, err := s.menuRepo.GetChildren(menu.ID)
	if err != nil {
		return fmt.Errorf("查询子菜单失败: %w", err)
	}
	if len(children) > 0 {
		// 级联删除：递归删除当前菜单及其所有子菜单
		return s.menuRepo.DeleteWithChildren(id)
	}

	return s.menuRepo.Delete(id)
}

// CheckNameExists 检查菜单名称是否存在
func (s *MenuService) CheckNameExists(name string, excludeID uint) (bool, error) {
	menu, err := s.menuRepo.GetByName(name)
	if err != nil {
		return false, nil
	}
	if menu.ID == excludeID {
		return false, nil
	}
	return true, nil
}

// CheckPathExists 检查菜单路径是否存在
func (s *MenuService) CheckPathExists(path string, excludeID uint) (bool, error) {
	menu, err := s.menuRepo.GetByPath(path)
	if err != nil {
		return false, nil
	}
	if menu.ID == excludeID {
		return false, nil
	}
	return true, nil
}

// buildTree 构建菜单树
func (s *MenuService) buildTree(menus []model.Menu, pid uint) []model.MenuTree {
	var trees []model.MenuTree
	for _, menu := range menus {
		if menu.PID == pid {
			tree := model.MenuTree{
				Menu: menu,
			}
			children := s.buildTree(menus, menu.ID)
			if len(children) > 0 {
				tree.Children = make([]*model.MenuTree, len(children))
				for i := range children {
					tree.Children[i] = &children[i]
				}
			}
			trees = append(trees, tree)
		}
	}
	return trees
}
