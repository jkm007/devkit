package service

import (
	"fmt"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
	"time"
)

func marshalAnswers(answers []string) string {
	if len(answers) == 0 {
		return "[]"
	}
	result := "["
	for i, a := range answers {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%q", a)
	}
	result += "]"
	return result
}

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

// StudyQuestionResponse 学习用题目响应
type StudyQuestionResponse struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	QuestionType string `json:"questionType"`
	Difficulty   int    `json:"difficulty"`
	CategoryID   uint   `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Stem         string `json:"stem"`
	Options      string `json:"options"`
	Answer       string `json:"answer,omitempty"`
	Analysis     string `json:"analysis,omitempty"`
	IsFavorited  bool   `json:"isFavorited"`
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
func (s *StudyService) ListQuestions(userID uint, page, pageSize int, filters map[string]interface{}) ([]StudyQuestionResponse, int64, error) {
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
	results := make([]StudyQuestionResponse, 0, len(items))
	for _, item := range items {
		q := s.toQuestionResponse(item, userID)
		results = append(results, q)
	}

	return results, total, nil
}

// GetQuestion 获取题目详情
func (s *StudyService) GetQuestion(userID, questionID uint) (*StudyQuestionResponse, error) {
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
	answersJSON := marshalAnswers(req.Answers)

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
func (s *StudyService) toQuestionResponse(item map[string]interface{}, userID uint) StudyQuestionResponse {
	q := StudyQuestionResponse{}

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
		q.Stem = v
	}
	if v, ok := item["options"].(string); ok {
		q.Options = v
	}
	if v, ok := item["answer"].(string); ok {
		q.Answer = v
	}
	if v, ok := item["analysis"].(string); ok {
		q.Analysis = v
	}

	// 检查收藏状态
	q.IsFavorited = false
	if userID > 0 {
		q.IsFavorited = repository.NewFavoriteNoteRepo(database.GetMySQL()).IsFavorited(userID, q.ID)
	}

	return q
}
