package sms

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"backend-server/pkg/database"
)

// Sender 短信发送接口
type Sender interface {
	Send(phone, code string) error
}

// Config 短信配置
type Config struct {
	Enabled bool
	Driver  string // aliyun / tencent

	// 阿里云
	AliyunAccessKeyID     string
	AliyunAccessKeySecret string
	AliyunSignName        string
	AliyunTemplateCode    string

	// 腾讯云
	TencentSecretID   string
	TencentSecretKey  string
	TencentAppID      string
	TencentSignName   string
	TencentTemplateID string
}

// 配置缓存
var configCache *Config
var configMutex sync.RWMutex

// GetConfig 获取短信配置（带缓存）
func GetConfig() *Config {
	configMutex.RLock()
	if configCache != nil {
		configMutex.RUnlock()
		return configCache
	}
	configMutex.RUnlock()

	configMutex.Lock()
	defer configMutex.Unlock()

	if configCache != nil {
		return configCache
	}

	configCache = loadConfigFromDB()
	return configCache
}

// RefreshConfig 刷新短信配置缓存
func RefreshConfig() {
	configMutex.Lock()
	configCache = loadConfigFromDB()
	configMutex.Unlock()
}

// loadConfigFromDB 从数据库加载短信配置
func loadConfigFromDB() *Config {
	db := database.GetMySQL()

	rows, err := db.Raw("SELECT `key`, value FROM sys_system_settings WHERE group_key = 'sms' AND deleted_at IS NULL").Rows()
	if err != nil {
		return &Config{}
	}
	defer rows.Close()

	kv := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		value = strings.Trim(value, "\"")
		kv[key] = value
	}

	return &Config{
		Enabled:               kv["sms_enabled"] == "true",
		Driver:                kv["sms_driver"],
		AliyunAccessKeyID:     kv["aliyun_access_key_id"],
		AliyunAccessKeySecret: kv["aliyun_access_key_secret"],
		AliyunSignName:        kv["aliyun_sign_name"],
		AliyunTemplateCode:    kv["aliyun_template_code"],
		TencentSecretID:       kv["tencent_secret_id"],
		TencentSecretKey:      kv["tencent_secret_key"],
		TencentAppID:          kv["tencent_app_id"],
		TencentSignName:       kv["tencent_sign_name"],
		TencentTemplateID:     kv["tencent_template_id"],
	}
}

// GetSender 根据配置返回对应的短信发送器
func GetSender() (Sender, error) {
	cfg := GetConfig()

	if !cfg.Enabled {
		return nil, errors.New("短信服务未启用")
	}

	switch cfg.Driver {
	case "aliyun":
		if cfg.AliyunAccessKeyID == "" || cfg.AliyunAccessKeySecret == "" {
			return nil, errors.New("阿里云短信配置不完整")
		}
		return NewAliyunSender(cfg), nil
	case "tencent":
		if cfg.TencentSecretID == "" || cfg.TencentSecretKey == "" {
			return nil, errors.New("腾讯云短信配置不完整")
		}
		return NewTencentSender(cfg), nil
	default:
		return nil, fmt.Errorf("不支持的短信驱动: %s", cfg.Driver)
	}
}

// formatCode 将验证码格式化为模板参数（如 "123456" → "\"123456\""）
func formatCode(code string) string {
	return strconv.Quote(code)
}
