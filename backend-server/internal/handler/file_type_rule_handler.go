package handler

import (
	"strconv"

	"backend-server/internal/model"
	"backend-server/internal/service"
	"backend-server/pkg/response"
	"backend-server/pkg/storage"

	"github.com/gin-gonic/gin"
)

// FileTypeRuleHandler 文件类型规则处理器
type FileTypeRuleHandler struct {
	service *service.FileTypeRuleService
}

// NewFileTypeRuleHandler 创建文件类型规则处理器
func NewFileTypeRuleHandler() *FileTypeRuleHandler {
	return &FileTypeRuleHandler{
		service: service.NewFileTypeRuleService(),
	}
}

// GetAll 获取所有文件类型规则
// @Summary      获取文件类型规则列表
// @Description  获取所有文件类型规则
// @Tags         文件类型规则
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]model.FileTypeRule} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/file-type-rules [get]
func (h *FileTypeRuleHandler) GetAll(c *gin.Context) {
	rules, err := h.service.GetAll()
	if err != nil {
		response.InternalError(c, "获取规则列表失败")
		return
	}
	response.Success(c, rules)
}

// GetGrouped 获取按类型分组的规则
// @Summary      获取分组文件类型规则
// @Description  获取按文件类型分组的规则
// @Tags         文件类型规则
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/file-type-rules/grouped [get]
func (h *FileTypeRuleHandler) GetGrouped(c *gin.Context) {
	groups, err := h.service.GetGroupedByType()
	if err != nil {
		response.InternalError(c, "获取规则列表失败")
		return
	}
	response.Success(c, groups)
}

// Create 创建文件类型规则
// @Summary      创建文件类型规则
// @Description  创建新的文件类型规则
// @Tags         文件类型规则
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        data  body  model.FileTypeRule  true  "文件类型规则"
// @Success      200  {object}  response.Response{data=model.FileTypeRule} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/file-type-rules [post]
func (h *FileTypeRuleHandler) Create(c *gin.Context) {
	var rule model.FileTypeRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if rule.Extension == "" {
		response.BadRequest(c, "扩展名不能为空")
		return
	}
	if rule.FileType == "" {
		response.BadRequest(c, "文件类型不能为空")
		return
	}

	if err := h.service.Create(&rule); err != nil {
		if isBusinessError(err) {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.SuccessWithMessage(c, "创建成功", rule)
}

// Update 更新文件类型规则
// @Summary      更新文件类型规则
// @Description  更新指定文件类型规则
// @Tags         文件类型规则
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "规则 ID"
// @Param        data  body  model.FileTypeRule  true  "文件类型规则"
// @Success      200  {object}  response.Response{data=model.FileTypeRule} "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/file-type-rules/{id} [put]
func (h *FileTypeRuleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	var rule model.FileTypeRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if rule.Extension == "" {
		response.BadRequest(c, "扩展名不能为空")
		return
	}
	if rule.FileType == "" {
		response.BadRequest(c, "文件类型不能为空")
		return
	}

	rule.ID = id
	if err := h.service.Update(&rule); err != nil {
		if isBusinessError(err) {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.SuccessWithMessage(c, "更新成功", rule)
}

// Delete 删除文件类型规则
// @Summary      删除文件类型规则
// @Description  删除指定文件类型规则
// @Tags         文件类型规则
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "规则 ID"
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/file-type-rules/{id} [delete]
func (h *FileTypeRuleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if isBusinessError(err) {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// RefreshAutoTagger 刷新 AutoTagger 的文件类型规则
// @Summary      刷新自动标签器规则
// @Description  从数据库刷新 AutoTagger 的文件类型规则
// @Tags         文件类型规则
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response "成功"
// @Failure      400  {object}  response.Response "参数错误"
// @Failure      401  {object}  response.Response "未授权"
// @Failure      500  {object}  response.Response "服务器错误"
// @Router       /system/file-type-rules/refresh [post]
func (h *FileTypeRuleHandler) RefreshAutoTagger(c *gin.Context) {
	rules, err := h.service.GetAllEnabled()
	if err != nil {
		response.InternalError(c, "获取规则失败")
		return
	}

	// 转换为 storage 包的数据结构（避免循环导入）
	data := make([]storage.FileTypeRuleData, len(rules))
	for i, r := range rules {
		data[i] = storage.FileTypeRuleData{
			Extension: r.Extension,
			FileType:  r.FileType,
		}
	}

	storage.GetGlobalAutoTagger().LoadFromDB(data)
	response.SuccessWithMessage(c, "刷新成功", gin.H{"count": len(rules)})
}
