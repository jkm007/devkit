package handler

import (
	"strconv"

	"backend-server/internal/service"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// RiskScoreHandler 风险评分处理器
type RiskScoreHandler struct {
	riskScoreService *service.RiskScoreService
}

// NewRiskScoreHandler 创建风险评分处理器
func NewRiskScoreHandler() *RiskScoreHandler {
	return &RiskScoreHandler{
		riskScoreService: service.NewRiskScoreService(),
	}
}

// GetRiskScores 获取所有风险评分
// @Summary      获取风险评分列表
// @Description  获取当前所有 IP 的风险评分，按分数降序排列
// @Tags         风险评分
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query  int  false  "返回数量限制，默认100"
// @Success      200  {object}  response.Response{data=[]service.RiskScoreItem} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /system/risk/scores [get]
func (h *RiskScoreHandler) GetRiskScores(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000 // 最大限制
	}

	scores, err := h.riskScoreService.GetRiskScores(limit)
	if err != nil {
		response.InternalError(c, "获取风险评分失败: "+err.Error())
		return
	}

	response.Success(c, scores)
}

// GetRiskScoreByIP 获取指定 IP 的风险评分
// @Summary      获取指定 IP 风险评分
// @Description  获取指定 IP 的当前风险评分
// @Tags         风险评分
// @Produce      json
// @Security     BearerAuth
// @Param        ip  query  string  true  "IP 地址"
// @Success      200  {object}  response.Response{data=service.RiskScoreItem} "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /system/risk/score [get]
func (h *RiskScoreHandler) GetRiskScoreByIP(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		response.BadRequest(c, "IP 地址不能为空")
		return
	}

	score, err := h.riskScoreService.GetRiskScoreByIP(ip)
	if err != nil {
		response.InternalError(c, "获取风险评分失败: "+err.Error())
		return
	}

	response.Success(c, score)
}

// ClearRiskScore 清除指定 IP 的风险评分
// @Summary      清除风险评分
// @Description  清除指定 IP 的所有风险评分数据（包括频率计数、间隔时间等）
// @Tags         风险评分
// @Produce      json
// @Security     BearerAuth
// @Param        ip  query  string  true  "IP 地址"
// @Success      200  {object}  response.Response "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /system/risk/clear [post]
func (h *RiskScoreHandler) ClearRiskScore(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		response.BadRequest(c, "IP 地址不能为空")
		return
	}

	err := h.riskScoreService.ClearRiskScore(ip)
	if err != nil {
		response.InternalError(c, "清除风险评分失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "已清除 " + ip + " 的风险评分"})
}

// GetRiskScoreStats 获取风险评分统计
// @Summary      获取风险评分统计
// @Description  获取风险评分系统的统计数据
// @Tags         风险评分
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response "成功"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /system/risk/stats [get]
func (h *RiskScoreHandler) GetRiskScoreStats(c *gin.Context) {
	stats, err := h.riskScoreService.GetRiskScoreStats()
	if err != nil {
		response.InternalError(c, "获取统计失败: "+err.Error())
		return
	}

	response.Success(c, stats)
}
