package service

import (
	"errors"
	"fmt"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

// ==================== 考试大类 ====================

type ExamCategoryService struct {
	repo *repository.ExamCategoryRepo
}

func NewExamCategoryService() *ExamCategoryService {
	return &ExamCategoryService{
		repo: repository.NewExamCategoryRepo(database.GetMySQL()),
	}
}

type ExamCategoryRequest struct {
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code"`
	ParentID  uint   `json:"parentId"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

type ExamCategoryResponse struct {
	ID        uint   `json:"id"`
	ParentID  uint   `json:"parentId"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Path      string `json:"path"`
	Level     int    `json:"level"`
	SortOrder int    `json:"sortOrder"`
	Status    int    `json:"status"`
	CreatedBy uint   `json:"createdBy"`
	CreatedAt string `json:"createTime"`
	Children  []*ExamCategoryResponse `json:"children,omitempty"`
}

func (s *ExamCategoryService) List(page, pageSize int, filters map[string]interface{}) ([]ExamCategoryResponse, int64, error) {
	items, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	var resp []ExamCategoryResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, total, nil
}

func (s *ExamCategoryService) GetAll() ([]ExamCategoryResponse, error) {
	items, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	var resp []ExamCategoryResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, nil
}

func (s *ExamCategoryService) GetByID(id uint) (*ExamCategoryResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("考试大类不存在")
		}
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *ExamCategoryService) Create(req *ExamCategoryRequest, createdBy uint) (*ExamCategoryResponse, error) {
	item := &model.ExamCategory{
		Name:      req.Name,
		Code:      req.Code,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
		Status:    1,
		CreatedBy: createdBy,
	}
	if req.Status != nil {
		item.Status = *req.Status
	}

	// 生成 path 和 level
	if req.ParentID > 0 {
		parent, err := s.repo.GetByID(req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("父级考试大类不存在")
		}
		item.Level = parent.Level + 1
	}

	if err := s.repo.Create(item); err != nil {
		return nil, err
	}

	// 更新 path
	item.Path = fmt.Sprintf("%d", item.ID)
	if item.ParentID > 0 {
		parent, _ := s.repo.GetByID(item.ParentID)
		if parent != nil {
			item.Path = parent.Path + "," + fmt.Sprintf("%d", item.ID)
		}
	}
	_ = s.repo.Update(item)

	resp := s.toResponse(item)
	return &resp, nil
}

func (s *ExamCategoryService) Update(id uint, req *ExamCategoryRequest) (*ExamCategoryResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("考试大类不存在")
		}
		return nil, err
	}

	item.Name = req.Name
	item.Code = req.Code
	item.SortOrder = req.SortOrder
	if req.Status != nil {
		item.Status = *req.Status
	}

	// 如果修改了父级
	if req.ParentID != item.ParentID {
		if req.ParentID == id {
			return nil, fmt.Errorf("不能将自己设为父级")
		}
		item.ParentID = req.ParentID
		if req.ParentID > 0 {
			parent, err := s.repo.GetByID(req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("父级考试大类不存在")
			}
			item.Level = parent.Level + 1
			item.Path = parent.Path + "," + fmt.Sprintf("%d", item.ID)
		} else {
			item.Level = 1
			item.Path = fmt.Sprintf("%d", item.ID)
		}
	}

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *ExamCategoryService) Delete(id uint) error {
	hasChildren, err := s.repo.HasChildren(id)
	if err != nil {
		return err
	}
	if hasChildren {
		return fmt.Errorf("该分类下有子分类，无法删除")
	}
	hasExams, err := s.repo.HasExams(id)
	if err != nil {
		return err
	}
	if hasExams {
		return fmt.Errorf("该分类下有关联考试，无法删除")
	}
	return s.repo.Delete(id)
}

func (s *ExamCategoryService) toResponse(item *model.ExamCategory) ExamCategoryResponse {
	return ExamCategoryResponse{
		ID:        item.ID,
		ParentID:  item.ParentID,
		Name:      item.Name,
		Code:      item.Code,
		Path:      item.Path,
		Level:     item.Level,
		SortOrder: item.SortOrder,
		Status:    item.Status,
		CreatedBy: item.CreatedBy,
		CreatedAt: item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
}

// ==================== 具体考试 ====================

type ExamService struct {
	repo *repository.ExamRepo
}

func NewExamService() *ExamService {
	return &ExamService{
		repo: repository.NewExamRepo(database.GetMySQL()),
	}
}

type ExamRequest struct {
	Name           string `json:"name" binding:"required"`
	Code           string `json:"code"`
	ExamCategoryID uint   `json:"examCategoryId" binding:"required"`
	Description    string `json:"description"`
	SortOrder      int    `json:"sortOrder"`
	Status         *int   `json:"status"`
}

type ExamResponse struct {
	ID             uint   `json:"id"`
	ExamCategoryID uint   `json:"examCategoryId"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	Status         int    `json:"status"`
	SortOrder      int    `json:"sortOrder"`
	CreatedBy      uint   `json:"createdBy"`
	CreatedAt      string `json:"createTime"`
}

func (s *ExamService) List(page, pageSize int, filters map[string]interface{}) ([]ExamResponse, int64, error) {
	items, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	var resp []ExamResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, total, nil
}

func (s *ExamService) GetAll() ([]ExamResponse, error) {
	items, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	var resp []ExamResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, nil
}

func (s *ExamService) GetByCategoryID(categoryId uint) ([]ExamResponse, error) {
	items, err := s.repo.GetByCategoryID(categoryId)
	if err != nil {
		return nil, err
	}
	var resp []ExamResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, nil
}

func (s *ExamService) GetByID(id uint) (*ExamResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("考试不存在")
		}
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *ExamService) Create(req *ExamRequest, createdBy uint) (*ExamResponse, error) {
	item := &model.Exam{
		Name:           req.Name,
		Code:           req.Code,
		ExamCategoryID: req.ExamCategoryID,
		Description:    req.Description,
		SortOrder:      req.SortOrder,
		Status:         1,
		CreatedBy:      createdBy,
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *ExamService) Update(id uint, req *ExamRequest) (*ExamResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("考试不存在")
		}
		return nil, err
	}
	item.Name = req.Name
	item.Code = req.Code
	item.ExamCategoryID = req.ExamCategoryID
	item.Description = req.Description
	item.SortOrder = req.SortOrder
	if req.Status != nil {
		item.Status = *req.Status
	}
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *ExamService) Delete(id uint) error {
	hasSubjects, err := s.repo.HasSubjects(id)
	if err != nil {
		return err
	}
	if hasSubjects {
		return fmt.Errorf("该考试下有科目，无法删除")
	}
	return s.repo.Delete(id)
}

func (s *ExamService) toResponse(item *model.Exam) ExamResponse {
	return ExamResponse{
		ID:             item.ID,
		ExamCategoryID: item.ExamCategoryID,
		Name:           item.Name,
		Code:           item.Code,
		Description:    item.Description,
		Status:         item.Status,
		SortOrder:      item.SortOrder,
		CreatedBy:      item.CreatedBy,
		CreatedAt:      item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
}

// ==================== 科目 ====================

type SubjectService struct {
	repo *repository.SubjectRepo
}

func NewSubjectService() *SubjectService {
	return &SubjectService{
		repo: repository.NewSubjectRepo(database.GetMySQL()),
	}
}

type SubjectRequest struct {
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code"`
	ExamID    uint   `json:"examId" binding:"required"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

type SubjectResponse struct {
	ID        uint   `json:"id"`
	ExamID    uint   `json:"examId"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	SortOrder int    `json:"sortOrder"`
	Status    int    `json:"status"`
	CreatedBy uint   `json:"createdBy"`
	CreatedAt string `json:"createTime"`
}

func (s *SubjectService) List(page, pageSize int, filters map[string]interface{}) ([]SubjectResponse, int64, error) {
	items, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	var resp []SubjectResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, total, nil
}

func (s *SubjectService) GetByExamID(examId uint) ([]SubjectResponse, error) {
	items, err := s.repo.GetByExamID(examId)
	if err != nil {
		return nil, err
	}
	var resp []SubjectResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, nil
}

func (s *SubjectService) GetByID(id uint) (*SubjectResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("科目不存在")
		}
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *SubjectService) Create(req *SubjectRequest, createdBy uint) (*SubjectResponse, error) {
	item := &model.Subject{
		Name:      req.Name,
		Code:      req.Code,
		ExamID:    req.ExamID,
		SortOrder: req.SortOrder,
		Status:    1,
		CreatedBy: createdBy,
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *SubjectService) Update(id uint, req *SubjectRequest) (*SubjectResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("科目不存在")
		}
		return nil, err
	}
	item.Name = req.Name
	item.Code = req.Code
	item.ExamID = req.ExamID
	item.SortOrder = req.SortOrder
	if req.Status != nil {
		item.Status = *req.Status
	}
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *SubjectService) Delete(id uint) error {
	hasCategories, err := s.repo.HasCategories(id)
	if err != nil {
		return err
	}
	if hasCategories {
		return fmt.Errorf("该科目下有分类，无法删除")
	}
	return s.repo.Delete(id)
}

func (s *SubjectService) toResponse(item *model.Subject) SubjectResponse {
	return SubjectResponse{
		ID:        item.ID,
		ExamID:    item.ExamID,
		Name:      item.Name,
		Code:      item.Code,
		SortOrder: item.SortOrder,
		Status:    item.Status,
		CreatedBy: item.CreatedBy,
		CreatedAt: item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
}

// ==================== 章节分类 ====================

type QuestionCategoryService struct {
	repo *repository.QuestionCategoryRepo
}

func NewQuestionCategoryService() *QuestionCategoryService {
	return &QuestionCategoryService{
		repo: repository.NewQuestionCategoryRepo(database.GetMySQL()),
	}
}

type QuestionCategoryRequest struct {
	Name      string `json:"name" binding:"required"`
	ExamID    uint   `json:"examId"`
	SubjectID uint   `json:"subjectId"`
	ParentID  uint   `json:"parentId"`
	SortOrder int    `json:"sortOrder"`
	Status    *int   `json:"status"`
}

type QuestionCategoryResponse struct {
	ID        uint   `json:"id"`
	ExamID    uint   `json:"examId"`
	SubjectID uint   `json:"subjectId"`
	ParentID  uint   `json:"parentId"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Level     int    `json:"level"`
	SortOrder int    `json:"sortOrder"`
	Status    int    `json:"status"`
	CreatedBy uint   `json:"createdBy"`
	CreatedAt string `json:"createTime"`
}

func (s *QuestionCategoryService) List(page, pageSize int, filters map[string]interface{}) ([]QuestionCategoryResponse, int64, error) {
	items, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	var resp []QuestionCategoryResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, total, nil
}

func (s *QuestionCategoryService) GetAll() ([]QuestionCategoryResponse, error) {
	items, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	var resp []QuestionCategoryResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, nil
}

func (s *QuestionCategoryService) GetByID(id uint) (*QuestionCategoryResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("分类不存在")
		}
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionCategoryService) Create(req *QuestionCategoryRequest, createdBy uint) (*QuestionCategoryResponse, error) {
	item := &model.Category{
		Name:      req.Name,
		ExamID:    req.ExamID,
		SubjectID: req.SubjectID,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
		Status:    1,
		CreatedBy: createdBy,
	}
	if req.Status != nil {
		item.Status = *req.Status
	}

	// 生成 path 和 level
	if req.ParentID > 0 {
		parent, err := s.repo.GetByID(req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("父级分类不存在")
		}
		item.Level = parent.Level + 1
	}

	if err := s.repo.Create(item); err != nil {
		return nil, err
	}

	item.Path = fmt.Sprintf("%d", item.ID)
	if item.ParentID > 0 {
		parent, _ := s.repo.GetByID(item.ParentID)
		if parent != nil {
			item.Path = parent.Path + "," + fmt.Sprintf("%d", item.ID)
		}
	}
	_ = s.repo.Update(item)

	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionCategoryService) Update(id uint, req *QuestionCategoryRequest) (*QuestionCategoryResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("分类不存在")
		}
		return nil, err
	}

	item.Name = req.Name
	item.ExamID = req.ExamID
	item.SubjectID = req.SubjectID
	item.SortOrder = req.SortOrder
	if req.Status != nil {
		item.Status = *req.Status
	}

	if req.ParentID != item.ParentID {
		if req.ParentID == id {
			return nil, fmt.Errorf("不能将自己设为父级")
		}
		item.ParentID = req.ParentID
		if req.ParentID > 0 {
			parent, err := s.repo.GetByID(req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("父级分类不存在")
			}
			item.Level = parent.Level + 1
			item.Path = parent.Path + "," + fmt.Sprintf("%d", item.ID)
		} else {
			item.Level = 1
			item.Path = fmt.Sprintf("%d", item.ID)
		}
	}

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionCategoryService) Delete(id uint) error {
	hasChildren, err := s.repo.HasChildren(id)
	if err != nil {
		return err
	}
	if hasChildren {
		return fmt.Errorf("该分类下有子分类，无法删除")
	}
	return s.repo.Delete(id)
}

func (s *QuestionCategoryService) toResponse(item *model.Category) QuestionCategoryResponse {
	return QuestionCategoryResponse{
		ID:        item.ID,
		ExamID:    item.ExamID,
		SubjectID: item.SubjectID,
		ParentID:  item.ParentID,
		Name:      item.Name,
		Path:      item.Path,
		Level:     item.Level,
		SortOrder: item.SortOrder,
		Status:    item.Status,
		CreatedBy: item.CreatedBy,
		CreatedAt: item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
}
