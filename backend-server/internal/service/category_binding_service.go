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
	CategoryID       uint      `json:"categoryId"`       // 绑定的ID（可能是科目或章节）
	Level            string    `json:"level"`            // "subject" 或 "category"
	SubjectID        uint      `json:"subjectId"`        // 所属科目ID
	SubjectName      string    `json:"subjectName"`      // 科目名称
	CategoryName     string    `json:"categoryName"`     // 章节名称（如果是L4绑定）
	ExamID           uint      `json:"examId"`           // 所属考试ID
	ExamName         string    `json:"examName"`         // 考试名称
	ExamCategoryID   uint      `json:"examCategoryId"`   // 考试大类ID
	ExamCategoryName string    `json:"examCategoryName"` // 考试大类名称
	Path             string    `json:"path"`             // 完整路径
	IsPrimary        bool      `json:"isPrimary"`
	BoundAt          time.Time `json:"boundAt"`
}

// BindCategoryRequest 绑定分类请求
// 支持绑定到科目(subjectId)或章节(categoryId)
type BindCategoryRequest struct {
	SubjectID  uint `json:"subjectId"`
	CategoryID uint `json:"categoryId"`
	IsPrimary  bool `json:"isPrimary"`
}

// ListBindings 获取用户绑定的分类列表
func (s *CategoryBindingService) ListBindings(userID uint) ([]CategoryBindingResponse, error) {
	type BindingRaw struct {
		ID               uint      `json:"id"`
		CategoryID       uint      `json:"categoryId"`
		IsPrimary        bool      `json:"isPrimary"`
		BoundAt          time.Time `json:"boundAt"`
	}
	var rawBindings []BindingRaw

	// 先获取原始绑定
	err := s.repo.GetDB().
		Table("user_category_bindings").
		Where("user_id = ?", userID).
		Order("is_primary DESC, bound_at ASC").
		Find(&rawBindings).Error
	if err != nil {
		return nil, err
	}

	result := make([]CategoryBindingResponse, 0, len(rawBindings))
	for _, b := range rawBindings {
		resp := CategoryBindingResponse{
			ID:        b.ID,
			CategoryID: b.CategoryID,
			IsPrimary:  b.IsPrimary,
			BoundAt:    b.BoundAt,
		}

		// 先尝试作为科目(L3)查询
		type SubjectInfo struct {
			SubjectID   uint   `json:"subjectId"`
			SubjectName string `json:"subjectName"`
			ExamID      uint   `json:"examId"`
			ExamName    string `json:"examName"`
			ECID        uint   `json:"ecId"`
			ECName      string `json:"ecName"`
		}
		var subjectInfo SubjectInfo
		err := s.repo.GetDB().Raw(`
			SELECT s.id as subject_id, s.name as subject_name,
				e.id as exam_id, e.name as exam_name,
				ec.id as ec_id, ec.name as ec_name
			FROM qb_subjects s
			LEFT JOIN qb_exams e ON s.exam_id = e.id
			LEFT JOIN qb_exam_categories ec ON e.exam_category_id = ec.id
			WHERE s.id = ?
		`, b.CategoryID).Scan(&subjectInfo).Error

		if err == nil && subjectInfo.SubjectID > 0 {
			// 是科目绑定
			resp.Level = "subject"
			resp.SubjectID = subjectInfo.SubjectID
			resp.SubjectName = subjectInfo.SubjectName
			resp.ExamID = subjectInfo.ExamID
			resp.ExamName = subjectInfo.ExamName
			resp.ExamCategoryID = subjectInfo.ECID
			resp.ExamCategoryName = subjectInfo.ECName
			resp.Path = fmt.Sprintf("%s > %s > %s", subjectInfo.ECName, subjectInfo.ExamName, subjectInfo.SubjectName)
		} else {
			// 尝试作为章节(L4)查询
			type CategoryInfo struct {
				CategoryID   uint   `json:"categoryId"`
				CategoryName string `json:"categoryName"`
				SubjectID    uint   `json:"subjectId"`
				SubjectName  string `json:"subjectName"`
				ExamID       uint   `json:"examId"`
				ExamName     string `json:"examName"`
				ECID         uint   `json:"ecId"`
				ECName       string `json:"ecName"`
			}
			var catInfo CategoryInfo
			err := s.repo.GetDB().Raw(`
				SELECT c.id as category_id, c.name as category_name,
					s.id as subject_id, s.name as subject_name,
					e.id as exam_id, e.name as exam_name,
					ec.id as ec_id, ec.name as ec_name
				FROM qb_categories c
				LEFT JOIN qb_subjects s ON c.subject_id = s.id
				LEFT JOIN qb_exams e ON s.exam_id = e.id
				LEFT JOIN qb_exam_categories ec ON e.exam_category_id = ec.id
				WHERE c.id = ?
			`, b.CategoryID).Scan(&catInfo).Error

			if err == nil && catInfo.CategoryID > 0 {
				resp.Level = "category"
				resp.CategoryID = catInfo.CategoryID
				resp.CategoryName = catInfo.CategoryName
				resp.SubjectID = catInfo.SubjectID
				resp.SubjectName = catInfo.SubjectName
				resp.ExamID = catInfo.ExamID
				resp.ExamName = catInfo.ExamName
				resp.ExamCategoryID = catInfo.ECID
				resp.ExamCategoryName = catInfo.ECName
				resp.Path = fmt.Sprintf("%s > %s > %s > %s", catInfo.ECName, catInfo.ExamName, catInfo.SubjectName, catInfo.CategoryName)
			}
		}

		result = append(result, resp)
	}

	return result, nil
}

// BindCategory 绑定分类
func (s *CategoryBindingService) BindCategory(userID uint, req *BindCategoryRequest) error {
	// 优先使用 categoryId，其次使用 subjectId
	bindID := req.CategoryID
	if bindID == 0 {
		bindID = req.SubjectID
	}
	if bindID == 0 {
		return fmt.Errorf("请指定要绑定的分类")
	}
	return s.repo.Create(userID, bindID, req.IsPrimary)
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
