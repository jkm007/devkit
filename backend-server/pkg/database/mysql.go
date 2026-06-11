package database

import (
	"fmt"
	"time"

	"backend-server/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// InitMySQL 初始化 MySQL 连接
func InitMySQL(cfg config.MySQLConfig) error {
	dsn := cfg.DSN()

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("连接 MySQL 失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取 SQL DB 失败: %w", err)
	}

	// 连接池配置
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	// ConnMaxLifetime 已是 Go duration 格式，无需转换
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("Ping MySQL 失败: %w", err)
	}

	return nil
}

// GetMySQL 获取 GORM 实例
func GetMySQL() *gorm.DB {
	return db
}

// CloseMySQL 关闭 MySQL 连接
func CloseMySQL() {
	if db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}
