package handler

import (
	"backend-server/pkg/database"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MobileCategoryHandler 移动端分类处理器
type MobileCategoryHandler struct {
	db *gorm.DB
}

// NewMobileCategoryHandler 创建移动端分类处理器
func NewMobileCategoryHandler() *MobileCategoryHandler {
	return &MobileCategoryHandler{
		db: database.GetMySQL(),
	}
}

// SubjectWithCount 带题目数量的科目
type SubjectWithCount struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	QuestionCount int64  `json:"questionCount"`
}

// ExamWithSubjects 带科目的考试
type ExamWithSubjects struct {
	ID       uint               `json:"id"`
	Name     string             `json:"name"`
	Subjects []SubjectWithCount `json:"subjects"`
}

// CategoryTreeResponse 分类树响应
type CategoryTreeResponse struct {
	ID    uint               `json:"id"`
	Name  string             `json:"name"`
	Exams []ExamWithSubjects `json:"exams"`
}

// GetCategoryTree 获取分类树（移动端用）
func (h *MobileCategoryHandler) GetCategoryTree(c *gin.Context) {
	// 1. 获取所有启用的考试大类 (L1)
	type ExamCategory struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	var examCategories []ExamCategory
	if err := h.db.Table("qb_exam_categories").
		Where("status = 1").
		Order("sort_order ASC, id ASC").
		Find(&examCategories).Error; err != nil {
		response.InternalError(c, "获取考试大类失败")
		return
	}

	// 2. 获取所有启用的考试 (L2)
	type Exam struct {
		ID               uint   `json:"id"`
		ExamCategoryID   uint   `json:"examCategoryId"`
		Name             string `json:"name"`
	}
	var exams []Exam
	if err := h.db.Table("qb_exams").
		Where("status = 1").
		Order("sort_order ASC, id ASC").
		Find(&exams).Error; err != nil {
		response.InternalError(c, "获取考试列表失败")
		return
	}

	// 3. 获取所有启用的科目 (L3) 及其题目数量
	type SubjectWithExam struct {
		ID     uint   `json:"id"`
		ExamID uint   `json:"examId"`
		Name   string `json:"name"`
	}
	var subjects []SubjectWithExam
	if err := h.db.Table("qb_subjects").
		Where("status = 1").
		Order("sort_order ASC, id ASC").
		Find(&subjects).Error; err != nil {
		response.InternalError(c, "获取科目列表失败")
		return
	}

	// 4. 统计每个科目的题目数量
	type SubjectCount struct {
		SubjectID uint  `json:"subjectId"`
		Count     int64 `json:"count"`
	}
	var subjectCounts []SubjectCount
	h.db.Table("qb_questions").
		Select("subject_id as subject_id, COUNT(*) as count").
		Where("status = ?", "published").
		Group("subject_id").
		Scan(&subjectCounts)

	// 构建题目数量映射
	countMap := make(map[uint]int64)
	for _, sc := range subjectCounts {
		countMap[sc.SubjectID] = sc.Count
	}

	// 5. 构建考试到科目的映射
	examSubjectMap := make(map[uint][]SubjectWithCount)
	for _, s := range subjects {
		examSubjectMap[s.ExamID] = append(examSubjectMap[s.ExamID], SubjectWithCount{
			ID:            s.ID,
			Name:          s.Name,
			QuestionCount: countMap[s.ID],
		})
	}

	// 6. 构建考试大类到考试的映射
	examCategoryExamMap := make(map[uint][]ExamWithSubjects)
	for _, e := range exams {
		examCategoryExamMap[e.ExamCategoryID] = append(examCategoryExamMap[e.ExamCategoryID], ExamWithSubjects{
			ID:       e.ID,
			Name:     e.Name,
			Subjects: examSubjectMap[e.ID],
		})
	}

	// 7. 组装最终结果
	result := make([]CategoryTreeResponse, 0, len(examCategories))
	for _, ec := range examCategories {
		result = append(result, CategoryTreeResponse{
			ID:    ec.ID,
			Name:  ec.Name,
			Exams: examCategoryExamMap[ec.ID],
		})
	}

	response.Success(c, result)
}
