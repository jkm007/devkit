package service

import (
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

// CategoryFavoriteService 分类收藏服务
type CategoryFavoriteService struct {
	repo *repository.CategoryFavoriteRepo
}

// NewCategoryFavoriteService 创建分类收藏服务
func NewCategoryFavoriteService() *CategoryFavoriteService {
	return &CategoryFavoriteService{
		repo: repository.NewCategoryFavoriteRepo(database.GetMySQL()),
	}
}

// CategoryFavoriteRequest 添加分类收藏请求
type CategoryFavoriteRequest struct {
	TargetID   uint   `json:"targetId" binding:"required"`
	TargetType string `json:"targetType" binding:"required,oneof=exam_category exam subject category"`
}

// CategoryFavoriteResponse 分类收藏响应
type CategoryFavoriteResponse struct {
	ID         uint      `json:"id"`
	TargetID   uint      `json:"targetId"`
	TargetType string    `json:"targetType"`
	TargetName string    `json:"targetName"`
	Path       string    `json:"path"`
	CreatedAt  time.Time `json:"createdAt"`
}

// List 获取用户分类收藏列表
func (s *CategoryFavoriteService) List(userID uint, page, pageSize int) ([]CategoryFavoriteResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	items, total, err := s.repo.List(userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	results := make([]CategoryFavoriteResponse, 0, len(items))
	for _, item := range items {
		results = append(results, CategoryFavoriteResponse{
			ID:         item.ID,
			TargetID:   item.TargetID,
			TargetType: item.TargetType,
			TargetName: item.TargetName,
			Path:       item.Path,
			CreatedAt:  item.CreatedAt,
		})
	}

	return results, total, nil
}

// Add 添加分类收藏
func (s *CategoryFavoriteService) Add(userID uint, req *CategoryFavoriteRequest) (*CategoryFavoriteResponse, error) {
	// 检查是否已收藏
	if s.repo.Exists(userID, req.TargetType, req.TargetID) {
		return nil, fmt.Errorf("已收藏该分类")
	}

	// 构建目标名称和路径
	targetName, path, err := s.buildTargetInfo(req.TargetType, req.TargetID)
	if err != nil {
		return nil, err
	}

	item := &model.UserCategoryFavorite{
		UserID:     userID,
		TargetID:   req.TargetID,
		TargetType: req.TargetType,
		TargetName: targetName,
		Path:       path,
	}

	if err := s.repo.Create(item); err != nil {
		if s.repo.IsDuplicateError(err) {
			return nil, fmt.Errorf("已收藏该分类")
		}
		return nil, err
	}

	return &CategoryFavoriteResponse{
		ID:         item.ID,
		TargetID:   item.TargetID,
		TargetType: item.TargetType,
		TargetName: item.TargetName,
		Path:       item.Path,
		CreatedAt:  item.CreatedAt,
	}, nil
}

// Remove 取消分类收藏
func (s *CategoryFavoriteService) Remove(userID, id uint) error {
	return s.repo.Delete(userID, id)
}

// buildTargetInfo 根据目标类型和ID查询名称并构建路径
func (s *CategoryFavoriteService) buildTargetInfo(targetType string, targetID uint) (string, string, error) {
	db := database.GetMySQL()

	switch targetType {
	case "exam_category":
		var ec model.ExamCategory
		if err := db.Where("id = ?", targetID).First(&ec).Error; err != nil {
			return "", "", fmt.Errorf("考试大类不存在")
		}
		return ec.Name, ec.Name, nil

	case "exam":
		var exam model.Exam
		if err := db.Where("id = ?", targetID).First(&exam).Error; err != nil {
			return "", "", fmt.Errorf("考试不存在")
		}
		var ec model.ExamCategory
		name := exam.Name
		path := exam.Name
		if err := db.Where("id = ?", exam.ExamCategoryID).First(&ec).Error; err == nil {
			path = ec.Name + " > " + exam.Name
		}
		return name, path, nil

	case "subject":
		var subject model.Subject
		if err := db.Where("id = ?", targetID).First(&subject).Error; err != nil {
			return "", "", fmt.Errorf("科目不存在")
		}
		var exam model.Exam
		var ec model.ExamCategory
		name := subject.Name
		path := subject.Name
		if err := db.Where("id = ?", subject.ExamID).First(&exam).Error; err == nil {
			path = exam.Name + " > " + subject.Name
			if err := db.Where("id = ?", exam.ExamCategoryID).First(&ec).Error; err == nil {
				path = ec.Name + " > " + exam.Name + " > " + subject.Name
			}
		}
		return name, path, nil

	case "category":
		var category model.Category
		if err := db.Where("id = ?", targetID).First(&category).Error; err != nil {
			return "", "", fmt.Errorf("章节分类不存在")
		}
		var subject model.Subject
		var exam model.Exam
		var ec model.ExamCategory
		name := category.Name
		path := category.Name
		if err := db.Where("id = ?", category.SubjectID).First(&subject).Error; err == nil {
			path = subject.Name + " > " + category.Name
			if err := db.Where("id = ?", subject.ExamID).First(&exam).Error; err == nil {
				path = exam.Name + " > " + subject.Name + " > " + category.Name
				if err := db.Where("id = ?", exam.ExamCategoryID).First(&ec).Error; err == nil {
					path = ec.Name + " > " + exam.Name + " > " + subject.Name + " > " + category.Name
				}
			}
		}
		return name, path, nil

	default:
		return "", "", fmt.Errorf("不支持的目标类型: %s", targetType)
	}
}

