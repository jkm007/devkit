package service

import (
	"errors"
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

type QuestionService struct {
	repo *repository.QuestionRepo
}

func NewQuestionService() *QuestionService {
	return &QuestionService{
		repo: repository.NewQuestionRepo(database.GetMySQL()),
	}
}

type QuestionRequest struct {
	Title                string `json:"title" binding:"required"`
	QuestionType         string `json:"questionType" binding:"required"`
	Stem                 string `json:"stem" binding:"required"`
	Content              string `json:"content"`
	Answer               string `json:"answer"`
	Analysis             string `json:"analysis"`
	Materials            string `json:"materials"`
	ScoreRule            string `json:"scoreRule"`
	ExamID               uint   `json:"examId"`
	SubjectID            uint   `json:"subjectId"`
	CategoryID           uint   `json:"categoryId"`
	SourceID             uint   `json:"sourceId"`
	Difficulty           *int   `json:"difficulty"`
	ResourceType         string `json:"resourceType"`
	AnalysisVisiblePolicy string `json:"analysisVisiblePolicy"`
	AnswerVisiblePolicy  string `json:"answerVisiblePolicy"`
}

type QuestionResponse struct {
	ID                   uint   `json:"id"`
	Title                string `json:"title"`
	QuestionType         string `json:"questionType"`
	Stem                 string `json:"stem"`
	Content              string `json:"content"`
	Answer               string `json:"answer"`
	Analysis             string `json:"analysis"`
	Materials            string `json:"materials"`
	ScoreRule            string `json:"scoreRule"`
	ExamID               uint   `json:"examId"`
	SubjectID            uint   `json:"subjectId"`
	CategoryID           uint   `json:"categoryId"`
	SourceID             uint   `json:"sourceId"`
	Difficulty           int    `json:"difficulty"`
	ResourceType         string `json:"resourceType"`
	Status               string `json:"status"`
	CurrentVersionID     uint   `json:"currentVersionId"`
	ParentID             uint   `json:"parentId"`
	IsGroup              int    `json:"isGroup"`
	SubIndex             int    `json:"subIndex"`
	AnalysisVisiblePolicy string `json:"analysisVisiblePolicy"`
	AnswerVisiblePolicy  string `json:"answerVisiblePolicy"`
	CreatedBy            uint   `json:"createdBy"`
	ReviewedBy           uint   `json:"reviewedBy"`
	ReviewedAt           string `json:"reviewedAt"`
	RejectReason         string `json:"rejectReason"`
	PublishedAt          string `json:"publishedAt"`
	CreatedAt            string `json:"createTime"`
}

func (s *QuestionService) List(page, pageSize int, filters map[string]interface{}) ([]QuestionResponse, int64, error) {
	items, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	var resp []QuestionResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, total, nil
}

func (s *QuestionService) GetByID(id uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func ensureJSON(s string) string {
	if s == "" {
		return "null"
	}
	return s
}

func (s *QuestionService) Create(req *QuestionRequest, createdBy uint) (*QuestionResponse, error) {
	item := &model.Question{
		Title:                req.Title,
		QuestionType:         req.QuestionType,
		Stem:                 ensureJSON(req.Stem),
		Content:              ensureJSON(req.Content),
		Answer:               ensureJSON(req.Answer),
		Analysis:             ensureJSON(req.Analysis),
		Materials:            ensureJSON(req.Materials),
		ScoreRule:            ensureJSON(req.ScoreRule),
		ExamID:               req.ExamID,
		SubjectID:            req.SubjectID,
		CategoryID:           req.CategoryID,
		SourceID:             req.SourceID,
		Difficulty:           1,
		ResourceType:         "private",
		Status:               "draft",
		AnalysisVisiblePolicy: "after_answer",
		AnswerVisiblePolicy:  "after_answer",
		CreatedBy:            createdBy,
	}
	if req.Difficulty != nil {
		item.Difficulty = *req.Difficulty
	}
	if req.ResourceType != "" {
		item.ResourceType = req.ResourceType
	}
	if req.AnalysisVisiblePolicy != "" {
		item.AnalysisVisiblePolicy = req.AnalysisVisiblePolicy
	}
	if req.AnswerVisiblePolicy != "" {
		item.AnswerVisiblePolicy = req.AnswerVisiblePolicy
	}

	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionService) Update(id uint, req *QuestionRequest) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	// 已发布的题目编辑不影响已发布版本
	item.Title = req.Title
	item.QuestionType = req.QuestionType
	item.Stem = ensureJSON(req.Stem)
	item.Content = ensureJSON(req.Content)
	item.Answer = ensureJSON(req.Answer)
	item.Analysis = ensureJSON(req.Analysis)
	item.Materials = ensureJSON(req.Materials)
	item.ScoreRule = ensureJSON(req.ScoreRule)
	item.ExamID = req.ExamID
	item.SubjectID = req.SubjectID
	item.CategoryID = req.CategoryID
	item.SourceID = req.SourceID
	if req.Difficulty != nil {
		item.Difficulty = *req.Difficulty
	}
	if req.ResourceType != "" {
		item.ResourceType = req.ResourceType
	}
	if req.AnalysisVisiblePolicy != "" {
		item.AnalysisVisiblePolicy = req.AnalysisVisiblePolicy
	}
	if req.AnswerVisiblePolicy != "" {
		item.AnswerVisiblePolicy = req.AnswerVisiblePolicy
	}

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *QuestionService) Publish(id uint, reviewedBy uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	if item.Status == "published" {
		return nil, fmt.Errorf("题目已发布")
	}

	// 校验必填字段
	if item.Stem == "" || item.Stem == "{}" || item.Stem == "null" {
		return nil, fmt.Errorf("题干不能为空")
	}

	now := time.Now()
	item.Status = "published"
	item.PublishedAt = &now
	item.ReviewedBy = reviewedBy
	item.ReviewedAt = &now

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionService) Archive(id uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	item.Status = "archived"
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionService) SubmitAudit(id uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	if item.Status != "draft" && item.Status != "rejected" {
		return nil, fmt.Errorf("只有草稿或被驳回的题目才能提交审核")
	}

	item.Status = "pending"
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionService) Approve(id uint, reviewedBy uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	if item.Status != "pending" {
		return nil, fmt.Errorf("只有待审核的题目才能审核")
	}

	now := time.Now()
	item.Status = "published"
	item.ReviewedBy = reviewedBy
	item.ReviewedAt = &now
	item.PublishedAt = &now
	item.RejectReason = ""

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionService) Reject(id uint, reviewedBy uint, reason string) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	if item.Status != "pending" {
		return nil, fmt.Errorf("只有待审核的题目才能驳回")
	}

	now := time.Now()
	item.Status = "rejected"
	item.ReviewedBy = reviewedBy
	item.ReviewedAt = &now
	item.RejectReason = reason

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionService) GetStats() (map[string]interface{}, error) {
	return s.repo.GetStats()
}

func (s *QuestionService) toResponse(item *model.Question) QuestionResponse {
	resp := QuestionResponse{
		ID:                   item.ID,
		Title:                item.Title,
		QuestionType:         item.QuestionType,
		Stem:                 item.Stem,
		Content:              item.Content,
		Answer:               item.Answer,
		Analysis:             item.Analysis,
		Materials:            item.Materials,
		ScoreRule:            item.ScoreRule,
		ExamID:               item.ExamID,
		SubjectID:            item.SubjectID,
		CategoryID:           item.CategoryID,
		SourceID:             item.SourceID,
		Difficulty:           item.Difficulty,
		ResourceType:         item.ResourceType,
		Status:               item.Status,
		CurrentVersionID:     item.CurrentVersionID,
		ParentID:             item.ParentID,
		IsGroup:              item.IsGroup,
		SubIndex:             item.SubIndex,
		AnalysisVisiblePolicy: item.AnalysisVisiblePolicy,
		AnswerVisiblePolicy:  item.AnswerVisiblePolicy,
		CreatedBy:            item.CreatedBy,
		ReviewedBy:           item.ReviewedBy,
		RejectReason:         item.RejectReason,
		CreatedAt:            item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
	if item.ReviewedAt != nil {
		resp.ReviewedAt = item.ReviewedAt.Format("2006-01-02T15:04:05.000-07:00")
	}
	if item.PublishedAt != nil {
		resp.PublishedAt = item.PublishedAt.Format("2006-01-02T15:04:05.000-07:00")
	}
	return resp
}
