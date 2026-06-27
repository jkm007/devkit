package service

import (
	"encoding/json"
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
	ID           uint        `json:"id"`
	Title        string      `json:"title"`
	QuestionType string      `json:"questionType"`
	Difficulty   int         `json:"difficulty"`
	CategoryID   uint        `json:"categoryId"`
	CategoryName string      `json:"categoryName"`
	Stem         interface{} `json:"stem"`    // 题干，blocks格式
	Options      interface{} `json:"options"` // 选项，JSON数组格式
	Answer       string      `json:"answer,omitempty"`
	Analysis     string      `json:"analysis,omitempty"`
	IsFavorited  bool        `json:"isFavorited"`
}

// PracticeRequest 练习请求
type PracticeRequest struct {
	Mode           string   `json:"mode" binding:"required"`
	Count          int      `json:"count" binding:"required,min=1,max=100"`
	QuestionID     uint     `json:"questionId"`
	CategoryID     uint     `json:"categoryId"`
	SubjectID      uint     `json:"subjectId"`
	ExamID         uint     `json:"examId"`
	ExamCategoryID uint     `json:"examCategoryId"`
	Difficulty     int      `json:"difficulty"`
	Types          []string `json:"types"`
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
	SubjectIDs     []uint   `json:"subjectIds"`
	FocusKnowledge []string `json:"focusKnowledge"`
	Difficulty     int      `json:"difficulty"`
	Mode           string   `json:"mode"` // review / weak / mixed
}

// SmartPracticeResponse 智能练习响应
type SmartPracticeResponse struct {
	Questions []StudyQuestionResponse `json:"questions"`
	Analysis  *PracticeAnalysis       `json:"analysis,omitempty"`
}

// PracticeAnalysis 练习分析
type PracticeAnalysis struct {
	WeakKnowledge []string `json:"weakKnowledge"`
	Accuracy      float64  `json:"accuracy"`
	SuggestedDiff int      `json:"suggestedDiff"`
	TotalWrong    int      `json:"totalWrong"`
	TotalPractice int      `json:"totalPractice"`
}

// ListQuestions 获取题目列表
func (s *StudyService) ListQuestions(userID, userGroupID uint, userClassIDs []uint, page, pageSize int, filters map[string]interface{}) ([]StudyQuestionResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	items, total, err := s.studyRepo.ListQuestions(userID, userGroupID, userClassIDs, offset, pageSize, filters)
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
func (s *StudyService) GetQuestion(userID, userGroupID uint, userClassIDs []uint, questionID uint) (*StudyQuestionResponse, error) {
	item, err := s.studyRepo.GetQuestionByID(questionID, userID, userGroupID, userClassIDs)
	if err != nil {
		return nil, err
	}

	q := s.toQuestionResponse(item, userID)
	return &q, nil
}

// GetRandomQuestions 获取随机题目（练习用）
func (s *StudyService) GetRandomQuestions(userID, userGroupID uint, userClassIDs []uint, req *PracticeRequest) ([]StudyQuestionResponse, error) {
	filters := make(map[string]interface{})
	if len(req.Types) == 1 {
		filters["questionType"] = req.Types[0]
	}
	if req.CategoryID > 0 {
		filters["categoryId"] = req.CategoryID
	}
	if req.SubjectID > 0 {
		filters["subjectId"] = req.SubjectID
	}
	if req.ExamID > 0 {
		filters["examId"] = req.ExamID
	}
	if req.ExamCategoryID > 0 {
		filters["examCategoryId"] = req.ExamCategoryID
	}
	if req.Difficulty > 0 {
		filters["difficulty"] = req.Difficulty
	}

	rawQuestions, err := s.studyRepo.GetRandomQuestions(userID, userGroupID, userClassIDs, req.Count, filters)
	if err != nil {
		return nil, err
	}

	// 转换为前端期望的格式
	questions := make([]StudyQuestionResponse, 0, len(rawQuestions))
	for _, q := range rawQuestions {
		questions = append(questions, s.toQuestionResponse(q, userID))
	}

	return questions, nil
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
	if v, exists := item["question_id"]; exists && v != nil {
		q.ID = toUint(v)
	} else if v, exists := item["id"]; exists && v != nil {
		q.ID = toUint(v)
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
	// stem 是 HTML 字符串，需要转换为前端期望的 blocks 格式
	if v, ok := item["stem"].(string); ok {
		q.Stem = htmlToBlocks(cleanJSONString(v))
	}
	// content 字段是选项的 JSON 数组: [{"id":"A","text":"xxx"}, ...]
	// 需要解析并转换为前端期望的格式
	if v, ok := item["content"].(string); ok {
		q.Options = parseOptions(v)
	}
	// answer 字段: {"correct":["A"]} 或其他格式
	if v, ok := item["answer"].(string); ok {
		q.Answer = v
	}
	// analysis 字段: {"text":"<p>...</p>","media":"..."} 或纯文本
	if v, ok := item["analysis"].(string); ok {
		q.Analysis = parseAnalysis(v)
	}

	// 检查收藏状态
	q.IsFavorited = false
	if userID > 0 {
		q.IsFavorited = repository.NewFavoriteNoteRepo(database.GetMySQL()).IsFavorited(userID, q.ID)
	}

	return q
}

// toUint 通用的数值转uint方法
func toUint(v interface{}) uint {
	switch val := v.(type) {
	case int64:
		return uint(val)
	case int:
		return uint(val)
	case uint:
		return val
	case uint64:
		return uint(val)
	case float64:
		return uint(val)
	default:
		var id uint
		fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &id)
		return id
	}
}

// cleanJSONString 清理JSON字符串的外层引号
// 数据库中可能存储为 "\"<p>xxx</p>\""，需要去除外层引号
func cleanJSONString(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		// 尝试JSON解析去除引号
		var unquoted string
		if err := json.Unmarshal([]byte(s), &unquoted); err == nil {
			return unquoted
		}
	}
	return s
}

// htmlToBlocks 将HTML字符串转换为前端期望的blocks格式
// 输入: "<p>测试</p>" 或 "<p>段落1</p><p>段落2</p>"
// 输出: {"blocks": [{"type": "text", "content": "测试"}]}
func htmlToBlocks(html string) interface{} {
	if html == "" {
		return map[string]interface{}{
			"blocks": []map[string]interface{}{},
		}
	}

	// 简单解析HTML，提取文本内容
	// 移除HTML标签，保留文本
	blocks := []map[string]interface{}{}

	// 按段落分割
	paragraphs := splitHTMLParagraphs(html)
	for _, p := range paragraphs {
		if p != "" {
			blocks = append(blocks, map[string]interface{}{
				"type":    "text",
				"content": p,
			})
		}
	}

	// 如果没有段落，整个作为一个block
	if len(blocks) == 0 {
		blocks = append(blocks, map[string]interface{}{
			"type":    "text",
			"content": stripHTMLTags(html),
		})
	}

	return map[string]interface{}{
		"blocks": blocks,
	}
}

// splitHTMLParagraphs 按<p>标签分割HTML
func splitHTMLParagraphs(html string) []string {
	var paragraphs []string
	current := ""
	runes := []rune(html)

	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if ch == '<' {
			// 检查是否是 </p>
			if i+3 < len(runes) && string(runes[i:i+4]) == "</p>" {
				if current != "" {
					paragraphs = append(paragraphs, current)
					current = ""
				}
				i += 3
				continue
			}
			// 跳过整个标签
			for i < len(runes) && runes[i] != '>' {
				i++
			}
			continue
		}

		current += string(ch)
	}

	if current != "" {
		paragraphs = append(paragraphs, current)
	}

	return paragraphs
}

// stripHTMLTags 移除HTML标签
func stripHTMLTags(html string) string {
	result := ""
	inTag := false

	for _, ch := range html {
		if ch == '<' {
			inTag = true
		} else if ch == '>' {
			inTag = false
		} else if !inTag {
			result += string(ch)
		}
	}

	return result
}

// parseOptions 解析选项JSON
// 数据库格式: [{"id":"A","text":"xxx"}, ...]
// 前端期望: [{"label":"A","content":{"blocks":[{"type":"text","content":"xxx"}]},"isCorrect":false}]
func parseOptions(content string) interface{} {
	if content == "" {
		return nil
	}

	// 尝试解析为数组
	var options []map[string]interface{}
	if err := json.Unmarshal([]byte(content), &options); err != nil {
		// 解析失败，返回原字符串
		return content
	}

	// 转换格式
	result := make([]map[string]interface{}, 0, len(options))
	for _, opt := range options {
		id, _ := opt["id"].(string)
		text, _ := opt["text"].(string)

		// 构建前端期望的格式
		newOpt := map[string]interface{}{
			"label": id,
			"content": map[string]interface{}{
				"blocks": []map[string]interface{}{
					{
						"type":    "text",
						"content": text,
					},
				},
			},
			"isCorrect": false, // 默认不显示正确答案，由 answer 字段控制
		}
		result = append(result, newOpt)
	}

	return result
}

// parseAnalysis 解析解析字段
// 数据库格式: {"text":"<p>...</p>","media":"<video>...</video><img ...>"} 或纯文本
// 前端期望: 直接显示HTML内容（包含图片和视频）
func parseAnalysis(analysis string) string {
	if analysis == "" {
		return ""
	}

	// 清理可能的外层引号
	analysis = cleanJSONString(analysis)

	// 尝试解析为JSON对象
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(analysis), &obj); err == nil {
		result := ""
		// 提取text字段
		if text, ok := obj["text"].(string); ok && text != "" {
			result = text
		}
		// 提取media字段（图片和视频）
		if media, ok := obj["media"].(string); ok && media != "" {
			result += media
		}
		if result != "" {
			return result
		}
	}

	// 返回原内容
	return analysis
}

// GetSmartPractice 智能练习推荐
func (s *StudyService) GetSmartPractice(userID, userGroupID uint, userClassIDs []uint, req *SmartPracticeRequest) (*SmartPracticeResponse, error) {
	if req.Count < 1 || req.Count > 100 {
		req.Count = 20
	}

	filters := make(map[string]interface{})
	var rawQuestions []map[string]interface{}

	// 根据模式筛选
	switch req.Mode {
	case "review":
		// 复习模式：优先错题
		wbRepo := repository.NewWrongBookRepo(database.GetMySQL())
		rawQuestions, _ = wbRepo.GetRandomQuestions(userID, req.Count)
	case "weak":
		// 薄弱模式：从错题中找薄弱知识点
		wbRepo := repository.NewWrongBookRepo(database.GetMySQL())
		rawQuestions, _ = wbRepo.GetRandomQuestions(userID, req.Count*2)
		if len(rawQuestions) > req.Count {
			rawQuestions = rawQuestions[:req.Count]
		}
	case "mixed":
		// 混合模式：错题 + 随机
		wbRepo := repository.NewWrongBookRepo(database.GetMySQL())
		wrongQs, _ := wbRepo.GetRandomQuestions(userID, req.Count/2)
		rawQuestions = wrongQs
		// 获取排除的题目ID
		excludeIDs := make([]uint, 0, len(wrongQs))
		for _, q := range wrongQs {
			if id, ok := q["question_id"]; ok {
				excludeIDs = append(excludeIDs, toUint(id))
			}
		}
		filters["excludeIDs"] = excludeIDs
		restQs, _ := s.studyRepo.GetRandomQuestions(userID, userGroupID, userClassIDs, req.Count-len(rawQuestions), filters)
		rawQuestions = append(rawQuestions, restQs...)
	default:
		// 默认随机
		if len(req.SubjectIDs) > 0 {
			filters["subjectId"] = req.SubjectIDs[0]
		} else if len(req.Categories) > 0 {
			filters["categoryId"] = req.Categories[0]
		}
		if req.Difficulty > 0 {
			filters["difficulty"] = req.Difficulty
		}
		rawQuestions, _ = s.studyRepo.GetRandomQuestions(userID, userGroupID, userClassIDs, req.Count, filters)
	}

	// 转换为前端期望的格式
	questions := make([]StudyQuestionResponse, 0, len(rawQuestions))
	for _, q := range rawQuestions {
		questions = append(questions, s.toQuestionResponse(q, userID))
	}

	return &SmartPracticeResponse{Questions: questions}, nil
}

// GetPracticeAnalysis 获取练习分析
func (s *StudyService) GetPracticeAnalysis(userID uint) *PracticeAnalysis {
	analysis := &PracticeAnalysis{
		WeakKnowledge: []string{},
		Accuracy:      0,
		SuggestedDiff: 2,
		TotalWrong:    0,
		TotalPractice: 0,
	}

	// 从错题本找薄弱知识点
	wbRepo := repository.NewWrongBookRepo(database.GetMySQL())
	wrongQs, _ := wbRepo.GetRandomQuestions(userID, 100)
	analysis.TotalWrong = len(wrongQs)

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

	// 从练习历史获取总练习题数和正确率
	var totalAnswered, totalCorrect int
	practiceHistory, _, _ := s.studyRepo.GetPracticeHistory(userID, 0, 100)
	for _, p := range practiceHistory {
		totalAnswered += p.Total
		totalCorrect += p.Correct
	}
	analysis.TotalPractice = totalAnswered

	// 计算正确率
	if totalAnswered > 0 {
		analysis.Accuracy = float64(totalCorrect) / float64(totalAnswered)
	}

	// 建议难度
	if analysis.Accuracy < 0.5 {
		analysis.SuggestedDiff = 1
	} else if analysis.Accuracy < 0.8 {
		analysis.SuggestedDiff = 2
	} else {
		analysis.SuggestedDiff = 3
	}

	return analysis
}
