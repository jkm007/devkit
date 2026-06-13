package handler

import (
	"strconv"

	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

type KnowledgePointHandler struct {
	service *service.KnowledgePointService
}

func NewKnowledgePointHandler() *KnowledgePointHandler {
	return &KnowledgePointHandler{
		service: service.NewKnowledgePointService(),
	}
}

func (h *KnowledgePointHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	filters := map[string]interface{}{
		"name":       c.Query("name"),
		"status":     c.Query("status"),
		"examId":     c.Query("examId"),
		"subjectId":  c.Query("subjectId"),
		"categoryId": c.Query("categoryId"),
		"parentId":   c.Query("parentId"),
	}
	items, total, err := h.service.List(page, pageSize, filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, items, total)
}

func (h *KnowledgePointHandler) GetAll(c *gin.Context) {
	items, err := h.service.GetAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *KnowledgePointHandler) GetDetail(c *gin.Context) {
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

func (h *KnowledgePointHandler) Create(c *gin.Context) {
	var req service.KnowledgePointRequest
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

func (h *KnowledgePointHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	var req service.KnowledgePointRequest
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

func (h *KnowledgePointHandler) Delete(c *gin.Context) {
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
