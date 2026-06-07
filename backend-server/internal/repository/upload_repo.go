package repository

import (
	"time"

	"backend-server/internal/model"

	"gorm.io/gorm"
)

// UploadRepo 上传任务仓库
type UploadRepo struct {
	db *gorm.DB
}

func NewUploadRepo(db *gorm.DB) *UploadRepo {
	return &UploadRepo{db: db}
}

// CreateTask 创建上传任务
func (r *UploadRepo) CreateTask(task *model.UploadTask) error {
	return r.db.Create(task).Error
}

// GetTaskByUploadID 根据 uploadID 获取任务
func (r *UploadRepo) GetTaskByUploadID(uploadID string) (*model.UploadTask, error) {
	var task model.UploadTask
	if err := r.db.Where("upload_id = ?", uploadID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTaskStatus 更新任务状态
func (r *UploadRepo) UpdateTaskStatus(uploadID string, status string) error {
	return r.db.Model(&model.UploadTask{}).Where("upload_id = ?", uploadID).Update("status", status).Error
}

// UpdateTaskProgress 更新任务进度
func (r *UploadRepo) UpdateTaskProgress(uploadID string, uploadedParts int, progress int) error {
	return r.db.Model(&model.UploadTask{}).Where("upload_id = ?", uploadID).
		Updates(map[string]interface{}{
			"uploaded_parts": uploadedParts,
			"progress":       progress,
		}).Error
}

// UpdateTaskFailed 更新任务失败状态
func (r *UploadRepo) UpdateTaskFailed(uploadID string, errorMessage string) error {
	return r.db.Model(&model.UploadTask{}).Where("upload_id = ?", uploadID).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": errorMessage,
		}).Error
}

// UpdateTaskCompleted 更新任务完成状态
func (r *UploadRepo) UpdateTaskCompleted(uploadID string) error {
	now := time.Now()
	return r.db.Model(&model.UploadTask{}).Where("upload_id = ?", uploadID).
		Updates(map[string]interface{}{
			"status":       "completed",
			"progress":     100,
			"completed_at": now,
		}).Error
}

// GetUserTasks 获取用户的上传任务列表
func (r *UploadRepo) GetUserTasks(userID uint, limit int) ([]model.UploadTask, error) {
	var tasks []model.UploadTask
	query := r.db.Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTaskByID 根据ID获取任务
func (r *UploadRepo) GetTaskByID(id uint) (*model.UploadTask, error) {
	var task model.UploadTask
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// CreatePart 记录已上传分片
func (r *UploadRepo) CreatePart(part *model.UploadedPart) error {
	return r.db.Create(part).Error
}

// GetParts 获取任务的所有已上传分片
func (r *UploadRepo) GetParts(uploadTaskID uint) ([]model.UploadedPart, error) {
	var parts []model.UploadedPart
	if err := r.db.Where("upload_task_id = ?", uploadTaskID).Order("part_number ASC").Find(&parts).Error; err != nil {
		return nil, err
	}
	return parts, nil
}

// DeletePartsByTaskID 删除任务的所有分片记录
func (r *UploadRepo) DeletePartsByTaskID(uploadTaskID uint) error {
	return r.db.Where("upload_task_id = ?", uploadTaskID).Delete(&model.UploadedPart{}).Error
}

// GetStaleTasks 获取超时未完成的上传任务
func (r *UploadRepo) GetStaleTasks(beforeUnix int64) ([]model.UploadTask, error) {
	var tasks []model.UploadTask
	if err := r.db.Where("status = ? AND created_at < ?", "uploading", beforeUnix).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}
