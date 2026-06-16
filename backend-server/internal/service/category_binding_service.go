package service

import (
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"fmt"
	"time"
)

// CategoryBindingService 用户分类绑定服务
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
	ID               uint      `json:"id"`
	SubjectID        uint      `json:"subjectId"`
	SubjectName      string    `json:"subjectName"`
	ExamName         string    `json:"examName"`
	ExamCategoryName string    `json:"examCategoryName"`
	Path             string    `json:"path"`
	IsPrimary        bool      `json:"isPrimary"`
	BoundAt          time.Time `json:"boundAt"`
}

// BindCategoryRequest 绑定分类请求
type BindCategoryRequest struct {
	SubjectID uint `json:"subjectId" binding:"required"`
	IsPrimary bool `json:"isPrimary"`
}

// ListBindings 获取用户绑定的分类列表
func (s *CategoryBindingService) ListBindings(userID uint) ([]CategoryBindingResponse, error) {
	type BindingWithSubject struct {
		ID               uint      `json:"id"`
		SubjectID        uint      `json:"subjectId"`
		SubjectName      string    `json:"subjectName"`
		ExamName         string    `json:"examName"`
		ExamCategoryName string    `json:"examCategoryName"`
		IsPrimary        bool      `json:"isPrimary"`
		BoundAt          time.Time `json:"boundAt"`
	}
	var bindings []BindingWithSubject

	// JOIN 查询：category_id 存储的是 subject_id (L3)
	err := s.repo.GetDB().
		Table("user_category_bindings b").
		Select(`b.id, b.category_id as subject_id,
			s.name as subject_name,
			e.name as exam_name,
			ec.name as exam_category_name,
			b.is_primary, b.bound_at`).
		Joins("LEFT JOIN qb_subjects s ON b.category_id = s.id").
		Joins("LEFT JOIN qb_exams e ON s.exam_id = e.id").
		Joins("LEFT JOIN qb_exam_categories ec ON e.exam_category_id = ec.id").
		Where("b.user_id = ?", userID).
		Order("b.is_primary DESC, b.bound_at ASC").
		Scan(&bindings).Error
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	result := make([]CategoryBindingResponse, 0, len(bindings))
	for _, b := range bindings {
		path := ""
		if b.ExamCategoryName != "" && b.ExamName != "" && b.SubjectName != "" {
			path = fmt.Sprintf("%s > %s > %s", b.ExamCategoryName, b.ExamName, b.SubjectName)
		}
		result = append(result, CategoryBindingResponse{
			ID:               b.ID,
			SubjectID:        b.SubjectID,
			SubjectName:      b.SubjectName,
			ExamName:         b.ExamName,
			ExamCategoryName: b.ExamCategoryName,
			Path:             path,
			IsPrimary:        b.IsPrimary,
			BoundAt:          b.BoundAt,
		})
	}

	return result, nil
}

// BindCategory 绑定分类
func (s *CategoryBindingService) BindCategory(userID uint, req *BindCategoryRequest) error {
	// 调用 repo.Create，内部会检查数量限制和重复
	return s.repo.Create(userID, req.SubjectID, req.IsPrimary)
}

// UnbindCategory 解绑分类
func (s *CategoryBindingService) UnbindCategory(userID, bindingID uint) error {
	return s.repo.Delete(userID, bindingID)
}

// SetPrimary 设置主分类
func (s *CategoryBindingService) SetPrimary(userID, bindingID uint) error {
	return s.repo.SetPrimary(userID, bindingID)
}

// GetBoundSubjectIDs 获取用户绑定的科目ID列表
func (s *CategoryBindingService) GetBoundSubjectIDs(userID uint) ([]uint, error) {
	return s.repo.GetBoundCategoryIDs(userID)
}
