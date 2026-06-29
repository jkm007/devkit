package handler

import (
	"strconv"

	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

type QuestionHandler struct {
	service *service.QuestionService
}

func NewQuestionHandler() *QuestionHandler {
	return &QuestionHandler{
		service: service.NewQuestionService(),
	}
}

func (h *QuestionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	userId, _ := c.Get("user_id")
	roles, _ := c.Get("roles")
	filters := map[string]interface{}{
		"userId":       userId,
		"roles":        roles,
		"title":        c.Query("title"),
		"questionType": c.Query("questionType"),
		"status":       c.Query("status"),
		"examId":       c.Query("examId"),
		"subjectId":    c.Query("subjectId"),
		"categoryId":   c.Query("categoryId"),
		"difficulty":   c.Query("difficulty"),
		"resourceType": c.Query("resourceType"),
		"createdBy":    c.Query("createdBy"),
	}
	items, total, err := h.service.List(page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, items, total)
}

func (h *QuestionHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userId, _ := c.Get("user_id")
	roles, _ := c.Get("roles")
	roleList, _ := roles.([]string)
	item, err := h.service.GetByID(uint(id), userId.(uint), roleList)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionHandler) Create(c *gin.Context) {
	var req service.QuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userId, _ := c.Get("user_id")
	item, err := h.service.Create(&req, userId.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req service.QuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userId, _ := c.Get("user_id")
	roles, _ := c.Get("roles")
	roleList, _ := roles.([]string)
	item, err := h.service.Update(uint(id), &req, userId.(uint), roleList)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userId, _ := c.Get("user_id")
	if err := h.service.Delete(uint(id), userId.(uint)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *QuestionHandler) Publish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userId, _ := c.Get("user_id")
	item, err := h.service.Publish(uint(id), userId.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionHandler) Archive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userId, _ := c.Get("user_id")
	item, err := h.service.Archive(uint(id), userId.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionHandler) SubmitAudit(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userId, _ := c.Get("user_id")
	item, err := h.service.SubmitAudit(uint(id), userId.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userId, _ := c.Get("user_id")
	item, err := h.service.Approve(uint(id), userId.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userId, _ := c.Get("user_id")
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	item, err := h.service.Reject(uint(id), userId.(uint), req.Reason)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionHandler) Withdraw(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userId, _ := c.Get("user_id")
	item, err := h.service.Withdraw(uint(id), userId.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionHandler) Reactivate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	userId, _ := c.Get("user_id")
	item, err := h.service.Reactivate(uint(id), userId.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	status := c.Query("status") // 可选，默认 published
	userId, _ := c.Get("user_id")

	if keyword == "" {
		response.BadRequest(c, "搜索关键词不能为空")
		return
	}

	items, total, err := h.service.Search(page, pageSize, keyword, userId.(uint), status)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, items, total)
}

func (h *QuestionHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

func (h *QuestionHandler) GetTypes(c *gin.Context) {
	types := []map[string]interface{}{
		{"value": "single_choice", "label": "单选题", "category": "基础客观题"},
		{"value": "multiple_choice", "label": "多选题", "category": "基础客观题"},
		{"value": "indefinite_choice", "label": "不定项选择题", "category": "基础客观题"},
		{"value": "true_false", "label": "判断题", "category": "基础客观题"},
		{"value": "fill_blank", "label": "填空题", "category": "填写类题"},
		{"value": "cloze", "label": "完形填空", "category": "填写类题"},
		{"value": "term_explanation", "label": "名词解释", "category": "填写类题"},
		{"value": "short_answer", "label": "简答题", "category": "主观题"},
		{"value": "essay_question", "label": "论述题", "category": "主观题"},
		{"value": "composition", "label": "作文题", "category": "主观题"},
		{"value": "material", "label": "材料题", "category": "组合题"},
		{"value": "case_analysis", "label": "案例分析题", "category": "组合题"},
		{"value": "reading", "label": "阅读理解题", "category": "组合题"},
		{"value": "matching", "label": "匹配题", "category": "组合题"},
		{"value": "ordering", "label": "排序题", "category": "排序归类题"},
		{"value": "classification", "label": "分类题", "category": "排序归类题"},
		{"value": "listening", "label": "听力题", "category": "音视频题"},
		{"value": "speaking", "label": "口语题", "category": "音视频题"},
		{"value": "video", "label": "视频题", "category": "音视频题"},
		{"value": "document", "label": "文档题", "category": "文档材料题"},
		{"value": "calculation", "label": "计算题", "category": "计算实操题"},
		{"value": "proof", "label": "证明题", "category": "计算实操题"},
		{"value": "operation", "label": "操作题", "category": "计算实操题"},
		{"value": "programming", "label": "编程题", "category": "技术类题"},
		{"value": "sql", "label": "SQL题", "category": "技术类题"},
		{"value": "code_reading", "label": "代码阅读题", "category": "技术类题"},
		{"value": "debugging", "label": "调试改错题", "category": "技术类题"},
	}
	response.Success(c, types)
}
