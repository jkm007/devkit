package service

import (
	"errors"
	"fmt"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

type QuestionImportService struct {
	taskRepo *repository.QuestionImportTaskRepo
	itemRepo *repository.QuestionImportItemRepo
}

func NewQuestionImportService() *QuestionImportService {
	return &QuestionImportService{
		taskRepo: repository.NewQuestionImportTaskRepo(database.GetMySQL()),
		itemRepo: repository.NewQuestionImportItemRepo(database.GetMySQL()),
	}
}

type ImportTaskRequest struct {
	FileID           uint   `json:"fileId" binding:"required"`
	FileName         string `json:"fileName" binding:"required"`
	FileType         string `json:"fileType" binding:"required"`
	TargetCategoryID uint   `json:"targetCategoryId"`
	TargetResourceType string `json:"targetResourceType"`
	TargetScopeType  string `json:"targetScopeType"`
	TargetScopeID    uint   `json:"targetScopeId"`
}

type ImportTaskResponse struct {
	ID                 uint   `json:"id"`
	FileID             uint   `json:"fileId"`
	FileName           string `json:"fileName"`
	FileType           string `json:"fileType"`
	Status             string `json:"status"`
	TotalCount         int    `json:"totalCount"`
	SuccessCount       int    `json:"successCount"`
	FailedCount        int    `json:"failedCount"`
	ErrorReport        string `json:"errorReport"`
	TargetCategoryID   uint   `json:"targetCategoryId"`
	TargetResourceType string `json:"targetResourceType"`
	TargetScopeType    string `json:"targetScopeType"`
	TargetScopeID      uint   `json:"targetScopeId"`
	CreatedBy          uint   `json:"createdBy"`
	ConfirmedAt        string `json:"confirmedAt"`
	CreatedAt          string `json:"createTime"`
}

type ImportItemResponse struct {
	ID           uint   `json:"id"`
	TaskID       uint   `json:"taskId"`
	RowNo        int    `json:"rowNo"`
	QuestionNo   string `json:"questionNo"`
	ParseStatus  string `json:"parseStatus"`
	QuestionID   uint   `json:"questionId"`
	ErrorMessage string `json:"errorMessage"`
	RawContent   string `json:"rawContent"`
	CreatedAt    string `json:"createTime"`
}

func (s *QuestionImportService) List(page, pageSize int, filters map[string]interface{}) ([]ImportTaskResponse, int64, error) {
	items, total, err := s.taskRepo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	var resp []ImportTaskResponse
	for _, item := range items {
		resp = append(resp, s.toTaskResponse(&item))
	}
	return resp, total, nil
}

func (s *QuestionImportService) GetByID(id uint) (*ImportTaskResponse, error) {
	item, err := s.taskRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("导入任务不存在")
		}
		return nil, err
	}
	resp := s.toTaskResponse(item)
	return &resp, nil
}

func ensureImportJSON(s string) string {
	if s == "" {
		return "null"
	}
	return s
}

func (s *QuestionImportService) Create(req *ImportTaskRequest, createdBy uint) (*ImportTaskResponse, error) {
	item := &model.QuestionImportTask{
		FileID:             req.FileID,
		FileName:           req.FileName,
		FileType:           req.FileType,
		Status:             "uploaded",
		TargetCategoryID:   req.TargetCategoryID,
		TargetResourceType: "private",
		TargetScopeType:    "user",
		TargetScopeID:      req.TargetScopeID,
		ErrorReport:        "null",
		CreatedBy:          createdBy,
	}
	if req.TargetResourceType != "" {
		item.TargetResourceType = req.TargetResourceType
	}
	if req.TargetScopeType != "" {
		item.TargetScopeType = req.TargetScopeType
	}
	if err := s.taskRepo.Create(item); err != nil {
		return nil, err
	}
	resp := s.toTaskResponse(item)
	return &resp, nil
}

func (s *QuestionImportService) GetItems(taskID uint) ([]ImportItemResponse, error) {
	items, err := s.itemRepo.ListByTaskID(taskID)
	if err != nil {
		return nil, err
	}
	var resp []ImportItemResponse
	for _, item := range items {
		resp = append(resp, s.toItemResponse(&item))
	}
	return resp, nil
}

func (s *QuestionImportService) Delete(id uint) error {
	return s.taskRepo.Delete(id)
}

func (s *QuestionImportService) toTaskResponse(item *model.QuestionImportTask) ImportTaskResponse {
	resp := ImportTaskResponse{
		ID:                 item.ID,
		FileID:             item.FileID,
		FileName:           item.FileName,
		FileType:           item.FileType,
		Status:             item.Status,
		TotalCount:         item.TotalCount,
		SuccessCount:       item.SuccessCount,
		FailedCount:        item.FailedCount,
		ErrorReport:        item.ErrorReport,
		TargetCategoryID:   item.TargetCategoryID,
		TargetResourceType: item.TargetResourceType,
		TargetScopeType:    item.TargetScopeType,
		TargetScopeID:      item.TargetScopeID,
		CreatedBy:          item.CreatedBy,
		CreatedAt:          item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
	if item.ConfirmedAt != nil {
		resp.ConfirmedAt = item.ConfirmedAt.Format("2006-01-02T15:04:05.000-07:00")
	}
	return resp
}

func (s *QuestionImportService) toItemResponse(item *model.QuestionImportItem) ImportItemResponse {
	return ImportItemResponse{
		ID:           item.ID,
		TaskID:       item.TaskID,
		RowNo:        item.RowNo,
		QuestionNo:   item.QuestionNo,
		ParseStatus:  item.ParseStatus,
		QuestionID:   item.QuestionID,
		ErrorMessage: item.ErrorMessage,
		RawContent:   item.RawContent,
		CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
}
