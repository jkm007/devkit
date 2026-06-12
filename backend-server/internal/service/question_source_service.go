package service

import (
	"errors"
	"fmt"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

type QuestionSourceService struct {
	repo *repository.QuestionSourceRepo
}

func NewQuestionSourceService() *QuestionSourceService {
	return &QuestionSourceService{
		repo: repository.NewQuestionSourceRepo(database.GetMySQL()),
	}
}

type QuestionSourceRequest struct {
	SourceType string `json:"sourceType" binding:"required"`
	Name       string `json:"name" binding:"required"`
	ExamID     uint   `json:"examId"`
	SubjectID  uint   `json:"subjectId"`
	Year       int    `json:"year"`
	Region     string `json:"region"`
	PaperName  string `json:"paperName"`
	QuestionNo string `json:"questionNo"`
	Copyright  string `json:"copyright"`
}

type QuestionSourceResponse struct {
	ID         uint   `json:"id"`
	SourceType string `json:"sourceType"`
	Name       string `json:"name"`
	ExamID     uint   `json:"examId"`
	SubjectID  uint   `json:"subjectId"`
	Year       int    `json:"year"`
	Region     string `json:"region"`
	PaperName  string `json:"paperName"`
	QuestionNo string `json:"questionNo"`
	Copyright  string `json:"copyright"`
	CreatedBy  uint   `json:"createdBy"`
	CreatedAt  string `json:"createTime"`
}

func (s *QuestionSourceService) List(page, pageSize int, filters map[string]interface{}) ([]QuestionSourceResponse, int64, error) {
	items, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	var resp []QuestionSourceResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, total, nil
}

func (s *QuestionSourceService) GetByID(id uint) (*QuestionSourceResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("来源不存在")
		}
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionSourceService) Create(req *QuestionSourceRequest, createdBy uint) (*QuestionSourceResponse, error) {
	item := &model.QuestionSource{
		SourceType: req.SourceType,
		Name:       req.Name,
		ExamID:     req.ExamID,
		SubjectID:  req.SubjectID,
		Year:       req.Year,
		Region:     req.Region,
		PaperName:  req.PaperName,
		QuestionNo: req.QuestionNo,
		Copyright:  req.Copyright,
		CreatedBy:  createdBy,
	}
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionSourceService) Update(id uint, req *QuestionSourceRequest) (*QuestionSourceResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("来源不存在")
		}
		return nil, err
	}
	item.SourceType = req.SourceType
	item.Name = req.Name
	item.ExamID = req.ExamID
	item.SubjectID = req.SubjectID
	item.Year = req.Year
	item.Region = req.Region
	item.PaperName = req.PaperName
	item.QuestionNo = req.QuestionNo
	item.Copyright = req.Copyright
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionSourceService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *QuestionSourceService) toResponse(item *model.QuestionSource) QuestionSourceResponse {
	return QuestionSourceResponse{
		ID:         item.ID,
		SourceType: item.SourceType,
		Name:       item.Name,
		ExamID:     item.ExamID,
		SubjectID:  item.SubjectID,
		Year:       item.Year,
		Region:     item.Region,
		PaperName:  item.PaperName,
		QuestionNo: item.QuestionNo,
		Copyright:  item.Copyright,
		CreatedBy:  item.CreatedBy,
		CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
}
