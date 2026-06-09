package service

import (
	"fmt"
	"strings"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

// FileTypeRuleService 文件类型规则服务
type FileTypeRuleService struct {
	repo *repository.FileTypeRuleRepo
}

// NewFileTypeRuleService 创建文件类型规则服务
func NewFileTypeRuleService() *FileTypeRuleService {
	return &FileTypeRuleService{
		repo: repository.NewFileTypeRuleRepo(database.GetMySQL()),
	}
}

// FileTypeRuleGrouped 按类型分组的规则
type FileTypeRuleGrouped struct {
	FileType string                `json:"fileType"`
	Rules    []model.FileTypeRule  `json:"rules"`
}

// GetAll 获取所有规则
func (s *FileTypeRuleService) GetAll() ([]model.FileTypeRule, error) {
	return s.repo.GetAll()
}

// GetGroupedByType 获取按文件类型分组的规则
func (s *FileTypeRuleService) GetGroupedByType() ([]FileTypeRuleGrouped, error) {
	rules, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	groupMap := make(map[string][]model.FileTypeRule)
	for _, rule := range rules {
		groupMap[rule.FileType] = append(groupMap[rule.FileType], rule)
	}

	var groups []FileTypeRuleGrouped
	// 固定顺序：image, video, audio, document, archive, other
	typeOrder := []string{"image", "video", "audio", "document", "archive", "other"}
	for _, t := range typeOrder {
		if rules, ok := groupMap[t]; ok {
			groups = append(groups, FileTypeRuleGrouped{FileType: t, Rules: rules})
			delete(groupMap, t)
		}
	}
	// 追加其他未在固定顺序中的类型
	for t, rules := range groupMap {
		groups = append(groups, FileTypeRuleGrouped{FileType: t, Rules: rules})
	}

	return groups, nil
}

// GetByID 根据 ID 获取规则
func (s *FileTypeRuleService) GetByID(id int64) (*model.FileTypeRule, error) {
	return s.repo.GetByID(id)
}

// Create 创建规则
func (s *FileTypeRuleService) Create(rule *model.FileTypeRule) error {
	// 规范化扩展名
	rule.Extension = normalizeExtension(rule.Extension)

	// 检查扩展名是否已存在
	existing, _ := s.repo.GetByExtension(rule.Extension)
	if existing != nil {
		return fmt.Errorf("扩展名 %s 已存在", rule.Extension)
	}

	// 验证文件类型
	if !isValidFileType(rule.FileType) {
		return fmt.Errorf("无效的文件类型: %s，支持: image, video, audio, document, archive, other", rule.FileType)
	}

	return s.repo.Create(rule)
}

// Update 更新规则
func (s *FileTypeRuleService) Update(rule *model.FileTypeRule) error {
	// 检查是否存在
	existing, err := s.repo.GetByID(rule.ID)
	if err != nil {
		return fmt.Errorf("规则不存在: %d", rule.ID)
	}

	// 规范化扩展名
	rule.Extension = normalizeExtension(rule.Extension)

	// 检查扩展名是否重复（排除自身）
	dup, _ := s.repo.GetByExtension(rule.Extension)
	if dup != nil && dup.ID != rule.ID {
		return fmt.Errorf("扩展名 %s 已存在", rule.Extension)
	}

	// 验证文件类型
	if !isValidFileType(rule.FileType) {
		return fmt.Errorf("无效的文件类型: %s，支持: image, video, audio, document, archive, other", rule.FileType)
	}

	// 保留创建时间
	rule.CreatedAt = existing.CreatedAt

	return s.repo.Update(rule)
}

// Delete 删除规则
func (s *FileTypeRuleService) Delete(id int64) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("规则不存在: %d", id)
	}
	return s.repo.Delete(id)
}

// GetAllEnabled 获取所有启用的规则（供 AutoTagger 使用）
func (s *FileTypeRuleService) GetAllEnabled() ([]model.FileTypeRule, error) {
	return s.repo.GetAllEnabled()
}

// normalizeExtension 规范化扩展名：转小写，确保以 . 开头
func normalizeExtension(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

// isValidFileType 验证文件类型是否有效（允许自定义类型）
func isValidFileType(fileType string) bool {
	fileType = strings.TrimSpace(fileType)
	if fileType == "" {
		return false
	}
	// 只允许字母、数字、下划线、短横线
	for _, c := range fileType {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}
