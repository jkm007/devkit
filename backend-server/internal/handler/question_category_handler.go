package handler

import (
	"strconv"

	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// ==================== 考试大类 ====================

type ExamCategoryHandler struct {
	service *service.ExamCategoryService
}

func NewExamCategoryHandler() *ExamCategoryHandler {
	return &ExamCategoryHandler{
		service: service.NewExamCategoryService(),
	}
}

func (h *ExamCategoryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	filters := map[string]interface{}{
		"name":   c.Query("name"),
		"status": c.Query("status"),
	}
	if c.Query("parentId") != "" {
		filters["parentId"] = c.Query("parentId")
	}
	items, total, err := h.service.List(page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, items, total)
}

func (h *ExamCategoryHandler) GetAll(c *gin.Context) {
	items, err := h.service.GetAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *ExamCategoryHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	item, err := h.service.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *ExamCategoryHandler) Create(c *gin.Context) {
	var req service.ExamCategoryRequest
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

func (h *ExamCategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req service.ExamCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.service.Update(uint(id), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *ExamCategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, nil)
}

// ==================== 具体考试 ====================

type ExamHandler struct {
	service *service.ExamService
}

func NewExamHandler() *ExamHandler {
	return &ExamHandler{
		service: service.NewExamService(),
	}
}

func (h *ExamHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	filters := map[string]interface{}{
		"name":           c.Query("name"),
		"status":         c.Query("status"),
		"examCategoryId": c.Query("examCategoryId"),
	}
	items, total, err := h.service.List(page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, items, total)
}

func (h *ExamHandler) GetAll(c *gin.Context) {
	categoryId := c.Query("examCategoryId")
	if categoryId != "" {
		cid, err := strconv.ParseUint(categoryId, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的考试大类ID")
			return
		}
		items, err := h.service.GetByCategoryID(uint(cid))
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		response.Success(c, items)
		return
	}
	items, err := h.service.GetAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *ExamHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	item, err := h.service.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *ExamHandler) Create(c *gin.Context) {
	var req service.ExamRequest
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

func (h *ExamHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req service.ExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.service.Update(uint(id), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *ExamHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, nil)
}

// ==================== 科目 ====================

type SubjectHandler struct {
	service *service.SubjectService
}

func NewSubjectHandler() *SubjectHandler {
	return &SubjectHandler{
		service: service.NewSubjectService(),
	}
}

func (h *SubjectHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	filters := map[string]interface{}{
		"name":   c.Query("name"),
		"status": c.Query("status"),
		"examId": c.Query("examId"),
	}
	items, total, err := h.service.List(page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, items, total)
}

func (h *SubjectHandler) GetAll(c *gin.Context) {
	examId := c.Query("examId")
	if examId != "" {
		eid, err := strconv.ParseUint(examId, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的考试ID")
			return
		}
		items, err := h.service.GetByExamID(uint(eid))
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		response.Success(c, items)
		return
	}
	response.BadRequest(c, "请提供考试ID")
}

func (h *SubjectHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	item, err := h.service.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *SubjectHandler) Create(c *gin.Context) {
	var req service.SubjectRequest
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

func (h *SubjectHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req service.SubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.service.Update(uint(id), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *SubjectHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, nil)
}

// ==================== 章节分类 ====================

type QuestionCategoryHandler struct {
	service *service.QuestionCategoryService
}

func NewQuestionCategoryHandler() *QuestionCategoryHandler {
	return &QuestionCategoryHandler{
		service: service.NewQuestionCategoryService(),
	}
}

func (h *QuestionCategoryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	filters := map[string]interface{}{
		"name":      c.Query("name"),
		"status":    c.Query("status"),
		"examId":    c.Query("examId"),
		"subjectId": c.Query("subjectId"),
		"parentId":  c.Query("parentId"),
	}
	items, total, err := h.service.List(page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, items, total)
}

func (h *QuestionCategoryHandler) GetAll(c *gin.Context) {
	items, err := h.service.GetAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *QuestionCategoryHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	item, err := h.service.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionCategoryHandler) Create(c *gin.Context) {
	var req service.QuestionCategoryRequest
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

func (h *QuestionCategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req service.QuestionCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.service.Update(uint(id), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionCategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, nil)
}
