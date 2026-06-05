package validator

import (
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// 注册自定义校验规则
	validate.RegisterValidation("mobile", validateMobile)
	validate.RegisterValidation("password", validatePassword)
}

// GetValidator 获取校验器实例
func GetValidator() *validator.Validate {
	return validate
}

// ValidateStruct 校验结构体
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// validateMobile 手机号校验（中国大陆）
func validateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	re := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return re.MatchString(mobile)
}

// validatePassword 密码强度校验（8-20位，包含大小写字母和数字）
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 || len(password) > 20 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasUpper && hasLower && hasDigit
}

// BindAndValidate 绑定并校验请求参数
func BindAndValidate(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}
	return validate.Struct(obj)
}
