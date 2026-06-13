package service

import (
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"time"
)

// WrongBookService 错题本服务
type WrongBookService struct {
	repo *repository.WrongBookRepo
}

// NewWrongBookService 创建错题本服务
func NewWrongBookService() *WrongBookService {
	return &WrongBookService{
		repo: repository.NewWrongBookRepo(database.GetMySQL()),
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
		if v, ok := item["id"].(uint); ok {
			r.ID = v
		}
		if v, ok := item["questionId"].(uint); ok {
			r.QuestionID = v
		}
		if v, ok := item["categoryId"].(uint); ok {
			r.CategoryID = v
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
		if v, ok := item["difficulty"].(int); ok {
			r.Difficulty = v
		}
		if v, ok := item["wrongCount"].(int); ok {
			r.WrongCount = v
		}
		if v, ok := item["isMastered"].(bool); ok {
			r.IsMastered = v
		}
		results = append(results, r)

		// 时间字段处理
		if v, ok := item["lastWrongAt"].(time.Time); ok {
			r.LastWrongAt = v
		}
		if v, ok := item["masteredAt"].(time.Time); ok && !v.IsZero() {
			r.MasteredAt = v
		}
	}

	return results, total, nil
}

// AddOrUpdate 添加或更新错题
func (s *WrongBookService) AddOrUpdate(userID, questionID, categoryID uint) error {
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
