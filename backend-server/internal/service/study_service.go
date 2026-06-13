package service

import (
	"encoding/json"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

// StudyService 学习服务
type StudyService struct {
	studyRepo *repository.StudyRepo
}

// NewStudyService 创建学习服务
func NewStudyService() *StudyService {
	return &StudyService{
		studyRepo: repository.NewStudyRepo(database.GetMySQL()),
	}
}

// QuestionResponse 题目响应
type QuestionResponse struct {
	ID           uint                  `json:"id"`
	Title        string                `json:"title"`
	QuestionType string                `json:"questionType"`
	Difficulty   int                   `json:"difficulty"`
	CategoryID   uint                  `json:"categoryId"`
	CategoryName string                `json:"categoryName"`
	Stem         json.RawMessage       `json:"stem"`
	Options      json.RawMessage       `json:"options"`
	Answer       json.RawMessage       `json:"answer,omitempty"`
	Analysis     json.RawMessage       `json:"analysis,omitempty"`
	IsFavorited  bool                  `json:"isFavorited"`
	Tags         []string              `json:"tags,omitempty"`
}

// PracticeRequest 练习请求
type PracticeRequest struct {
	Mode       string   `json:"mode" binding:"required"`
	Count      int      `json:"count" binding:"required,min=1,max=100"`
	QuestionID uint     `json:"questionId"`
	CategoryID uint     `json:"categoryId"`
	Difficulty int      `json:"difficulty"`
	Types      []string `json:"types"`
}

// PracticeSubmitRequest 练习提交请求
type PracticeSubmitRequest struct {
	Total    int      `json:"total"`
	Answered int      `json:"answered"`
	Correct  int      `json:"correct"`
	Elapsed  int      `json:"elapsed"`
	Answers  []string `json:"answers"`
}

// PracticeHistoryResponse 练习历史响应
type PracticeHistoryResponse struct {
	ID        uint      `json:"id"`
	Mode      string    `json:"mode"`
	Total     int       `json:"total"`
	Answered  int       `json:"answered"`
	Correct   int       `json:"correct"`
	Elapsed   int       `json:"elapsed"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListQuestions 获取题目列表
func (s *StudyService) ListQuestions(userID uint, page, pageSize int, filters map[string]interface{}) ([]QuestionResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	items, total, err := s.studyRepo.ListQuestions(offset, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	results := make([]QuestionResponse, 0, len(items))
	for _, item := range items {
		q := s.toQuestionResponse(item, userID)
		results = append(results, q)
	}

	return results, total, nil
}

// GetQuestion 获取题目详情
func (s *StudyService) GetQuestion(userID, questionID uint) (*QuestionResponse, error) {
	item, err := s.studyRepo.GetQuestionByID(questionID)
	if err != nil {
		return nil, err
	}

	q := s.toQuestionResponse(item, userID)
	return &q, nil
}

// GetRandomQuestions 获取随机题目（练习用）
func (s *StudyService) GetRandomQuestions(req *PracticeRequest) ([]map[string]interface{}, error) {
	filters := make(map[string]interface{})
	if len(req.Types) == 1 {
		filters["questionType"] = req.Types[0]
	}
	if req.CategoryID > 0 {
		filters["categoryId"] = req.CategoryID
	}
	if req.Difficulty > 0 {
		filters["difficulty"] = req.Difficulty
	}

	return s.studyRepo.GetRandomQuestions(req.Count, filters)
}

// SubmitPractice 提交练习结果
func (s *StudyService) SubmitPractice(userID uint, req *PracticeSubmitRequest) error {
	answersJSON, _ := json.Marshal(req.Answers)

	record := &model.PracticeRecord{
		UserID:   userID,
		Mode:     "random",
		Total:    req.Total,
		Answered: req.Answered,
		Correct:  req.Correct,
		Elapsed:  req.Elapsed,
		Answers:  string(answersJSON),
	}

	return s.studyRepo.CreatePracticeRecord(record)
}

// GetPracticeHistory 获取练习历史
func (s *StudyService) GetPracticeHistory(userID uint, page, pageSize int) ([]PracticeHistoryResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	records, total, err := s.studyRepo.GetPracticeHistory(userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	results := make([]PracticeHistoryResponse, 0, len(records))
	for _, r := range records {
		results = append(results, PracticeHistoryResponse{
			ID:        r.ID,
			Mode:      r.Mode,
			Total:     r.Total,
			Answered:  r.Answered,
			Correct:   r.Correct,
			Elapsed:   r.Elapsed,
			CreatedAt: r.CreatedAt,
		})
	}

	return results, total, nil
}

// toQuestionResponse 转换为题目响应
func (s *StudyService) toQuestionResponse(item map[string]interface{}, userID uint) QuestionResponse {
	q := QuestionResponse{}

	if v, ok := item["id"].(uint); ok {
		q.ID = v
	}
	if v, ok := item["title"].(string); ok {
		q.Title = v
	}
	if v, ok := item["question_type"].(string); ok {
		q.QuestionType = v
	}
	if v, ok := item["difficulty"].(int); ok {
		q.Difficulty = v
	}
	if v, ok := item["category_id"].(uint); ok {
		q.CategoryID = v
	}
	if v, ok := item["category_name"].(string); ok {
		q.CategoryName = v
	}
	if v, ok := item["stem"].(string); ok {
		q.Stem = json.RawMessage(v)
	}
	if v, ok := item["options"].(string); ok {
		q.Options = json.RawMessage(v)
	}
	if v, ok := item["answer"].(string); ok {
		q.Answer = json.RawMessage(v)
	}
	if v, ok := item["analysis"].(string); ok {
		q.Analysis = json.RawMessage(v)
	}

	// 检查收藏状态
	q.IsFavorited = false
	if userID > 0 {
		q.IsFavorited = repository.NewFavoriteNoteRepo(database.GetMySQL()).IsFavorited(userID, q.ID)
	}

	return q
}
