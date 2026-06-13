package handler

import (
	"strconv"

	"backend-server/internal/middleware"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// StudyHandler 学习相关接口
type StudyHandler struct {
	studyService     *service.StudyService
	favoriteService  *service.FavoriteNoteService
}

// NewStudyHandler 创建学习接口处理器
func NewStudyHandler() *StudyHandler {
	return &StudyHandler{
		studyService:    service.NewStudyService(),
		favoriteService: service.NewFavoriteNoteService(),
	}
}

// ListQuestions 获取题目列表
func (h *StudyHandler) ListQuestions(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	filters := make(map[string]interface{})
	if questionType := c.Query("questionType"); questionType != "" {
		filters["questionType"] = questionType
	}
	if categoryIDStr := c.Query("categoryId"); categoryIDStr != "" {
		if categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
			filters["categoryId"] = uint(categoryID)
		}
	}
	if difficultyStr := c.Query("difficulty"); difficultyStr != "" {
		if difficulty, err := strconv.Atoi(difficultyStr); err == nil && difficulty > 0 {
			filters["difficulty"] = difficulty
		}
	}
	if keyword := c.Query("keyword"); keyword != "" {
		filters["keyword"] = keyword
	}
	if knowledgePoint := c.Query("knowledgePoint"); knowledgePoint != "" {
		filters["knowledgePoint"] = knowledgePoint
	}

	items, total, err := h.studyService.ListQuestions(userID, page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessPage(c, items, total)
}

// GetQuestion 获取题目详情
func (h *StudyHandler) GetQuestion(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的题目ID")
		return
	}

	question, err := h.studyService.GetQuestion(userID, uint(id))
	if err != nil {
		response.NotFound(c, "题目不存在")
		return
	}

	response.Success(c, question)
}

// AddFavorite 收藏题目
func (h *StudyHandler) AddFavorite(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的题目ID")
		return
	}

	if err := h.favoriteService.AddFavorite(userID, uint(id)); err != nil {
		response.InternalError(c, "收藏失败")
		return
	}

	response.SuccessWithMessage(c, "收藏成功", nil)
}

// RemoveFavorite 取消收藏
func (h *StudyHandler) RemoveFavorite(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的题目ID")
		return
	}

	if err := h.favoriteService.RemoveFavorite(userID, uint(id)); err != nil {
		response.InternalError(c, "取消收藏失败")
		return
	}

	response.SuccessWithMessage(c, "已取消收藏", nil)
}

// ListFavorites 获取收藏列表
func (h *StudyHandler) ListFavorites(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	items, total, err := h.favoriteService.ListFavorites(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, "获取收藏列表失败")
		return
	}

	response.SuccessPage(c, items, total)
}

// CreateNote 创建/更新笔记
func (h *StudyHandler) CreateNote(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	var req service.NoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	note, err := h.favoriteService.CreateNote(userID, &req)
	if err != nil {
		response.InternalError(c, "保存笔记失败")
		return
	}

	response.Success(c, note)
}

// ListNotes 获取笔记列表
func (h *StudyHandler) ListNotes(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	items, total, err := h.favoriteService.ListNotes(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, "获取笔记列表失败")
		return
	}

	response.SuccessPage(c, items, total)
}

// UpdateNote 更新笔记
func (h *StudyHandler) UpdateNote(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	noteIDStr := c.Param("id")
	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的笔记ID")
		return
	}

	var req service.NoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	note, err := h.favoriteService.UpdateNote(uint(noteID), userID, &req)
	if err != nil {
		response.InternalError(c, "更新笔记失败")
		return
	}

	response.Success(c, note)
}

// DeleteNote 删除笔记
func (h *StudyHandler) DeleteNote(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	noteIDStr := c.Param("id")
	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的笔记ID")
		return
	}

	if err := h.favoriteService.DeleteNote(uint(noteID), userID); err != nil {
		response.InternalError(c, "删除笔记失败")
		return
	}

	response.SuccessWithMessage(c, "笔记已删除", nil)
}

// GetPracticeQuestions 获取练习题目（随机）
func (h *StudyHandler) GetPracticeQuestions(c *gin.Context) {
	var req service.PracticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Count < 1 || req.Count > 100 {
		req.Count = 20
	}

	questions, err := h.studyService.GetRandomQuestions(&req)
	if err != nil {
		response.InternalError(c, "获取练习题目失败")
		return
	}

	response.Success(c, questions)
}

// SubmitPractice 提交练习结果
func (h *StudyHandler) SubmitPractice(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	var req service.PracticeSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.studyService.SubmitPractice(userID, &req); err != nil {
		response.InternalError(c, "提交练习结果失败")
		return
	}

	response.SuccessWithMessage(c, "练习结果已保存", nil)
}

// GetPracticeHistory 获取练习历史
func (h *StudyHandler) GetPracticeHistory(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	items, total, err := h.studyService.GetPracticeHistory(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, "获取练习历史失败")
		return
	}

	response.SuccessPage(c, items, total)
}
