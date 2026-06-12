package service

import (
	"errors"
	"fmt"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

type KnowledgePointService struct {
	repo *repository.KnowledgePointRepo
}

func NewKnowledgePointService() *KnowledgePointService {
	return &KnowledgePointService{
		repo: repository.NewKnowledgePointRepo(database.GetMySQL()),
	}
}

type KnowledgePointRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code"`
	ExamID      uint   `json:"examId"`
	SubjectID   uint   `json:"subjectId"`
	CategoryID  uint   `json:"categoryId"`
	ParentID    uint   `json:"parentId"`
	Importance  *int   `json:"importance"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
	Status      *int   `json:"status"`
}

type KnowledgePointResponse struct {
	ID          uint   `json:"id"`
	ExamID      uint   `json:"examId"`
	SubjectID   uint   `json:"subjectId"`
	CategoryID  uint   `json:"categoryId"`
	ParentID    uint   `json:"parentId"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Path        string `json:"path"`
	Level       int    `json:"level"`
	Importance  int    `json:"importance"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
	Status      int    `json:"status"`
	CreatedBy   uint   `json:"createdBy"`
	CreatedAt   string `json:"createTime"`
}

func (s *KnowledgePointService) List(page, pageSize int, filters map[string]interface{}) ([]KnowledgePointResponse, int64, error) {
	items, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	var resp []KnowledgePointResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, total, nil
}

func (s *KnowledgePointService) GetAll() ([]KnowledgePointResponse, error) {
	items, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	var resp []KnowledgePointResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, nil
}

func (s *KnowledgePointService) GetByID(id uint) (*KnowledgePointResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("知识点不存在")
		}
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *KnowledgePointService) Create(req *KnowledgePointRequest, createdBy uint) (*KnowledgePointResponse, error) {
	item := &model.KnowledgePoint{
		Name:        req.Name,
		Code:        req.Code,
		ExamID:      req.ExamID,
		SubjectID:   req.SubjectID,
		CategoryID:  req.CategoryID,
		ParentID:    req.ParentID,
		Importance:  3,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Status:      1,
		CreatedBy:   createdBy,
	}
	if req.Importance != nil {
		item.Importance = *req.Importance
	}
	if req.Status != nil {
		item.Status = *req.Status
	}

	if req.ParentID > 0 {
		parent, err := s.repo.GetByID(req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("父级知识点不存在")
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

func (s *KnowledgePointService) Update(id uint, req *KnowledgePointRequest) (*KnowledgePointResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("知识点不存在")
		}
		return nil, err
	}

	item.Name = req.Name
	item.Code = req.Code
	item.ExamID = req.ExamID
	item.SubjectID = req.SubjectID
	item.CategoryID = req.CategoryID
	item.Description = req.Description
	item.SortOrder = req.SortOrder
	if req.Importance != nil {
		item.Importance = *req.Importance
	}
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
				return nil, fmt.Errorf("父级知识点不存在")
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

func (s *KnowledgePointService) Delete(id uint) error {
	hasChildren, err := s.repo.HasChildren(id)
	if err != nil {
		return err
	}
	if hasChildren {
		return fmt.Errorf("该知识点下有子知识点，无法删除")
	}
	return s.repo.Delete(id)
}

func (s *KnowledgePointService) toResponse(item *model.KnowledgePoint) KnowledgePointResponse {
	return KnowledgePointResponse{
		ID:          item.ID,
		ExamID:      item.ExamID,
		SubjectID:   item.SubjectID,
		CategoryID:  item.CategoryID,
		ParentID:    item.ParentID,
		Name:        item.Name,
		Code:        item.Code,
		Path:        item.Path,
		Level:       item.Level,
		Importance:  item.Importance,
		Description: item.Description,
		SortOrder:   item.SortOrder,
		Status:      item.Status,
		CreatedBy:   item.CreatedBy,
		CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
}
