package handler

import (
	"strconv"

	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

type QuestionImportHandler struct {
	service *service.QuestionImportService
}

func NewQuestionImportHandler() *QuestionImportHandler {
	return &QuestionImportHandler{
		service: service.NewQuestionImportService(),
	}
}

func (h *QuestionImportHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	filters := map[string]interface{}{
		"status":    c.Query("status"),
		"createdBy": c.Query("createdBy"),
	}
	items, total, err := h.service.List(page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, items, total)
}

func (h *QuestionImportHandler) GetDetail(c *gin.Context) {
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

func (h *QuestionImportHandler) GetItems(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	items, err := h.service.GetItems(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *QuestionImportHandler) Create(c *gin.Context) {
	var req service.ImportTaskRequest
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

func (h *QuestionImportHandler) Delete(c *gin.Context) {
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
