package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONMap JSON 字段类型
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONMap: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, j)
}

// ScheduledTask 定时任务配置
type ScheduledTask struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string     `gorm:"size:100;not null;comment:任务名称" json:"name"`
	TaskType   string     `gorm:"size:50;not null;comment:任务类型" json:"taskType"`
	CronExpr   string     `gorm:"size:50;not null;default:'0 3 * * *';comment:Cron表达式" json:"cronExpr"`
	Config     JSONMap    `gorm:"type:json;comment:任务配置" json:"config"`
	Enabled    bool       `gorm:"default:true;comment:是否启用" json:"enabled"`
	Status     string     `gorm:"size:20;default:idle;comment:状态" json:"status"`
	LastRunAt  *time.Time `gorm:"comment:最后执行时间" json:"lastRunAt,omitempty"`
	LastResult string     `gorm:"type:text;comment:最后执行结果" json:"lastResult,omitempty"`
	NextRunAt  *time.Time `gorm:"comment:下次执行时间" json:"nextRunAt,omitempty"`
	RunCount   int        `gorm:"default:0;comment:执行次数" json:"runCount"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func (ScheduledTask) TableName() string { return "sys_scheduled_tasks" }

// RecycleBinItem 回收站列表项（用于 API 返回）
type RecycleBinItem struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Size            int64      `json:"size"`
	ContentType     string     `json:"contentType"`
	FolderID        uint       `json:"folderId"`
	UserID          uint       `json:"userId"`
	UserName        string     `json:"userName,omitempty"`
	DeletedAt       *time.Time `json:"deletedAt"`
	RecycleExpireAt *time.Time `json:"recycleExpireAt"`
	DaysRemaining   int        `json:"daysRemaining"`
}
