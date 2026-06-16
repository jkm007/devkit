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

// CategoryItem 分类项（通用）
type CategoryItem struct {
	ID            uint           `json:"id"`
	Name          string         `json:"name"`
	QuestionCount int64          `json:"questionCount"`
	Children      []CategoryItem `json:"children,omitempty"`
}

// CategoryTreeNode 分类树节点
type CategoryTreeNode struct {
	ID       uint           `json:"id"`
	Name     string         `json:"name"`
	Exams    []CategoryItem `json:"exams"`
}

// GetCategoryTree 获取分类树（移动端用）
// 返回 L1 > L2 > L3 > L4 完整树结构
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
		ID             uint   `json:"id"`
		ExamCategoryID uint   `json:"examCategoryId"`
		Name           string `json:"name"`
	}
	var exams []Exam
	if err := h.db.Table("qb_exams").
		Where("status = 1").
		Order("sort_order ASC, id ASC").
		Find(&exams).Error; err != nil {
		response.InternalError(c, "获取考试列表失败")
		return
	}

	// 3. 获取所有启用的科目 (L3)
	type Subject struct {
		ID     uint   `json:"id"`
		ExamID uint   `json:"examId"`
		Name   string `json:"name"`
	}
	var subjects []Subject
	if err := h.db.Table("qb_subjects").
		Where("status = 1").
		Order("sort_order ASC, id ASC").
		Find(&subjects).Error; err != nil {
		response.InternalError(c, "获取科目列表失败")
		return
	}

	// 4. 获取所有启用的章节分类 (L4)
	type Category struct {
		ID        uint   `json:"id"`
		SubjectID uint   `json:"subjectId"`
		ExamID    uint   `json:"examId"`
		Name      string `json:"name"`
	}
	var categories []Category
	if err := h.db.Table("qb_categories").
		Where("status = 1 AND deleted_at IS NULL").
		Order("sort_order ASC, id ASC").
		Find(&categories).Error; err != nil {
		response.InternalError(c, "获取章节分类失败")
		return
	}

	// 5. 统计每个科目和章节的题目数量
	type CountResult struct {
		SubjectID  uint  `json:"subjectId"`
		CategoryID uint  `json:"categoryId"`
		Count      int64 `json:"count"`
	}
	var countResults []CountResult
	h.db.Table("qb_questions").
		Select("subject_id as subject_id, category_id as category_id, COUNT(*) as count").
		Where("status = ?", "published").
		Group("subject_id, category_id").
		Scan(&countResults)

	// 构建科目题目数量映射
	subjectCountMap := make(map[uint]int64)
	// 构建章节题目数量映射
	categoryCountMap := make(map[uint]int64)
	for _, cr := range countResults {
		subjectCountMap[cr.SubjectID] += cr.Count
		categoryCountMap[cr.CategoryID] = cr.Count
	}

	// 6. 构建章节到科目的映射（L4 -> L3）
	subjectCategoryMap := make(map[uint][]CategoryItem)
	for _, cat := range categories {
		subjectCategoryMap[cat.SubjectID] = append(subjectCategoryMap[cat.SubjectID], CategoryItem{
			ID:            cat.ID,
			Name:          cat.Name,
			QuestionCount: categoryCountMap[cat.ID],
		})
	}

	// 7. 构建科目到考试的映射（L3 -> L2），包含 L4 子节点
	examSubjectMap := make(map[uint][]CategoryItem)
	for _, s := range subjects {
		examSubjectMap[s.ExamID] = append(examSubjectMap[s.ExamID], CategoryItem{
			ID:            s.ID,
			Name:          s.Name,
			QuestionCount: subjectCountMap[s.ID],
			Children:      subjectCategoryMap[s.ID],
		})
	}

	// 8. 构建考试到考试大类的映射（L2 -> L1）
	examCategoryExamMap := make(map[uint][]CategoryItem)
	for _, e := range exams {
		examCategoryExamMap[e.ExamCategoryID] = append(examCategoryExamMap[e.ExamCategoryID], CategoryItem{
			ID:       e.ID,
			Name:     e.Name,
			Children: examSubjectMap[e.ID],
		})
	}

	// 9. 组装最终结果
	result := make([]CategoryTreeNode, 0, len(examCategories))
	for _, ec := range examCategories {
		result = append(result, CategoryTreeNode{
			ID:    ec.ID,
			Name:  ec.Name,
			Exams: examCategoryExamMap[ec.ID],
		})
	}

	response.Success(c, result)
}
