package service

import (
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"time"
)

// FavoriteNoteService 收藏和笔记服务
type FavoriteNoteService struct {
	repo *repository.FavoriteNoteRepo
}

// NewFavoriteNoteService 创建收藏和笔记服务
func NewFavoriteNoteService() *FavoriteNoteService {
	return &FavoriteNoteService{
		repo: repository.NewFavoriteNoteRepo(database.GetMySQL()),
	}
}

// FavoriteResponse 收藏响应
type FavoriteResponse struct {
	ID           uint      `json:"id"`
	QuestionID   uint      `json:"questionId"`
	Title        string    `json:"title"`
	QuestionType string    `json:"questionType"`
	Difficulty   int       `json:"difficulty"`
	CategoryName string    `json:"categoryName"`
	FavoritedAt  time.Time `json:"favoritedAt"`
}

// NoteRequest 笔记请求
type NoteRequest struct {
	QuestionID uint   `json:"questionId" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// NoteResponse 笔记响应
type NoteResponse struct {
	ID             uint      `json:"id"`
	QuestionID     uint      `json:"questionId"`
	QuestionTitle  string    `json:"questionTitle"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// AddFavorite 添加收藏
func (s *FavoriteNoteService) AddFavorite(userID, questionID uint) error {
	if s.repo.IsFavorited(userID, questionID) {
		return nil // 已收藏，幂等处理
	}
	return s.repo.AddFavorite(userID, questionID)
}

// RemoveFavorite 取消收藏
func (s *FavoriteNoteService) RemoveFavorite(userID, questionID uint) error {
	return s.repo.RemoveFavorite(userID, questionID)
}

// ListFavorites 获取收藏列表
func (s *FavoriteNoteService) ListFavorites(userID uint, page, pageSize int) ([]FavoriteResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	items, total, err := s.repo.ListFavorites(userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	results := make([]FavoriteResponse, 0, len(items))
	for _, item := range items {
		r := FavoriteResponse{}
		if v, ok := item["id"].(uint); ok {
			r.ID = v
		}
		if v, ok := item["questionId"].(uint); ok {
			r.QuestionID = v
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
		if v, ok := item["categoryName"].(string); ok {
			r.CategoryName = v
		}
		if v, ok := item["createdAt"].(time.Time); ok {
			r.FavoritedAt = v
		}
		results = append(results, r)
	}

	return results, total, nil
}

// CreateNote 创建笔记
func (s *FavoriteNoteService) CreateNote(userID uint, req *NoteRequest) (*NoteResponse, error) {
	// 检查是否已有笔记，有则更新
	existing, err := s.repo.GetNoteByQuestionID(req.QuestionID, userID)
	if err == nil && existing != nil {
		existing.Content = req.Content
		existing.UpdatedAt = time.Now()
		if updateErr := s.repo.UpdateNote(existing); updateErr != nil {
			return nil, updateErr
		}
		return s.toNoteResponse(existing), nil
	}

	note := &model.UserNote{
		UserID:     userID,
		QuestionID: req.QuestionID,
		Content:    req.Content,
	}

	if createErr := s.repo.CreateNote(note); createErr != nil {
		return nil, createErr
	}

	return s.toNoteResponse(note), nil
}

// UpdateNote 更新笔记
func (s *FavoriteNoteService) UpdateNote(noteID, userID uint, req *NoteRequest) (*NoteResponse, error) {
	note, err := s.repo.GetNoteByID(noteID, userID)
	if err != nil {
		return nil, err
	}

	note.Content = req.Content
	note.UpdatedAt = time.Now()

	if updateErr := s.repo.UpdateNote(note); updateErr != nil {
		return nil, updateErr
	}

	return s.toNoteResponse(note), nil
}

// DeleteNote 删除笔记
func (s *FavoriteNoteService) DeleteNote(noteID, userID uint) error {
	return s.repo.DeleteNote(noteID, userID)
}

// ListNotes 获取笔记列表
func (s *FavoriteNoteService) ListNotes(userID uint, page, pageSize int) ([]NoteResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	items, total, err := s.repo.ListNotes(userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	results := make([]NoteResponse, 0, len(items))
	for _, item := range items {
		r := NoteResponse{}
		if v, ok := item["id"].(uint); ok {
			r.ID = v
		}
		if v, ok := item["questionId"].(uint); ok {
			r.QuestionID = v
		}
		if v, ok := item["questionTitle"].(string); ok {
			r.QuestionTitle = v
		}
		if v, ok := item["content"].(string); ok {
			r.Content = v
		}
		if v, ok := item["createdAt"].(time.Time); ok {
			r.CreatedAt = v
		}
		if v, ok := item["updatedAt"].(time.Time); ok {
			r.UpdatedAt = v
		}
		results = append(results, r)
	}

	return results, total, nil
}

// toNoteResponse 转换为笔记响应
func (s *FavoriteNoteService) toNoteResponse(note *model.UserNote) *NoteResponse {
	return &NoteResponse{
		ID:        note.ID,
		QuestionID: note.QuestionID,
		Content:   note.Content,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}
}
