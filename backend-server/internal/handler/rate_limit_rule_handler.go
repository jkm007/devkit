package handler

import (
	"strconv"

	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

type RateLimitRuleHandler struct {
	service *service.RateLimitRuleService
}

func NewRateLimitRuleHandler() *RateLimitRuleHandler {
	return &RateLimitRuleHandler{
		service: service.NewRateLimitRuleService(),
	}
}

type rateLimitRuleRequest struct {
	PathPattern    string  `json:"pathPattern" binding:"required"`
	Method         string  `json:"method"`
	Rate           float64 `json:"rate" binding:"required,gt=0"`
	Burst          int     `json:"burst" binding:"required,gt=0"`
	Cooldown       int     `json:"cooldown" binding:"min=0,max=86400"`
	BlockDuration  int     `json:"blockDuration" binding:"min=0,max=604800"`
	MaxViolations  int     `json:"maxViolations" binding:"min=0,max=10000"`
	ViolationScore int     `json:"violationScore" binding:"min=0,max=1000"`
	Description    string  `json:"description" binding:"max=500"`
	Enabled        *bool   `json:"enabled"`
	Priority       int     `json:"priority" binding:"min=0,max=9999"`
}

// List 获取所有限流规则
func (h *RateLimitRuleHandler) List(c *gin.Context) {
	rules, err := h.service.GetAll()
	if err != nil {
		response.InternalError(c, "获取限流规则失败")
		return
	}
	response.Success(c, rules)
}

// GetByID 获取单个规则
func (h *RateLimitRuleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的规则 ID")
		return
	}

	rule, err := h.service.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "规则不存在")
		return
	}
	response.Success(c, rule)
}

// Create 创建规则
func (h *RateLimitRuleHandler) Create(c *gin.Context) {
	var req rateLimitRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	method := req.Method
	if method == "" {
		method = "*"
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	rule := &model.RateLimitRule{
		PathPattern:    req.PathPattern,
		Method:         method,
		Rate:           req.Rate,
		Burst:          req.Burst,
		Cooldown:       req.Cooldown,
		BlockDuration:  req.BlockDuration,
		MaxViolations:  req.MaxViolations,
		ViolationScore: req.ViolationScore,
		Description:    req.Description,
		Enabled:        enabled,
		Priority:       req.Priority,
	}

	if err := h.service.Create(rule); err != nil {
		response.InternalError(c, "创建规则失败")
		return
	}
	response.Success(c, rule)
}

// Update 更新规则
func (h *RateLimitRuleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的规则 ID")
		return
	}

	rule, err := h.service.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "规则不存在")
		return
	}

	var req rateLimitRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	rule.PathPattern = req.PathPattern
	if req.Method != "" {
		rule.Method = req.Method
	}
	rule.Rate = req.Rate
	rule.Burst = req.Burst
	rule.Cooldown = req.Cooldown
	rule.BlockDuration = req.BlockDuration
	rule.MaxViolations = req.MaxViolations
	rule.ViolationScore = req.ViolationScore
	rule.Description = req.Description
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	rule.Priority = req.Priority

	if err := h.service.Update(rule); err != nil {
		response.InternalError(c, "更新规则失败")
		return
	}
	response.Success(c, rule)
}

// Delete 删除规则
func (h *RateLimitRuleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的规则 ID")
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除规则失败")
		return
	}
	response.Success(c, nil)
}

// UpdateStatus 更新启用状态
func (h *RateLimitRuleHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的规则 ID")
		return
	}

	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Enabled == nil {
		response.BadRequest(c, "enabled 字段必填")
		return
	}

	if err := h.service.UpdateEnabled(uint(id), *req.Enabled); err != nil {
		response.InternalError(c, "更新状态失败")
		return
	}
	response.Success(c, nil)
}
