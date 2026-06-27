package repository

import (
	"fmt"
	"strings"

	"backend-server/internal/model"

	"gorm.io/gorm"
)

// StudyRepo 学习相关数据访问
type StudyRepo struct {
	db *gorm.DB
}

// NewStudyRepo 创建学习数据仓库
func NewStudyRepo(db *gorm.DB) *StudyRepo {
	return &StudyRepo{db: db}
}

// ListQuestions 获取题目列表（移动端学习用）
func (r *StudyRepo) ListQuestions(userID, userGroupID uint, userClassIDs []uint, offset, limit int, filters map[string]interface{}) ([]map[string]interface{}, int64, error) {
	var results []map[string]interface{}
	var total int64

	query := r.db.Table("qb_questions q").
		Select("q.id as question_id, q.title, q.question_type, q.difficulty, q.category_id, q.status, "+
			"q.stem, q.content, q.answer, q.analysis, "+
			"kc.name as category_name").
		Joins("LEFT JOIN qb_categories kc ON q.category_id = kc.id").
		Where("q.status = ?", "published").
		Scopes(r.resourceTypeScope(userID, userGroupID, userClassIDs))

	// 应用筛选条件
	if questionType, ok := filters["questionType"].(string); ok && questionType != "" {
		query = query.Where("q.question_type = ?", questionType)
	}
	if categoryID, ok := filters["categoryId"].(uint); ok && categoryID > 0 {
		query = query.Where("q.category_id = ?", categoryID)
	}
	if subjectID, ok := filters["subjectId"].(uint); ok && subjectID > 0 {
		query = query.Where("q.subject_id = ?", subjectID)
	}
	if examID, ok := filters["examId"].(uint); ok && examID > 0 {
		query = query.Where("q.exam_id = ?", examID)
	}
	if examCategoryID, ok := filters["examCategoryId"].(uint); ok && examCategoryID > 0 {
		query = query.Where("q.exam_id IN (SELECT id FROM qb_exams WHERE exam_category_id = ?)", examCategoryID)
	}
	if classID, ok := filters["classId"].(uint); ok && classID > 0 {
		query = query.Where("q.resource_type = 'class' AND JSON_CONTAINS(q.class_ids, JSON_ARRAY(?))", classID)
	}
	if difficulty, ok := filters["difficulty"].(int); ok && difficulty > 0 {
		query = query.Where("q.difficulty = ?", difficulty)
	}
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		query = query.Where("q.title LIKE ? OR q.stem LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if knowledgePoint, ok := filters["knowledgePoint"].(string); ok && knowledgePoint != "" {
		query = query.Where("q.knowledge_points LIKE ?", "%"+knowledgePoint+"%")
	}

	// 统计总数
	countQuery := query
	countQuery.Count(&total)

	// 分页查询
	query.Order("q.id DESC").Offset(offset).Limit(limit).Scan(&results)

	return results, total, nil
}

// GetQuestionByID 获取题目详情
func (r *StudyRepo) GetQuestionByID(id uint, userID, userGroupID uint, userClassIDs []uint) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := r.db.Table("qb_questions q").
		Select("q.id as question_id, q.title, q.question_type, q.difficulty, q.category_id, "+
			"q.stem, q.content, q.answer, q.analysis, "+
			"kc.name as category_name").
		Joins("LEFT JOIN qb_categories kc ON q.category_id = kc.id").
		Where("q.id = ? AND q.status = ?", id, "published").
		Scopes(r.resourceTypeScope(userID, userGroupID, userClassIDs)).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("题目不存在")
	}
	return result, nil
}

// GetRandomQuestions 获取随机题目
func (r *StudyRepo) GetRandomQuestions(userID, userGroupID uint, userClassIDs []uint, limit int, filters map[string]interface{}) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := r.db.Table("qb_questions q").
		Select("q.id as question_id, q.title, q.question_type, q.difficulty, q.stem, q.content, q.answer, q.analysis, q.category_id").
		Where("q.status = ?", "published").
		Scopes(r.resourceTypeScope(userID, userGroupID, userClassIDs))

	if questionType, ok := filters["questionType"].(string); ok && questionType != "" {
		query = query.Where("q.question_type = ?", questionType)
	}
	if categoryID, ok := filters["categoryId"].(uint); ok && categoryID > 0 {
		query = query.Where("q.category_id = ?", categoryID)
	}
	if subjectID, ok := filters["subjectId"].(uint); ok && subjectID > 0 {
		query = query.Where("q.subject_id = ?", subjectID)
	}
	if examID, ok := filters["examId"].(uint); ok && examID > 0 {
		query = query.Where("q.exam_id = ?", examID)
	}
	if examCategoryID, ok := filters["examCategoryId"].(uint); ok && examCategoryID > 0 {
		query = query.Where("q.exam_id IN (SELECT id FROM qb_exams WHERE exam_category_id = ?)", examCategoryID)
	}
	if classID, ok := filters["classId"].(uint); ok && classID > 0 {
		query = query.Where("q.resource_type = 'class' AND JSON_CONTAINS(q.class_ids, JSON_ARRAY(?))", classID)
	}
	if difficulty, ok := filters["difficulty"].(int); ok && difficulty > 0 {
		query = query.Where("q.difficulty = ?", difficulty)
	}
	// 排除指定题目ID
	if excludeIDs, ok := filters["excludeIDs"].([]uint); ok && len(excludeIDs) > 0 {
		query = query.Where("q.id NOT IN ?", excludeIDs)
	}

	query.Order("RAND()").Limit(limit).Scan(&results)

	return results, nil
}

// resourceTypeScope 题目资源类型可见性过滤
// public: 所有人可见；private: 仅创建者；group: 分组内成员或创建者；
// class: 班级成员或创建者；user: 指定用户或创建者
func (r *StudyRepo) resourceTypeScope(userID, userGroupID uint, userClassIDs []uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		conditions := []string{
			"q.resource_type = 'public'",
			"q.resource_type = 'private' AND q.created_by = ?",
			"q.resource_type = 'group' AND (q.group_id = ? OR q.created_by = ?)",
		}
		args := []interface{}{userID, userGroupID, userID}

		if len(userClassIDs) > 0 {
			classCondition := "q.resource_type = 'class' AND ("
			for i, classID := range userClassIDs {
				if i > 0 {
					classCondition += " OR "
				}
				classCondition += "JSON_CONTAINS(q.class_ids, JSON_ARRAY(?))"
				args = append(args, classID)
			}
			classCondition += " OR q.created_by = ?)"
			args = append(args, userID)
			conditions = append(conditions, classCondition)
		} else {
			conditions = append(conditions, "q.resource_type = 'class' AND q.created_by = ?")
			args = append(args, userID)
		}

		conditions = append(conditions, "q.resource_type = 'user' AND (JSON_CONTAINS(q.user_ids, JSON_ARRAY(?)) OR q.created_by = ?)")
		args = append(args, userID, userID)

		where := "(" + strings.Join(conditions, ") OR (") + ")"
		return db.Where(where, args...)
	}
}

// CreatePracticeRecord 创建练习记录
func (r *StudyRepo) CreatePracticeRecord(record *model.PracticeRecord) error {
	return r.db.Create(record).Error
}

// GetPracticeHistory 获取练习历史
func (r *StudyRepo) GetPracticeHistory(userID uint, offset, limit int) ([]model.PracticeRecord, int64, error) {
	var records []model.PracticeRecord
	var total int64

	r.db.Model(&model.PracticeRecord{}).Where("user_id = ?", userID).Count(&total)

	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&records).Error

	return records, total, err
}
