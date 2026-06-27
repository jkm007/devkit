package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"

	"gorm.io/gorm"
)

const (
	MaxJSONSize       = 2 * 1024 * 1024 // 2MB 单字段
	MaxJSONDepth      = 10              // 最大嵌套深度
	MaxBlocksPerField = 50              // 单字段最大内容块数
)

// ValidateJSONField 校验JSON字段大小和深度
func ValidateJSONField(data string) error {
	if data == "" || data == "null" {
		return nil
	}
	if len(data) > MaxJSONSize {
		return fmt.Errorf("JSON内容过大，最大允许2MB")
	}
	// 检查JSON格式是否正确
	var js json.RawMessage
	if err := json.Unmarshal([]byte(data), &js); err != nil {
		return fmt.Errorf("JSON格式错误: %v", err)
	}
	return nil
}

type QuestionService struct {
	repo      *repository.QuestionRepo
	userRepo  *repository.UserRepo
	classRepo *repository.ClassRepo
}

func NewQuestionService() *QuestionService {
	return &QuestionService{
		repo:      repository.NewQuestionRepo(database.GetMySQL()),
		userRepo:  repository.NewUserRepo(database.GetMySQL()),
		classRepo: repository.NewClassRepo(database.GetMySQL()),
	}
}

type QuestionRequest struct {
	Title                 string `json:"title" binding:"required"`
	QuestionType          string `json:"questionType" binding:"required"`
	Stem                  string `json:"stem" binding:"required"`
	Content               string `json:"content"`
	Answer                string `json:"answer"`
	Analysis              string `json:"analysis"`
	Materials             string `json:"materials"`
	ScoreRule             string `json:"scoreRule"`
	ExamID                uint   `json:"examId"`
	SubjectID             uint   `json:"subjectId"`
	CategoryID            uint   `json:"categoryId"`
	SourceID              uint   `json:"sourceId"`
	Difficulty            *int   `json:"difficulty"`
	ResourceType          string `json:"resourceType"`
	GroupID               uint   `json:"groupId"`
	ClassIDs              []uint `json:"classIds"`
	UserIDs               []uint `json:"userIds"`
	AnalysisVisiblePolicy string `json:"analysisVisiblePolicy"`
	AnswerVisiblePolicy   string `json:"answerVisiblePolicy"`
}

type QuestionResponse struct {
	ID                    uint   `json:"id"`
	Title                 string `json:"title"`
	QuestionType          string `json:"questionType"`
	Stem                  string `json:"stem"`
	Content               string `json:"content"`
	Answer                string `json:"answer"`
	Analysis              string `json:"analysis"`
	Materials             string `json:"materials"`
	ScoreRule             string `json:"scoreRule"`
	ExamID                uint   `json:"examId"`
	SubjectID             uint   `json:"subjectId"`
	CategoryID            uint   `json:"categoryId"`
	SourceID              uint   `json:"sourceId"`
	Difficulty            int    `json:"difficulty"`
	ResourceType          string `json:"resourceType"`
	GroupID               uint   `json:"groupId"`
	ClassIDs              []uint `json:"classIds"`
	UserIDs               []uint `json:"userIds"`
	Status                string `json:"status"`
	CurrentVersionID      uint   `json:"currentVersionId"`
	ParentID              uint   `json:"parentId"`
	IsGroup               int    `json:"isGroup"`
	SubIndex              int    `json:"subIndex"`
	AnalysisVisiblePolicy string `json:"analysisVisiblePolicy"`
	AnswerVisiblePolicy   string `json:"answerVisiblePolicy"`
	CreatedBy             uint   `json:"createdBy"`
	ReviewedBy            uint   `json:"reviewedBy"`
	ReviewedAt            string `json:"reviewedAt"`
	RejectReason          string `json:"rejectReason"`
	PublishedAt           string `json:"publishedAt"`
	CreatedAt             string `json:"createTime"`
}

func (s *QuestionService) List(page, pageSize int, filters map[string]interface{}) ([]QuestionResponse, int64, error) {
	// 自动补全用户分组ID和班级ID列表，用于资源类型权限过滤
	if uid, ok := filters["userId"]; ok && uid != nil {
		var userID uint
		switch v := uid.(type) {
		case uint:
			userID = v
		case float64:
			userID = uint(v)
		case int:
			userID = uint(v)
		}
		if userID > 0 {
			if _, ok := filters["userGroupId"]; !ok {
				if user, err := s.userRepo.GetByID(userID); err == nil {
					filters["userGroupId"] = user.GroupID
				}
			}
			if _, ok := filters["userClassIds"]; !ok {
				if classIDs, err := s.classRepo.ListClassIDsByUserID(userID); err == nil {
					filters["userClassIds"] = classIDs
				}
			}
		}
	}

	items, total, err := s.repo.List(page, pageSize, filters)
	if err != nil {
		return nil, 0, err
	}
	var resp []QuestionResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, total, nil
}

func (s *QuestionService) Search(page, pageSize int, keyword string, userID uint) ([]QuestionResponse, int64, error) {
	userGroupID := uint(0)
	if user, err := s.userRepo.GetByID(userID); err == nil {
		userGroupID = user.GroupID
	}
	userClassIDs, _ := s.classRepo.ListClassIDsByUserID(userID)
	items, total, err := s.repo.Search(page, pageSize, keyword, userID, userGroupID, userClassIDs)
	if err != nil {
		return nil, 0, err
	}
	var resp []QuestionResponse
	for _, item := range items {
		resp = append(resp, s.toResponse(&item))
	}
	return resp, total, nil
}

func (s *QuestionService) GetByID(id uint, currentUserID uint, roles []string) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	// 权限校验：管理员/超管可查看全部
	isAdmin := false
	for _, role := range roles {
		if role == "admin" || role == "super_admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		// 非管理员按资源类型过滤
		allowed := false
		switch item.ResourceType {
		case "public":
			allowed = true
		case "private":
			allowed = item.CreatedBy == currentUserID
		case "group":
			userGroupID := uint(0)
			if user, uerr := s.userRepo.GetByID(currentUserID); uerr == nil {
				userGroupID = user.GroupID
			}
			allowed = item.GroupID > 0 && item.GroupID == userGroupID || item.CreatedBy == currentUserID
		case "class":
			allowed = item.CreatedBy == currentUserID
			if !allowed {
				classIDs, _ := s.classRepo.ListClassIDsByUserID(currentUserID)
				for _, cid := range item.ClassIDs {
					for _, userCID := range classIDs {
						if cid == userCID {
							allowed = true
							break
						}
					}
					if allowed {
						break
					}
				}
			}
		case "user":
			for _, uid := range item.UserIDs {
				if uid == currentUserID {
					allowed = true
					break
				}
			}
			if !allowed {
				allowed = item.CreatedBy == currentUserID
			}
		default:
			allowed = item.CreatedBy == currentUserID
		}
		if !allowed {
			return nil, fmt.Errorf("题目不存在")
		}
		// 非终态题目只有创建者或管理员可见
		if item.Status != "published" && item.Status != "approved" && item.CreatedBy != currentUserID {
			return nil, fmt.Errorf("题目不存在")
		}
	}

	resp := s.toResponse(item)
	return &resp, nil
}

func ensureJSON(s string) string {
	if s == "" || s == "null" {
		return "[]"
	}
	return s
}

func (s *QuestionService) Create(req *QuestionRequest, createdBy uint) (*QuestionResponse, error) {
	// 校验JSON字段大小
	jsonFields := map[string]string{
		"stem":      req.Stem,
		"content":   req.Content,
		"answer":    req.Answer,
		"analysis":  req.Analysis,
		"materials": req.Materials,
		"scoreRule": req.ScoreRule,
	}
	for name, value := range jsonFields {
		if err := ValidateJSONField(value); err != nil {
			return nil, fmt.Errorf("%s字段校验失败: %v", name, err)
		}
	}

	item := &model.Question{
		Title:                 req.Title,
		QuestionType:          req.QuestionType,
		Stem:                  ensureJSON(req.Stem),
		Content:               ensureJSON(req.Content),
		Answer:                ensureJSON(req.Answer),
		Analysis:              ensureJSON(req.Analysis),
		Materials:             ensureJSON(req.Materials),
		ScoreRule:             ensureJSON(req.ScoreRule),
		ExamID:                req.ExamID,
		SubjectID:             req.SubjectID,
		CategoryID:            req.CategoryID,
		SourceID:              req.SourceID,
		Difficulty:            1,
		ResourceType:          "private",
		Status:                "draft",
		AnalysisVisiblePolicy: "after_answer",
		AnswerVisiblePolicy:   "after_answer",
		CreatedBy:             createdBy,
		UpdatedBy:             createdBy,
	}
	if req.Difficulty != nil {
		item.Difficulty = *req.Difficulty
	}
	if req.ResourceType != "" {
		item.ResourceType = req.ResourceType
	}
	if req.GroupID > 0 {
		item.GroupID = req.GroupID
	}
	if len(req.ClassIDs) > 0 {
		item.ClassIDs = req.ClassIDs
	}
	if len(req.UserIDs) > 0 {
		item.UserIDs = req.UserIDs
	}
	if req.AnalysisVisiblePolicy != "" {
		item.AnalysisVisiblePolicy = req.AnalysisVisiblePolicy
	}
	if req.AnswerVisiblePolicy != "" {
		item.AnswerVisiblePolicy = req.AnswerVisiblePolicy
	}

	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionService) Update(id uint, req *QuestionRequest, updatedBy uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	// 校验JSON字段大小
	jsonFields := map[string]string{
		"stem":      req.Stem,
		"content":   req.Content,
		"answer":    req.Answer,
		"analysis":  req.Analysis,
		"materials": req.Materials,
		"scoreRule": req.ScoreRule,
	}
	for name, value := range jsonFields {
		if err := ValidateJSONField(value); err != nil {
			return nil, fmt.Errorf("%s字段校验失败: %v", name, err)
		}
	}

	item.Title = req.Title
	item.QuestionType = req.QuestionType
	item.Stem = ensureJSON(req.Stem)
	item.Content = ensureJSON(req.Content)
	item.Answer = ensureJSON(req.Answer)
	item.Analysis = ensureJSON(req.Analysis)
	item.Materials = ensureJSON(req.Materials)
	item.ScoreRule = ensureJSON(req.ScoreRule)
	item.ExamID = req.ExamID
	item.SubjectID = req.SubjectID
	item.CategoryID = req.CategoryID
	item.SourceID = req.SourceID
	item.UpdatedBy = updatedBy
	if req.Difficulty != nil {
		item.Difficulty = *req.Difficulty
	}
	if req.ResourceType != "" {
		item.ResourceType = req.ResourceType
	}
	if req.GroupID > 0 || req.ResourceType == "group" {
		item.GroupID = req.GroupID
	}
	if len(req.ClassIDs) > 0 || req.ResourceType == "class" {
		item.ClassIDs = req.ClassIDs
	}
	if len(req.UserIDs) > 0 || req.ResourceType == "user" {
		item.UserIDs = req.UserIDs
	}
	if req.AnalysisVisiblePolicy != "" {
		item.AnalysisVisiblePolicy = req.AnalysisVisiblePolicy
	}
	if req.AnswerVisiblePolicy != "" {
		item.AnswerVisiblePolicy = req.AnswerVisiblePolicy
	}

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionService) Delete(id uint, deletedBy uint) error {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	// 只有创建者才能删除
	if item.CreatedBy != deletedBy {
		return fmt.Errorf("只有题目创建者才能删除")
	}
	// 只有草稿、已驳回、已下架可以删除
	allowed := map[string]bool{"draft": true, "rejected": true, "archived": true}
	if !allowed[item.Status] {
		return fmt.Errorf("当前状态【%s】不允许删除", item.Status)
	}
	return s.repo.Delete(id)
}

// Publish 发布题目
// 私有题目：草稿可直接发布
// 其他类型：必须审核通过(approved)才能发布
func (s *QuestionService) Publish(id uint, publishedBy uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	if item.Status == "published" {
		return nil, fmt.Errorf("题目已发布")
	}

	// 校验必填字段
	if item.Stem == "" || item.Stem == "{}" || item.Stem == "null" {
		return nil, fmt.Errorf("题干不能为空")
	}

	// 私有题目：草稿可直接发布；其他类型：必须审核通过
	if item.ResourceType == "private" {
		if item.Status != "draft" {
			return nil, fmt.Errorf("私有题目只有草稿状态才能直接发布")
		}
	} else {
		if item.Status != "approved" {
			return nil, fmt.Errorf("当前状态【%s】不允许发布，请先提交审核并等待审核通过", item.Status)
		}
	}

	now := time.Now()
	err = s.repo.DB().Model(&model.Question{}).
		Where("id = ?", id).
		Select("status", "published_at", "updated_at").
		Updates(map[string]interface{}{
			"status":       "published",
			"published_at": now,
			"updated_at":   now,
		}).Error
	if err != nil {
		return nil, err
	}

	item, err = s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

// Archive 下架题目（已发布 → 已下架）
func (s *QuestionService) Archive(id uint, archivedBy uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}
	// 只有创建者才能下架
	if item.CreatedBy != archivedBy {
		return nil, fmt.Errorf("只有题目创建者才能下架")
	}
	if item.Status != "published" {
		return nil, fmt.Errorf("只有已发布的题目才能下架")
	}

	item.Status = "archived"
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

// SubmitAudit 提交审核（草稿/已驳回 → 待审核）
func (s *QuestionService) SubmitAudit(id uint, submittedBy uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	// 只有创建者才能提交审核
	if item.CreatedBy != submittedBy {
		return nil, fmt.Errorf("只有题目创建者才能提交审核")
	}

	if item.Status != "draft" && item.Status != "rejected" {
		return nil, fmt.Errorf("只有草稿或被驳回的题目才能提交审核")
	}

	item.Status = "pending"
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

// Approve 审核通过（待审核 → 审核通过）
func (s *QuestionService) Approve(id uint, reviewedBy uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	if item.Status != "pending" {
		return nil, fmt.Errorf("只有待审核的题目才能审核")
	}

	// 审核人不能审核自己的题目
	if item.CreatedBy == reviewedBy {
		return nil, fmt.Errorf("审核人不能审核自己创建的题目")
	}

	now := time.Now()
	item.Status = "approved"
	item.ReviewedBy = reviewedBy
	item.ReviewedAt = &now
	item.RejectReason = ""

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

// Reject 审核驳回（待审核 → 已驳回）
func (s *QuestionService) Reject(id uint, reviewedBy uint, reason string) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	if item.Status != "pending" {
		return nil, fmt.Errorf("只有待审核的题目才能驳回")
	}

	now := time.Now()
	item.Status = "rejected"
	item.ReviewedBy = reviewedBy
	item.ReviewedAt = &now
	item.RejectReason = reason

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

// Withdraw 撤回到草稿（待审核/审核通过/已发布 → 草稿）
func (s *QuestionService) Withdraw(id uint, withdrawnBy uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	// 只有创建者才能撤回
	if item.CreatedBy != withdrawnBy {
		return nil, fmt.Errorf("只有题目创建者才能撤回")
	}

	allowed := map[string]bool{"pending": true, "approved": true, "published": true}
	if !allowed[item.Status] {
		return nil, fmt.Errorf("当前状态【%s】不允许撤回", item.Status)
	}

	item.Status = "draft"
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

// Reactivate 重新上架（已下架 → 已发布）
func (s *QuestionService) Reactivate(id uint, reactivatedBy uint) (*QuestionResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("题目不存在")
		}
		return nil, err
	}

	// 只有创建者才能重新上架
	if item.CreatedBy != reactivatedBy {
		return nil, fmt.Errorf("只有题目创建者才能重新上架")
	}

	if item.Status != "archived" {
		return nil, fmt.Errorf("只有已下架的题目才能重新上架")
	}

	now := time.Now()
	err = s.repo.DB().Model(&model.Question{}).
		Where("id = ?", id).
		Select("status", "updated_at").
		Updates(map[string]interface{}{
			"status":     "published",
			"updated_at": now,
		}).Error
	if err != nil {
		return nil, err
	}

	item, err = s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	resp := s.toResponse(item)
	return &resp, nil
}

func (s *QuestionService) GetStats() (map[string]interface{}, error) {
	return s.repo.GetStats()
}

func (s *QuestionService) toResponse(item *model.Question) QuestionResponse {
	resp := QuestionResponse{
		ID:                    item.ID,
		Title:                 item.Title,
		QuestionType:          item.QuestionType,
		Stem:                  item.Stem,
		Content:               item.Content,
		Answer:                item.Answer,
		Analysis:              item.Analysis,
		Materials:             item.Materials,
		ScoreRule:             item.ScoreRule,
		ExamID:                item.ExamID,
		SubjectID:             item.SubjectID,
		CategoryID:            item.CategoryID,
		SourceID:              item.SourceID,
		Difficulty:            item.Difficulty,
		ResourceType:          item.ResourceType,
		GroupID:               item.GroupID,
		ClassIDs:              item.ClassIDs,
		UserIDs:               item.UserIDs,
		Status:                item.Status,
		CurrentVersionID:      item.CurrentVersionID,
		ParentID:              item.ParentID,
		IsGroup:               item.IsGroup,
		SubIndex:              item.SubIndex,
		AnalysisVisiblePolicy: item.AnalysisVisiblePolicy,
		AnswerVisiblePolicy:   item.AnswerVisiblePolicy,
		CreatedBy:             item.CreatedBy,
		ReviewedBy:            item.ReviewedBy,
		RejectReason:          item.RejectReason,
		CreatedAt:             item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
	}
	if item.ReviewedAt != nil {
		resp.ReviewedAt = item.ReviewedAt.Format("2006-01-02T15:04:05.000-07:00")
	}
	if item.PublishedAt != nil {
		resp.PublishedAt = item.PublishedAt.Format("2006-01-02T15:04:05.000-07:00")
	}
	return resp
}
