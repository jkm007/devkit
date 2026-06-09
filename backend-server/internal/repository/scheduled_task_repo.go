package repository

import (
	"backend-server/internal/model"
	"time"

	"gorm.io/gorm"
)

type ScheduledTaskRepo struct {
	db *gorm.DB
}

func NewScheduledTaskRepo(db *gorm.DB) *ScheduledTaskRepo {
	return &ScheduledTaskRepo{db: db}
}

// GetAll 获取所有定时任务
func (r *ScheduledTaskRepo) GetAll() ([]model.ScheduledTask, error) {
	var tasks []model.ScheduledTask
	err := r.db.Order("id ASC").Find(&tasks).Error
	return tasks, err
}

// GetByID 根据 ID 获取任务
func (r *ScheduledTaskRepo) GetByID(id uint) (*model.ScheduledTask, error) {
	var task model.ScheduledTask
	err := r.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByType 根据类型获取任务
func (r *ScheduledTaskRepo) GetByType(taskType string) (*model.ScheduledTask, error) {
	var task model.ScheduledTask
	err := r.db.Where("task_type = ?", taskType).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetEnabled 获取所有启用的任务
func (r *ScheduledTaskRepo) GetEnabled() ([]model.ScheduledTask, error) {
	var tasks []model.ScheduledTask
	err := r.db.Where("enabled = ?", true).Find(&tasks).Error
	return tasks, err
}

// GetDueTasks 获取到期需要执行的任务
func (r *ScheduledTaskRepo) GetDueTasks() ([]model.ScheduledTask, error) {
	var tasks []model.ScheduledTask
	now := time.Now()
	err := r.db.Where("enabled = ? AND (next_run_at IS NULL OR next_run_at <= ?)", true, now).Find(&tasks).Error
	return tasks, err
}

// Create 创建任务
func (r *ScheduledTaskRepo) Create(task *model.ScheduledTask) error {
	return r.db.Create(task).Error
}

// Update 更新任务
func (r *ScheduledTaskRepo) Update(task *model.ScheduledTask) error {
	return r.db.Save(task).Error
}

// UpdateStatus 更新任务状态
func (r *ScheduledTaskRepo) UpdateStatus(id uint, status string, result string) error {
	now := time.Now()
	return r.db.Model(&model.ScheduledTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      status,
		"last_run_at": now,
		"last_result": result,
		"run_count":   gorm.Expr("run_count + 1"),
	}).Error
}

// UpdateNextRun 更新下次执行时间
func (r *ScheduledTaskRepo) UpdateNextRun(id uint, nextRun time.Time) error {
	return r.db.Model(&model.ScheduledTask{}).Where("id = ?", id).Update("next_run_at", nextRun).Error
}

// UpdateEnabled 更新启用状态
func (r *ScheduledTaskRepo) UpdateEnabled(id uint, enabled bool) error {
	return r.db.Model(&model.ScheduledTask{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// Delete 删除任务
func (r *ScheduledTaskRepo) Delete(id uint) error {
	return r.db.Delete(&model.ScheduledTask{}, id).Error
}
