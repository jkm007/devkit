package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

const (
	// maxConcurrentTasks 定时任务最大并发执行数
	maxConcurrentTasks = 3
)

// ScheduledTaskService 定时任务服务
type ScheduledTaskService struct {
	taskRepo *repository.ScheduledTaskRepo
}

func NewScheduledTaskService() *ScheduledTaskService {
	db := database.GetMySQL()
	return &ScheduledTaskService{
		taskRepo: repository.NewScheduledTaskRepo(db),
	}
}

// Create 创建任务
func (s *ScheduledTaskService) Create(name string, taskType string, cronExpr string, config model.JSONMap) (*model.ScheduledTask, error) {
	// 验证 cron 表达式
	if !isValidCronExpr(cronExpr) {
		return nil, fmt.Errorf("无效的 Cron 表达式: %s", cronExpr)
	}

	task := &model.ScheduledTask{
		Name:     name,
		TaskType: taskType,
		CronExpr: cronExpr,
		Config:   config,
		Enabled:  true,
		Status:   "idle",
	}

	// 计算下次执行时间
	task.NextRunAt = s.calculateNextRun(cronExpr)

	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

// GetAll 获取所有任务
func (s *ScheduledTaskService) GetAll() ([]model.ScheduledTask, error) {
	return s.taskRepo.GetAll()
}

// GetByID 获取任务详情
func (s *ScheduledTaskService) GetByID(id uint) (*model.ScheduledTask, error) {
	return s.taskRepo.GetByID(id)
}

// Update 更新任务配置
func (s *ScheduledTaskService) Update(id uint, name string, cronExpr string, config model.JSONMap, enabled bool) error {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	// 验证 cron 表达式
	if !isValidCronExpr(cronExpr) {
		return fmt.Errorf("无效的 Cron 表达式: %s", cronExpr)
	}

	task.Name = name
	task.CronExpr = cronExpr
	if config != nil {
		task.Config = config
	}
	task.Enabled = enabled

	// 计算下次执行时间
	task.NextRunAt = s.calculateNextRun(cronExpr)

	return s.taskRepo.Update(task)
}

// UpdateEnabled 更新启用状态
func (s *ScheduledTaskService) UpdateEnabled(id uint, enabled bool) error {
	return s.taskRepo.UpdateEnabled(id, enabled)
}

// Delete 删除任务
func (s *ScheduledTaskService) Delete(id uint) error {
	return s.taskRepo.Delete(id)
}

// RunTask 手动执行任务
func (s *ScheduledTaskService) RunTask(id uint) error {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	return s.executeTask(task)
}

// CheckAndRunDueTasks 检查并执行到期任务（定时调用）
// 使用 worker pool 限制并发执行数量，避免同时运行过多任务
func (s *ScheduledTaskService) CheckAndRunDueTasks() {
	tasks, err := s.taskRepo.GetDueTasks()
	if err != nil {
		log.Printf("[ERROR] 获取到期任务失败: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	// 使用带缓冲的 channel 作为信号量，限制并发数
	sem := make(chan struct{}, maxConcurrentTasks)
	var wg sync.WaitGroup

	for _, task := range tasks {
		task := task // 显式捕获循环变量，避免闭包引用问题
		wg.Add(1)
		sem <- struct{}{} // 获取信号量，超过并发上限时阻塞
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量
			s.executeTask(&task)
		}()
	}

	// 等待所有任务执行完毕，避免下一轮检查与本轮任务并发冲突
	wg.Wait()
}

// executeTask 执行单个任务
func (s *ScheduledTaskService) executeTask(task *model.ScheduledTask) error {
	log.Printf("[INFO] 执行定时任务: %s (type=%s)", task.Name, task.TaskType)

	// 更新状态为运行中
	s.taskRepo.UpdateStatus(task.ID, "running", "")

	var result string
	var err error

	switch task.TaskType {
	case "recycle_cleanup":
		result, err = s.executeRecycleCleanup(task)
	default:
		result = fmt.Sprintf("未知任务类型: %s", task.TaskType)
		err = fmt.Errorf("unknown task type: %s", task.TaskType)
	}

	if err != nil {
		s.taskRepo.UpdateStatus(task.ID, "failed", fmt.Sprintf("执行失败: %v", err))
		log.Printf("[ERROR] 任务执行失败: %s, error=%v", task.Name, err)
		return err
	}

	// 更新状态为成功
	s.taskRepo.UpdateStatus(task.ID, "success", result)

	// 计算下次执行时间
	nextRun := s.calculateNextRun(task.CronExpr)
	if nextRun != nil {
		s.taskRepo.UpdateNextRun(task.ID, *nextRun)
	}

	log.Printf("[INFO] 任务执行成功: %s, result=%s", task.Name, result)
	return nil
}

// executeRecycleCleanup 执行回收站清理
func (s *ScheduledTaskService) executeRecycleCleanup(task *model.ScheduledTask) (string, error) {
	recycleService := NewRecycleBinService()

	// 从任务配置中读取 retention_days
	retentionDays := 7 // 默认保留 7 天
	if task.Config != nil {
		if val, ok := task.Config["retention_days"]; ok {
			switch v := val.(type) {
			case float64:
				retentionDays = int(v)
			case int:
				retentionDays = v
			case string:
				if d, err := strconv.Atoi(v); err == nil && d > 0 {
					retentionDays = d
				}
			}
		}
	}

	deleted, err := recycleService.CleanupExpiredFilesWithRetention(retentionDays)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("已清理 %d 个过期文件（保留天数: %d）", deleted, retentionDays), nil
}

// calculateNextRun 计算下次执行时间（支持 * 和具体数字，找到下一个大于当前时间的执行时间）
func (s *ScheduledTaskService) calculateNextRun(cronExpr string) *time.Time {
	parts := strings.Fields(cronExpr)
	if len(parts) != 5 {
		return nil
	}

	minute, err1 := parseCronField(parts[0], 0, 59)
	hour, err2 := parseCronField(parts[1], 0, 23)

	if err1 != nil || err2 != nil {
		return nil
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 在今天和明天中查找下一个匹配的执行时间（最多搜索2天）
	for dayOffset := 0; dayOffset <= 1; dayOffset++ {
		base := today.Add(time.Duration(dayOffset) * 24 * time.Hour)
		for _, h := range hour {
			for _, m := range minute {
				candidate := base.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute)
				if candidate.After(now) {
					return &candidate
				}
			}
		}
	}

	return nil
}

// parseCronField 解析 cron 字段（简化版，支持 * 和具体数字）
func parseCronField(field string, min, max int) ([]int, error) {
	if field == "*" {
		result := make([]int, max-min+1)
		for i := min; i <= max; i++ {
			result[i-min] = i
		}
		return result, nil
	}

	val, err := strconv.Atoi(field)
	if err != nil || val < min || val > max {
		return nil, fmt.Errorf("invalid cron field: %s", field)
	}

	return []int{val}, nil
}

// isValidCronExpr 验证 cron 表达式格式
func isValidCronExpr(expr string) bool {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return false
	}

	// 简单验证每个字段
	for i, part := range parts {
		if part == "*" {
			continue
		}
		val, err := strconv.Atoi(part)
		if err != nil {
			return false
		}
		switch i {
		case 0: // minute
			if val < 0 || val > 59 {
				return false
			}
		case 1: // hour
			if val < 0 || val > 23 {
				return false
			}
		case 2: // day of month
			if val < 1 || val > 31 {
				return false
			}
		case 3: // month
			if val < 1 || val > 12 {
				return false
			}
		case 4: // day of week
			if val < 0 || val > 6 {
				return false
			}
		}
	}

	return true
}

// InitScheduledTasks 初始化定时任务（启动时调用）
func InitScheduledTasks() {
	taskService := NewScheduledTaskService()

	// 计算所有启用任务的下次执行时间
	tasks, err := taskService.GetAll()
	if err != nil {
		log.Printf("[WARN] 初始化定时任务失败: %v", err)
		return
	}

	for _, task := range tasks {
		if task.Enabled && task.NextRunAt == nil {
			nextRun := taskService.calculateNextRun(task.CronExpr)
			if nextRun != nil {
				taskService.taskRepo.UpdateNextRun(task.ID, *nextRun)
			}
		}
	}

	// 启动定时任务检查器（每分钟检查一次）
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			taskService.CheckAndRunDueTasks()
		}
	}()

	log.Printf("[INFO] 定时任务服务已启动")
}
