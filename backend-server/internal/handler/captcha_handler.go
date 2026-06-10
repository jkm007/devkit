package handler

import (
	"context"
	"fmt"
	"time"

	"backend-server/pkg/captcha"
	"backend-server/pkg/database"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// CaptchaHandler 验证码处理器
type CaptchaHandler struct{}

// NewCaptchaHandler 创建验证码处理器
func NewCaptchaHandler() *CaptchaHandler {
	return &CaptchaHandler{}
}

// GetCaptcha 获取验证码
// @Summary      获取图形验证码
// @Description  根据 type 参数生成对应类型的验证码
// @Tags         认证
// @Produce      json
// @Param        type  query  string  false  "验证码类型: numeric/slider/puzzle/rotation/point"  default(numeric)
// @Success      200  {object}  response.Response{data=captcha.CaptchaData} "成功"
// @Router       /auth/captcha [get]
func (h *CaptchaHandler) GetCaptcha(c *gin.Context) {
	// 基于 IP 的速率限制：每分钟最多 10 次请求
	rdb := database.GetRedis()
	ctx := context.Background()
	rateLimitKey := fmt.Sprintf("captcha_rate:%s", c.ClientIP())

	count, err := rdb.Incr(ctx, rateLimitKey).Result()
	if err == nil && count == 1 {
		rdb.Expire(ctx, rateLimitKey, time.Minute)
	}
	if count > 10 {
		response.TooManyRequests(c, "验证码请求过于频繁，请稍后再试")
		return
	}

	captchaType := c.DefaultQuery("type", "numeric")

	data, err := captcha.Generate(captchaType)
	if err != nil {
		response.InternalError(c, "生成验证码失败")
		return
	}

	response.Success(c, data)
}

// TestCaptcha 生成测试验证码（管理员用）
// @Summary      生成测试验证码
// @Description  生成一个用于测试的验证码
// @Tags         系统设置
// @Produce      json
// @Param        type  query  string  false  "验证码类型: numeric/slider/puzzle/rotation/point"  default(numeric)
// @Success      200  {object}  response.Response{data=captcha.CaptchaData} "成功"
// @Router       /system/captcha/test [get]
func (h *CaptchaHandler) TestCaptcha(c *gin.Context) {
	captchaType := c.DefaultQuery("type", "numeric")

	data, err := captcha.Generate(captchaType)
	if err != nil {
		response.InternalError(c, "生成验证码失败")
		return
	}

	response.Success(c, data)
}

// VerifyCaptcha 验证验证码（管理员测试用）
// @Summary      验证验证码
// @Description  验证用户提交的验证码是否正确
// @Tags         系统设置
// @Accept       json
// @Produce      json
// @Param        body  body  handler.VerifyCaptchaRequest  true  "验证请求"
// @Success      200  {object}  response.Response{data=handler.VerifyCaptchaResponse} "成功"
// @Router       /system/captcha/verify [post]
func (h *CaptchaHandler) VerifyCaptcha(c *gin.Context) {
	var req VerifyCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if req.CaptchaID == "" {
		response.BadRequest(c, "验证码ID不能为空")
		return
	}

	valid, msg := captcha.Verify(req.CaptchaID, req.CaptchaCode, req.StartTime, req.Points)

	response.Success(c, VerifyCaptchaResponse{
		Valid:   valid,
		Message: msg,
	})
}

// VerifyCaptchaRequest 验证码验证请求
type VerifyCaptchaRequest struct {
	CaptchaID   string          `json:"captchaId"`
	CaptchaCode string          `json:"captchaCode"`
	StartTime   int64           `json:"startTime"`   // 前端记录的开始时间戳（毫秒）
	Points      []captcha.Point `json:"points"`      // 点选验证码的点击坐标序列
}

// VerifyCaptchaResponse 验证码验证响应
type VerifyCaptchaResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}
