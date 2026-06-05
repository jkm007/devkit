package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
	Message string      `json:"message"`
}

// PageData 分页数据
type PageData struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

const (
	CodeSuccess = 0
	CodeFail    = -1
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Data:    data,
		Error:   nil,
		Message: "ok",
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Data:    data,
		Error:   nil,
		Message: msg,
	})
}

// SuccessPage 分页成功响应
func SuccessPage(c *gin.Context, items interface{}, total int64) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Data: PageData{
			Items: items,
			Total: total,
		},
		Error:   nil,
		Message: "ok",
	})
}

// Fail 错误响应
func Fail(c *gin.Context, httpStatus int, msg string) {
	c.JSON(httpStatus, Response{
		Code:    CodeFail,
		Data:    nil,
		Error:   msg,
		Message: msg,
	})
}

// BadRequest 参数错误 (400)
func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, msg)
}

// Unauthorized 未授权 (401)
func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, msg)
}

// Forbidden 禁止访问 (403)
func Forbidden(c *gin.Context, msg string) {
	Fail(c, http.StatusForbidden, msg)
}

// NotFound 资源不存在 (404)
func NotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, msg)
}

// InternalError 内部错误 (500)
func InternalError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, msg)
}

// TooManyRequests 请求过于频繁 (429)
func TooManyRequests(c *gin.Context, msg string) {
	Fail(c, http.StatusTooManyRequests, msg)
}
