package handler

import (
	"strconv"

	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

type QuestionSourceHandler struct {
	service *service.QuestionSourceService
}

func NewQuestionSourceHandler() *QuestionSourceHandler {
	return &QuestionSourceHandler{
		service: service.NewQuestionSourceService(),
	}
}

func (h *QuestionSourceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	filters := map[string]interface{}{
		"name":       c.Query("name"),
		"sourceType": c.Query("sourceType"),
		"examId":     c.Query("examId"),
		"year":       c.Query("year"),
	}
	items, total, err := h.service.List(page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, items, total)
}

func (h *QuestionSourceHandler) GetDetail(c *gin.Context) {
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

func (h *QuestionSourceHandler) Create(c *gin.Context) {
	var req service.QuestionSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userId, _ := c.Get("userId")
	item, err := h.service.Create(&req, userId.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *QuestionSourceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req service.QuestionSourceRequest
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

func (h *QuestionSourceHandler) Delete(c *gin.Context) {
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
