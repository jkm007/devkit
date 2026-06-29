package repository

import (
	"backend-server/internal/model"
	"strings"

	"gorm.io/gorm"
)

type QuestionRepo struct {
	db *gorm.DB
}

func NewQuestionRepo(db *gorm.DB) *QuestionRepo {
	return &QuestionRepo{db: db}
}

// DB returns the underlying database connection for custom queries
func (r *QuestionRepo) DB() *gorm.DB {
	return r.db
}

func (r *QuestionRepo) List(page, pageSize int, filters map[string]interface{}) ([]model.Question, int64, error) {
	var items []model.Question
	var total int64
	query := r.db.Model(&model.Question{})

	// 私有题目过滤：管理员/超级管理员可查看全部，普通用户只能看自己的私有题目
	isAdmin := false
	if roles, ok := filters["roles"]; ok && roles != nil {
		// 将 roles 转为字符串切片
		var roleList []string
		switch v := roles.(type) {
		case []string:
			roleList = v
		case []interface{}:
			for _, r := range v {
				if s, ok := r.(string); ok {
					roleList = append(roleList, s)
				}
			}
		}
		for _, role := range roleList {
			if role == "admin" || role == "super_admin" {
				isAdmin = true
				break
			}
		}
	}
	if !isAdmin {
		userId := uint(0)
		if uid, ok := filters["userId"]; ok && uid != nil {
			switch v := uid.(type) {
			case uint:
				userId = v
			case float64:
				userId = uint(v)
			case int:
				userId = uint(v)
			}
		}
		userGroupId := uint(0)
		if gid, ok := filters["userGroupId"]; ok && gid != nil {
			switch v := gid.(type) {
			case uint:
				userGroupId = v
			case float64:
				userGroupId = uint(v)
			case int:
				userGroupId = uint(v)
			}
		}
		var userClassIds []uint
		if cids, ok := filters["userClassIds"].([]uint); ok {
			userClassIds = cids
		}

		query = r.applyVisibilityFilter(query, userId, userGroupId, userClassIds)
	}

	if title, ok := filters["title"]; ok && title != "" {
		query = query.Where("title LIKE ?", "%"+escapeLike(title.(string))+"%")
	}
	if questionType, ok := filters["questionType"]; ok && questionType != "" {
		query = query.Where("question_type = ?", questionType)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if examId, ok := filters["examId"]; ok && examId != "" {
		query = query.Where("exam_id = ?", examId)
	}
	if subjectId, ok := filters["subjectId"]; ok && subjectId != "" {
		query = query.Where("subject_id = ?", subjectId)
	}
	if categoryId, ok := filters["categoryId"]; ok && categoryId != "" {
		query = query.Where("category_id = ?", categoryId)
	}
	if difficulty, ok := filters["difficulty"]; ok && difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if resourceType, ok := filters["resourceType"]; ok && resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if createdBy, ok := filters["createdBy"]; ok && createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *QuestionRepo) Search(page, pageSize int, keyword string, userID uint, userGroupID uint, userClassIDs []uint, status string) ([]model.Question, int64, error) {
	var items []model.Question
	var total int64

	// status 默认为 published（移动端场景），管理员可传入其他状态搜索草稿/待审核等
	if status == "" {
		status = "published"
	}

	query := r.db.Model(&model.Question{}).
		Where("(title LIKE ? OR stem LIKE ?) AND status = ?",
			"%"+escapeLike(keyword)+"%",
			"%"+escapeLike(keyword)+"%",
			status)

	// 非管理员按资源类型过滤可见性
	query = r.applyVisibilityFilter(query, userID, userGroupID, userClassIDs)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *QuestionRepo) GetByID(id uint) (*model.Question, error) {
	var item model.Question
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *QuestionRepo) Create(item *model.Question) error {
	return r.db.Create(item).Error
}

func (r *QuestionRepo) Update(item *model.Question) error {
	// 仅更新业务字段，防止覆盖状态/审核/发布等受控字段
	return r.db.Model(&model.Question{}).
		Where("id = ?", item.ID).
		Select(
			"title", "question_type", "stem", "content", "answer", "analysis",
			"materials", "score_rule", "exam_id", "subject_id", "category_id",
			"source_id", "difficulty", "resource_type", "group_id", "class_ids",
			"user_ids", "analysis_visible_policy", "answer_visible_policy",
			"updated_by", "updated_at",
		).
		Updates(item).Error
}

func (r *QuestionRepo) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Question{}).Error
}

func (r *QuestionRepo) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总题数
	var total int64
	if err := r.db.Model(&model.Question{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 按状态统计
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusCounts []StatusCount
	if err := r.db.Model(&model.Question{}).Where("deleted_at IS NULL").
		Select("status, count(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		return nil, err
	}
	stats["byStatus"] = statusCounts

	// 按题型统计
	type TypeCount struct {
		QuestionType string `json:"questionType"`
		Count        int64  `json:"count"`
	}
	var typeCounts []TypeCount
	if err := r.db.Model(&model.Question{}).Where("deleted_at IS NULL").
		Select("question_type, count(*) as count").Group("question_type").Scan(&typeCounts).Error; err != nil {
		return nil, err
	}
	stats["byType"] = typeCounts

	// 按难度统计
	type DifficultyCount struct {
		Difficulty int   `json:"difficulty"`
		Count      int64 `json:"count"`
	}
	var diffCounts []DifficultyCount
	if err := r.db.Model(&model.Question{}).Where("deleted_at IS NULL").
		Select("difficulty, count(*) as count").Group("difficulty").Scan(&diffCounts).Error; err != nil {
		return nil, err
	}
	stats["byDifficulty"] = diffCounts

	return stats, nil
}

// applyVisibilityFilter 应用题目资源类型可见性过滤
// public: 所有人；private: 创建者；group: 分组内成员或创建者；class: 班级成员或创建者；user: 指定用户或创建者
func (r *QuestionRepo) applyVisibilityFilter(query *gorm.DB, userID, userGroupID uint, userClassIDs []uint) *gorm.DB {
	conditions := []string{
		"resource_type = 'public'",
		"resource_type = 'private' AND created_by = ?",
		"resource_type = 'group' AND (group_id = ? OR created_by = ?)",
	}
	args := []interface{}{userID, userGroupID, userID}

	if len(userClassIDs) > 0 {
		classCondition := "resource_type = 'class' AND ("
		for i, classID := range userClassIDs {
			if i > 0 {
				classCondition += " OR "
			}
			classCondition += "JSON_CONTAINS(class_ids, JSON_ARRAY(?))"
			args = append(args, classID)
		}
		classCondition += " OR created_by = ?)"
		args = append(args, userID)
		conditions = append(conditions, classCondition)
	} else {
		conditions = append(conditions, "resource_type = 'class' AND created_by = ?")
		args = append(args, userID)
	}

	conditions = append(conditions, "resource_type = 'user' AND (JSON_CONTAINS(user_ids, JSON_ARRAY(?)) OR created_by = ?)")
	args = append(args, userID, userID)

	where := "(" + strings.Join(conditions, ") OR (") + ")"
	return query.Where(where, args...)
}
