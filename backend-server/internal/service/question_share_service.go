package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

type QuestionShareService struct {
	repo *repository.QuestionShareRepo
}

func NewQuestionShareService() *QuestionShareService {
	return &QuestionShareService{
		repo: repository.NewQuestionShareRepo(database.GetMySQL()),
	}
}

type ShareRequest struct {
	QuestionID      uint   `json:"questionId" binding:"required"`
	ShareType       string `json:"shareType" binding:"required"`
	TargetID        uint   `json:"targetId"`
	ExpireHours     int    `json:"expireHours"`
	MaxAccess       int    `json:"maxAccess"`
}

type ShareResponse struct {
	ID              uint   `json:"id"`
	QuestionID      uint   `json:"questionId"`
	QuestionVersionID uint `json:"questionVersionId"`
	ShareCode       string `json:"shareCode"`
	ShareType       string `json:"shareType"`
	TargetID        uint   `json:"targetId"`
	ExpireAt        string `json:"expireAt"`
	MaxAccess       int    `json:"maxAccess"`
	AccessCount     int    `json:"accessCount"`
	Status          int    `json:"status"`
	CreatedBy       uint   `json:"createdBy"`
	CreatedAt       string `json:"createTime"`
	AccessedAt      string `json:"accessedAt"`
}

func (s *QuestionShareService) List(page, pageSize int, filters map[string]interface{}) ([]ShareResponse, int64, error) {
	items, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	var resp []ShareResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, total, nil
}

func (s *QuestionShareService) GetByID(id uint) (*ShareResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("分享不存在")
		}
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionShareService) Create(req *ShareRequest, createdBy uint) (*ShareResponse, error) {
	code, err := generateQuestionShareCode()
	if err != nil {
		return nil, fmt.Errorf("生成分享码失败")
	}

	item := &model.QuestionShare{
		QuestionID:      req.QuestionID,
		QuestionVersionID: 0,
		ShareCode:       code,
		ShareType:       req.ShareType,
		TargetID:        req.TargetID,
		MaxAccess:       req.MaxAccess,
		Status:          1,
		CreatedBy:       createdBy,
	}

	if req.ExpireHours > 0 {
		expireAt := time.Now().Add(time.Duration(req.ExpireHours) * time.Hour)
		item.ExpireAt = &expireAt
	}

	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionShareService) Disable(id uint) (*ShareResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("分享不存在")
		}
		return nil, err
	}

	item.Status = 3
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionShareService) Enable(id uint) (*ShareResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("分享不存在")
		}
		return nil, err
	}

	item.Status = 1
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionShareService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *QuestionShareService) toResponse(item *model.QuestionShare) ShareResponse {
	resp := ShareResponse{
		ID:              item.ID,
		QuestionID:      item.QuestionID,
		QuestionVersionID: item.QuestionVersionID,
		ShareCode:       item.ShareCode,
		ShareType:       item.ShareType,
		TargetID:        item.TargetID,
		MaxAccess:       item.MaxAccess,
		AccessCount:     item.AccessCount,
		Status:          item.Status,
		CreatedBy:       item.CreatedBy,
		CreatedAt:       item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
	if item.ExpireAt != nil {
		resp.ExpireAt = item.ExpireAt.Format("2006-01-02T15:04:05.000-07:00")
	}
	if item.AccessedAt != nil {
		resp.AccessedAt = item.AccessedAt.Format("2006-01-02T15:04:05.000-07:00")
	}
	return resp
}

func generateQuestionShareCode() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
