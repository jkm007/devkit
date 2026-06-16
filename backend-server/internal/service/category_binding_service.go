package service

import (
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"time"
)

// CategoryBindingService 分类绑定服务
type CategoryBindingService struct {
	repo *repository.CategoryBindingRepo
}

// NewCategoryBindingService 创建分类绑定服务
func NewCategoryBindingService() *CategoryBindingService {
	return &CategoryBindingService{
		repo: repository.NewCategoryBindingRepo(database.GetMySQL()),
	}
}

// CategoryBindingResponse 绑定响应
type CategoryBindingResponse struct {
	ID         uint      `json:"id"`
	CategoryID uint      `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	IsPrimary  bool      `json:"isPrimary"`
	BoundAt    time.Time `json:"boundAt"`
}

// BindCategoryRequest 绑定请求
type BindCategoryRequest struct {
	CategoryID uint `json:"categoryId" binding:"required"`
	IsPrimary  bool `json:"isPrimary"`
}

// ListBindings 获取绑定列表
func (s *CategoryBindingService) ListBindings(userID uint) ([]CategoryBindingResponse, error) {
	// 使用JOIN查询获取分类名称
	type BindingWithCategory struct {
		ID           uint      `json:"id"`
		CategoryID   uint      `json:"categoryId"`
		CategoryName string    `json:"categoryName"`
		IsPrimary    bool      `json:"isPrimary"`
		BoundAt      time.Time `json:"boundAt"`
	}

	var bindings []BindingWithCategory
	err := s.repo.GetDB().
		Table("user_category_bindings b").
		Select("b.id, b.category_id as category_id, c.name as category_name, b.is_primary, b.bound_at").
		Joins("LEFT JOIN qb_categories c ON b.category_id = c.id").
		Where("b.user_id = ?", userID).
		Order("b.is_primary DESC, b.bound_at ASC").
		Scan(&bindings).Error
	if err != nil {
		return nil, err
	}

	results := make([]CategoryBindingResponse, 0, len(bindings))
	for _, b := range bindings {
		results = append(results, CategoryBindingResponse{
			ID:           b.ID,
			CategoryID:   b.CategoryID,
			CategoryName: b.CategoryName,
			IsPrimary:    b.IsPrimary,
			BoundAt:      b.BoundAt,
		})
	}

	return results, nil
}

// BindCategory 绑定分类
func (s *CategoryBindingService) BindCategory(userID uint, req *BindCategoryRequest) error {
	return s.repo.Create(userID, req.CategoryID, req.IsPrimary)
}

// SetPrimary 设为主分类
func (s *CategoryBindingService) SetPrimary(userID, id uint) error {
	return s.repo.SetPrimary(userID, id)
}

// UnbindCategory 解绑
func (s *CategoryBindingService) UnbindCategory(userID, id uint) error {
	return s.repo.Delete(userID, id)
}

// GetBoundCategoryIDs 获取已绑定分类 ID
func (s *CategoryBindingService) GetBoundCategoryIDs(userID uint) ([]uint, error) {
	return s.repo.GetBoundCategoryIDs(userID)
}

// GetPrimaryBinding 获取主分类
func (s *CategoryBindingService) GetPrimaryBinding(userID uint) (*model.UserCategoryBinding, error) {
	bindings, err := s.repo.List(userID)
	if err != nil {
		return nil, err
	}
	for _, b := range bindings {
		if b.IsPrimary {
			return &b, nil
		}
	}
	return nil, nil
}
