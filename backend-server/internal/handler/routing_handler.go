package handler

import (
	"strconv"

	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/response"
	"backend-server/pkg/storage"

	"github.com/gin-gonic/gin"
)

// RoutingHandler 路由规则处理器
type RoutingHandler struct {
	routingService *service.RoutingService
}

// NewRoutingHandler 创建路由规则处理器
func NewRoutingHandler(routingService *service.RoutingService) *RoutingHandler {
	return &RoutingHandler{routingService: routingService}
}

// GetAllRules 获取所有路由规则
// @Summary      获取路由规则列表
// @Description  获取所有标签路由规则
// @Tags         标签路由
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]model.TagRouting} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/routing-rules [get]
func (h *RoutingHandler) GetAllRules(c *gin.Context) {
	rules, err := h.routingService.GetAllRules()
	if err != nil {
		response.InternalError(c, "获取规则失败")
		return
	}
	response.Success(c, rules)
}

// GetRuleByID 根据ID获取规则
// @Summary      获取路由规则详情
// @Description  根据 ID 获取标签路由规则详情
// @Tags         标签路由
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "规则 ID"
// @Success      200  {object}  response.Response{data=model.TagRouting} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/routing-rules/{id} [get]
func (h *RoutingHandler) GetRuleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	rule, err := h.routingService.GetRuleByID(id)
	if err != nil {
		response.NotFound(c, "规则不存在")
		return
	}
	response.Success(c, rule)
}

// CreateRule 创建路由规则
// @Summary      创建路由规则
// @Description  创建新的标签路由规则
// @Tags         标签路由
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  model.TagRouting  true  "路由规则"
// @Success      200  {object}  response.Response{data=model.TagRouting} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/routing-rules [post]
func (h *RoutingHandler) CreateRule(c *gin.Context) {
	var rule model.TagRouting
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if rule.RuleName == "" {
		response.BadRequest(c, "规则名称不能为空")
		return
	}

	if rule.Driver == "" {
		response.BadRequest(c, "目标存储驱动不能为空")
		return
	}

	if err := h.routingService.CreateRule(&rule); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 刷新路由引擎
	storage.RefreshRoutingEngine()

	response.SuccessWithMessage(c, "创建成功", rule)
}

// UpdateRule 更新路由规则
// @Summary      更新路由规则
// @Description  更新指定标签路由规则
// @Tags         标签路由
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "规则 ID"
// @Param        data  body  model.TagRouting  true  "路由规则"
// @Success      200  {object}  response.Response{data=model.TagRouting} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/routing-rules/{id} [put]
func (h *RoutingHandler) UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	// 先获取现有规则，保留 created_at
	existing, err := h.routingService.GetRuleByID(id)
	if err != nil {
		response.NotFound(c, "规则不存在")
		return
	}

	var rule model.TagRouting
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 保留原始的创建时间
	rule.ID = id
	rule.CreatedAt = existing.CreatedAt
	if err := h.routingService.UpdateRule(&rule); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", rule)
}

// DeleteRule 删除路由规则
// @Summary      删除路由规则
// @Description  删除指定标签路由规则
// @Tags         标签路由
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "规则 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/routing-rules/{id} [delete]
func (h *RoutingHandler) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	if err := h.routingService.DeleteRule(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// UpdateStatus 更新规则状态
// @Summary      更新路由规则状态
// @Description  更新指定标签路由规则状态
// @Tags         标签路由
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "规则 ID"
// @Param        data  body  object  true  "状态"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/routing-rules/{id}/status [put]
func (h *RoutingHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.routingService.UpdateStatus(id, req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// UpdatePriority 更新规则优先级
// @Summary      更新路由规则优先级
// @Description  更新指定标签路由规则优先级
// @Tags         标签路由
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "规则 ID"
// @Param        data  body  object  true  "优先级"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/routing-rules/{id}/priority [put]
func (h *RoutingHandler) UpdatePriority(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	var req struct {
		Priority int `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.routingService.UpdatePriority(id, req.Priority); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// BatchUpdatePriority 批量更新优先级
// @Summary      批量更新路由规则优先级
// @Description  批量更新标签路由规则优先级
// @Tags         标签路由
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  object  true  "优先级映射"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/routing-rules/batch-priority [post]
func (h *RoutingHandler) BatchUpdatePriority(c *gin.Context) {
	var req struct {
		Priorities map[int64]int `json:"priorities"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.routingService.BatchUpdatePriority(req.Priorities); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// TestRule 测试规则匹配
// @Summary      测试路由规则
// @Description  测试指定标签路由规则是否匹配
// @Tags         标签路由
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "规则 ID"
// @Param        data  body  object  true  "标签列表"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/routing-rules/{id}/test [post]
func (h *RoutingHandler) TestRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	var req struct {
		Tags []storage.RoutingTag `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	matched, err := h.routingService.TestRule(id, req.Tags)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"matched": matched})
}

// TestRoute 测试文件路由
// @Summary      测试文件路由
// @Description  根据文件信息测试路由结果
// @Tags         标签路由
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  object  true  "文件信息"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/routing-rules/test-route [post]
func (h *RoutingHandler) TestRoute(c *gin.Context) {
	var req struct {
		FileName    string `json:"fileName"`
		ContentType string `json:"contentType"`
		Source      string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, tags, err := h.routingService.TestRoute(req.FileName, req.ContentType, req.Source)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"result": result,
		"tags":   tags,
	})
}
