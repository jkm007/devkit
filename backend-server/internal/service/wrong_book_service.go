package service

import (
	"fmt"

	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"time"
)

// WrongBookService 错题本服务
type WrongBookService struct {
	repo      *repository.WrongBookRepo
	studyRepo *repository.StudyRepo
	classRepo *repository.ClassRepo
}

// NewWrongBookService 创建错题本服务
func NewWrongBookService() *WrongBookService {
	return &WrongBookService{
		repo:      repository.NewWrongBookRepo(database.GetMySQL()),
		studyRepo: repository.NewStudyRepo(database.GetMySQL()),
		classRepo: repository.NewClassRepo(database.GetMySQL()),
	}
}

// WrongBookResponse 错题响应
type WrongBookResponse struct {
	ID           uint      `json:"id"`
	QuestionID   uint      `json:"questionId"`
	CategoryID   uint      `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Title        string    `json:"title"`
	QuestionType string    `json:"questionType"`
	Difficulty   int       `json:"difficulty"`
	WrongCount   int       `json:"wrongCount"`
	LastWrongAt  time.Time `json:"lastWrongAt"`
	IsMastered   bool      `json:"isMastered"`
	MasteredAt   time.Time `json:"masteredAt,omitempty"`
}

// WrongBookStats 错题统计响应
type WrongBookStats struct {
	Total        int64                  `json:"total"`
	Mastered     int64                  `json:"mastered"`
	ThisWeek     int64                  `json:"thisWeek"`
	CategoryDist []CategoryDistribution `json:"categoryDist"`
}

// CategoryDistribution 分类分布
type CategoryDistribution struct {
	CategoryName string `json:"categoryName"`
	CategoryID   uint   `json:"categoryId"`
	Count        int64  `json:"count"`
}

// List 获取错题列表
func (s *WrongBookService) List(userID uint, categoryID uint, isMastered *bool, page, pageSize int) ([]WrongBookResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	items, total, err := s.repo.List(userID, categoryID, isMastered, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	results := make([]WrongBookResponse, 0, len(items))
	for _, item := range items {
		r := WrongBookResponse{}
		// GORM Scan 到 map 时数值类型为 int64，需要兼容处理
		if v, ok := item["id"]; ok {
			switch val := v.(type) {
			case int64:
				r.ID = uint(val)
			case uint:
				r.ID = val
			case uint64:
				r.ID = uint(val)
			}
		}
		if v, ok := item["questionId"]; ok {
			switch val := v.(type) {
			case int64:
				r.QuestionID = uint(val)
			case uint:
				r.QuestionID = val
			case uint64:
				r.QuestionID = uint(val)
			}
		}
		if v, ok := item["categoryId"]; ok {
			switch val := v.(type) {
			case int64:
				r.CategoryID = uint(val)
			case uint:
				r.CategoryID = val
			case uint64:
				r.CategoryID = uint(val)
			}
		}
		if v, ok := item["categoryName"].(string); ok {
			r.CategoryName = v
		}
		if v, ok := item["title"].(string); ok {
			r.Title = v
		}
		if v, ok := item["questionType"].(string); ok {
			r.QuestionType = v
		}
		if v, ok := item["difficulty"]; ok {
			switch val := v.(type) {
			case int64:
				r.Difficulty = int(val)
			case int:
				r.Difficulty = val
			}
		}
		if v, ok := item["wrongCount"]; ok {
			switch val := v.(type) {
			case int64:
				r.WrongCount = int(val)
			case int:
				r.WrongCount = val
			}
		}
		if v, ok := item["isMastered"].(bool); ok {
			r.IsMastered = v
		}

		// 时间字段处理
		if v, ok := item["lastWrongAt"].(time.Time); ok {
			r.LastWrongAt = v
		}
		if v, ok := item["masteredAt"].(time.Time); ok && !v.IsZero() {
			r.MasteredAt = v
		}

		results = append(results, r)
	}

	return results, total, nil
}

// AddOrUpdate 添加或更新错题
func (s *WrongBookService) AddOrUpdate(userID, userGroupID, questionID, categoryID uint) error {
	// 校验题目是否对用户可见（已发布且非私有或本人创建）
	userClassIDs, _ := s.classRepo.ListClassIDsByUserID(userID)
	if _, err := s.studyRepo.GetQuestionByID(questionID, userID, userGroupID, userClassIDs); err != nil {
		return fmt.Errorf("题目不可见或不存在")
	}
	return s.repo.AddOrUpdate(userID, questionID, categoryID)
}

// MarkMastered 标记已掌握
func (s *WrongBookService) MarkMastered(userID, questionID uint) error {
	return s.repo.MarkMastered(userID, questionID)
}

// BatchMarkMastered 批量标记已掌握
func (s *WrongBookService) BatchMarkMastered(userID uint, questionIDs []uint) error {
	return s.repo.BatchMarkMastered(userID, questionIDs)
}

// Delete 移除错题
func (s *WrongBookService) Delete(userID, questionID uint) error {
	return s.repo.Delete(userID, questionID)
}

// GetRandomQuestions 获取随机错题
func (s *WrongBookService) GetRandomQuestions(userID uint, limit int) ([]map[string]interface{}, error) {
	return s.repo.GetRandomQuestions(userID, limit)
}

// GetStats 获取统计
func (s *WrongBookService) GetStats(userID uint) (*WrongBookStats, error) {
	raw, err := s.repo.GetStats(userID)
	if err != nil {
		return nil, err
	}

	stats := &WrongBookStats{
		Total:    raw["total"].(int64),
		Mastered: raw["mastered"].(int64),
		ThisWeek: raw["thisWeek"].(int64),
	}

	if dist, ok := raw["categoryDist"].([]struct {
		CategoryName string `json:"categoryName"`
		CategoryID   uint   `json:"categoryId"`
		Count        int64  `json:"count"`
	}); ok {
		stats.CategoryDist = make([]CategoryDistribution, len(dist))
		for i, d := range dist {
			stats.CategoryDist[i] = CategoryDistribution{
				CategoryName: d.CategoryName,
				CategoryID:   d.CategoryID,
				Count:        d.Count,
			}
		}
	}

	return stats, nil
}
