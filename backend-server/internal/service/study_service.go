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

// SmartPracticeRequest 智能练习请求
type SmartPracticeRequest struct {
	Count          int      `json:"count"`
	Categories     []uint   `json:"categories"`
	FocusKnowledge []string `json:"focusKnowledge"`
	Difficulty     int      `json:"difficulty"`
	Mode           string   `json:"mode"` // review / weak / mixed
}

// SmartPracticeResponse 智能练习响应
type SmartPracticeResponse struct {
	Questions []map[string]interface{} `json:"questions"`
	Analysis  *PracticeAnalysis        `json:"analysis,omitempty"`
}

// PracticeAnalysis 练习分析
type PracticeAnalysis struct {
	WeakKnowledge []string `json:"weakKnowledge"`
	Accuracy      float64  `json:"accuracy"`
	SuggestedDiff int      `json:"suggestedDiff"`
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
		Answers:  answersJSON,
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

	// GORM Scan 到 map[string]interface{} 时，数值类型可能为 int64 或其他类型
	// 使用 fmt.Sprintf 转换为字符串再解析，确保兼容性
	if v, exists := item["question_id"]; exists && v != nil {
		var id uint
		switch val := v.(type) {
		case int64:
			id = uint(val)
		case int:
			id = uint(val)
		case uint:
			id = val
		case uint64:
			id = uint(val)
		case float64:
			id = uint(val)
		default:
			// 尝试字符串解析
			fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &id)
		}
		q.ID = id
	} else if v, exists := item["id"]; exists && v != nil {
		var id uint
		switch val := v.(type) {
		case int64:
			id = uint(val)
		case int:
			id = uint(val)
		case uint:
			id = val
		case uint64:
			id = uint(val)
		case float64:
			id = uint(val)
		default:
			fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &id)
		}
		q.ID = id
	}
	if v, ok := item["title"].(string); ok {
		q.Title = v
	}
	if v, ok := item["question_type"].(string); ok {
		q.QuestionType = v
	}
	if v, ok := item["difficulty"].(int64); ok {
		q.Difficulty = int(v)
	} else if v, ok := item["difficulty"].(int); ok {
		q.Difficulty = v
	}
	if v, ok := item["category_id"].(int64); ok {
		q.CategoryID = uint(v)
	} else if v, ok := item["category_id"].(uint); ok {
		q.CategoryID = v
	}
	if v, ok := item["category_name"].(string); ok {
		q.CategoryName = v
	}
	if v, ok := item["stem"].(string); ok {
		q.Stem = v
	}
	// content 字段作为 options 返回（前端用 content 解析选项）
	if v, ok := item["content"].(string); ok {
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

// GetSmartPractice 智能练习推荐
func (s *StudyService) GetSmartPractice(userID uint, req *SmartPracticeRequest) (*SmartPracticeResponse, error) {
	if req.Count < 1 || req.Count > 100 {
		req.Count = 20
	}

	filters := make(map[string]interface{})

	// 根据模式筛选
	switch req.Mode {
	case "review":
		// 复习模式：优先错题
		wbRepo := repository.NewWrongBookRepo(database.GetMySQL())
		wrongQs, _ := wbRepo.GetRandomQuestions(userID, req.Count)
		return &SmartPracticeResponse{Questions: wrongQs}, nil
	case "weak":
		// 薄弱模式：从错题中找薄弱知识点
		wbRepo := repository.NewWrongBookRepo(database.GetMySQL())
		wrongQs, _ := wbRepo.GetRandomQuestions(userID, req.Count*2)
		if len(wrongQs) > req.Count {
			wrongQs = wrongQs[:req.Count]
		}
		return &SmartPracticeResponse{Questions: wrongQs}, nil
	case "mixed":
		// 混合模式：错题 + 随机
		wbRepo := repository.NewWrongBookRepo(database.GetMySQL())
		wrongQs, _ := wbRepo.GetRandomQuestions(userID, req.Count/2)
		filters["excludeIDs"] = wrongQs
		restQs, _ := s.studyRepo.GetRandomQuestions(req.Count-len(wrongQs), filters)
		all := append(wrongQs, restQs...)
		return &SmartPracticeResponse{Questions: all}, nil
	default:
		// 默认随机
		if len(req.Categories) > 0 {
			filters["categoryId"] = req.Categories[0]
		}
		if req.Difficulty > 0 {
			filters["difficulty"] = req.Difficulty
		}
		questions, _ := s.studyRepo.GetRandomQuestions(req.Count, filters)
		return &SmartPracticeResponse{Questions: questions}, nil
	}
}

// GetPracticeAnalysis 获取练习分析
func (s *StudyService) GetPracticeAnalysis(userID uint) *PracticeAnalysis {
	analysis := &PracticeAnalysis{
		WeakKnowledge: []string{},
		Accuracy:      0.75,
		SuggestedDiff: 2,
	}

	// 从错题本找薄弱知识点
	wbRepo := repository.NewWrongBookRepo(database.GetMySQL())
	wrongQs, _ := wbRepo.GetRandomQuestions(userID, 50)

	// 统计错误次数最多的知识点
	kpCount := make(map[string]int)
	for _, q := range wrongQs {
		if kp, ok := q["knowledge_points"].(string); ok && kp != "" {
			kpCount[kp]++
		}
	}

	for kp, count := range kpCount {
		if count >= 2 {
			analysis.WeakKnowledge = append(analysis.WeakKnowledge, kp)
		}
	}

	// 建议难度
	if len(analysis.WeakKnowledge) > 5 {
		analysis.SuggestedDiff = 1
		analysis.Accuracy = 0.5
	} else if len(analysis.WeakKnowledge) > 2 {
		analysis.SuggestedDiff = 2
		analysis.Accuracy = 0.7
	} else {
		analysis.SuggestedDiff = 3
		analysis.Accuracy = 0.85
	}

	return analysis
}
