package handler

import (
	"backend-server/internal/service"
	"backend-server/pkg/captcha"
	"backend-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// VerifyCodeHandler 验证码处理器
type VerifyCodeHandler struct {
	verifyCodeService *service.VerifyCodeService
}

// NewVerifyCodeHandler 创建验证码处理器
func NewVerifyCodeHandler() *VerifyCodeHandler {
	return &VerifyCodeHandler{
		verifyCodeService: service.NewVerifyCodeService(),
	}
}

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Purpose     string `json:"purpose" binding:"required,oneof=register reset_password login"`
	CaptchaID   string `json:"captchaId" binding:"required"`
	CaptchaCode string `json:"captchaCode" binding:"required,min=1,max=256"`
}

// VerifyCodeRequest 验证验证码请求
type VerifyCodeRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Code    string `json:"code" binding:"required,len=6"`
	Purpose string `json:"purpose" binding:"required,oneof=register reset_password login"`
}

// SendCode 发送邮箱验证码
// @Summary      发送邮箱验证码
// @Description  向指定邮箱发送验证码（注册/重置密码）
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  handler.SendCodeRequest  true  "发送请求"
// @Success      200   {object}  response.Response
// @Router       /auth/send-code [post]
func (h *VerifyCodeHandler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 先校验图形验证码，防止接口被滥用
	ok, msg := captcha.Verify(req.CaptchaID, req.CaptchaCode, 0, nil)
	if !ok {
		response.BadRequest(c, msg)
		return
	}

	err := h.verifyCodeService.SendCode(req.Email, req.Purpose)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "验证码已发送到您的邮箱", nil)
}

// VerifyCode 验证邮箱验证码
// @Summary      验证邮箱验证码
// @Description  验证用户提交的邮箱验证码是否正确
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  handler.VerifyCodeRequest  true  "验证请求"
// @Success      200   {object}  response.Response
// @Router       /auth/verify-code [post]
func (h *VerifyCodeHandler) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	valid, err := h.verifyCodeService.VerifyCode(req.Email, req.Code, req.Purpose)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if !valid {
		response.BadRequest(c, "验证码错误")
		return
	}

	response.SuccessWithMessage(c, "验证通过", nil)
}

// SendSMSCodeRequest 发送短信验证码请求
type SendSMSCodeRequest struct {
	Phone       string `json:"phone" binding:"required,len=11"`
	Purpose     string `json:"purpose" binding:"required,oneof=login"`
	CaptchaID   string `json:"captchaId" binding:"required"`
	CaptchaCode string `json:"captchaCode" binding:"required,min=1,max=256"`
}

// SendSMSCode 发送短信验证码
// @Summary      发送短信验证码
// @Description  向指定手机号发送验证码
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  handler.SendSMSCodeRequest  true  "发送请求"
// @Success      200   {object}  response.Response
// @Router       /auth/send-sms-code [post]
func (h *VerifyCodeHandler) SendSMSCode(c *gin.Context) {
	var req SendSMSCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 先校验图形验证码，防止接口被滥用
	ok, msg := captcha.Verify(req.CaptchaID, req.CaptchaCode, 0, nil)
	if !ok {
		response.BadRequest(c, msg)
		return
	}

	err := h.verifyCodeService.SendSMSCode(req.Phone, req.Purpose)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "验证码已发送到您的手机", nil)
}
