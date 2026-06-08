package service

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/captcha"
	"backend-server/pkg/database"
	"backend-server/pkg/storage"
)

// 有效分组列表
var validGroups = map[string]bool{
	"basic":      true,
	"auth":       true,
	"email":      true,
	"sms":        true,
	"captcha":    true,
	"storage":    true,
	"wechat":     true,
	"security":   true,
	"risk_score": true,
}

// SystemSettingService 系统配置服务
type SystemSettingService struct {
	repo *repository.SystemSettingRepo
}

// NewSystemSettingService 创建系统配置服务
func NewSystemSettingService() *SystemSettingService {
	return &SystemSettingService{
		repo: repository.NewSystemSettingRepo(database.GetMySQL()),
	}
}

// SettingItem 配置项响应
type SettingItem struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Label       string      `json:"label"`
	Type        string      `json:"type"`
	Options     interface{} `json:"options"`
	Tip         string      `json:"tip"`
	Sort        int         `json:"sort"`
	IsPublic    bool        `json:"isPublic"`
	IsSensitive bool        `json:"isSensitive"`
}

// UpdateSettingsRequest 批量更新请求
type UpdateSettingsRequest struct {
	Settings map[string]map[string]interface{} `json:"settings" binding:"required"`
}

// UpdateResult 更新结果
type UpdateResult struct {
	Updated         int      `json:"updated"`
	RestartRequired bool     `json:"restartRequired"`
	RestartItems    []string `json:"restartItems"`
}

// TestEmailRequest 测试邮件请求
type TestEmailRequest struct {
	To string `json:"to" binding:"required,email"`
}

// TestSMSRequest 测试短信请求
type TestSMSRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// GetPublic 获取公开配置
func (s *SystemSettingService) GetPublic() (map[string]map[string]interface{}, error) {
	settings, err := s.repo.GetPublic()
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]interface{})
	for _, setting := range settings {
		if _, ok := result[setting.GroupKey]; !ok {
			result[setting.GroupKey] = make(map[string]interface{})
		}
		result[setting.GroupKey][setting.Key] = parseValue(setting.Value, setting.Type)
	}
	return result, nil
}

// GetAll 获取所有配置（按分组，敏感字段脱敏）
func (s *SystemSettingService) GetAll() (map[string][]SettingItem, error) {
	settings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]SettingItem)
	for _, setting := range settings {
		item := toSettingItem(setting)
		result[setting.GroupKey] = append(result[setting.GroupKey], item)
	}
	return result, nil
}

// GetByGroup 获取指定分组配置
func (s *SystemSettingService) GetByGroup(groupKey string) ([]SettingItem, error) {
	if !validGroups[groupKey] {
		return nil, fmt.Errorf("invalid settings group: %s", groupKey)
	}

	settings, err := s.repo.GetByGroup(groupKey)
	if err != nil {
		return nil, err
	}

	var result []SettingItem
	for _, setting := range settings {
		result = append(result, toSettingItem(setting))
	}
	return result, nil
}

// Update 批量更新配置
func (s *SystemSettingService) Update(req *UpdateSettingsRequest, userID uint) (*UpdateResult, error) {
	var updates []model.SystemSetting

	for groupKey, kv := range req.Settings {
		if !validGroups[groupKey] {
			return nil, fmt.Errorf("invalid settings group: %s", groupKey)
		}

		// 获取该分组的有效 key
		validKeys, err := s.repo.GetKeysByGroup(groupKey)
		if err != nil {
			return nil, err
		}

		for key, val := range kv {
			if !validKeys[key] {
				return nil, fmt.Errorf("invalid settings key in group %s: %s", groupKey, key)
			}

			// 敏感字段：值为 "******" 则跳过
			strVal := fmt.Sprintf("%v", val)
			if strVal == "******" {
				continue
			}

			// 获取当前配置检查是否敏感
			setting, err := s.repo.GetByKey(groupKey, key)
			if err != nil {
				return nil, err
			}

			// 将值转为 JSON 格式存储
			jsonVal, err := toJSONValue(val, setting.Type)
			if err != nil {
				return nil, fmt.Errorf("invalid value for %s.%s: %v", groupKey, key, err)
			}

			updates = append(updates, model.SystemSetting{
				GroupKey: groupKey,
				Key:      key,
				Value:    jsonVal,
			})
		}
	}

	if len(updates) == 0 {
		return &UpdateResult{Updated: 0}, nil
	}

	updated, err := s.repo.BatchUpdate(updates, userID)
	if err != nil {
		return nil, err
	}

	// 如果更新了 captcha 或 risk_score 配置，刷新缓存
	for groupKey := range req.Settings {
		if groupKey == "captcha" {
			captcha.RefreshConfig()
		}
		if groupKey == "risk_score" {
			RefreshRiskConfig()
		}
		if groupKey == "storage" {
			if err := storage.RefreshStorage(); err != nil {
				return nil, fmt.Errorf("刷新存储配置失败: %w", err)
			}
			// 同步默认桶到存储桶管理
			if err := SyncDefaultBuckets(); err != nil {
				return nil, fmt.Errorf("同步默认存储桶失败: %w", err)
			}
		}
	}

	return &UpdateResult{
		Updated:         updated,
		RestartRequired: false,
		RestartItems:    []string{},
	}, nil
}

// SettingGroupUpdateRequest 按分组更新请求（settings 是 key-value 对）
type SettingGroupUpdateRequest struct {
	Settings map[string]interface{} `json:"settings" binding:"required"`
}

// UpdateByGroup 更新指定分组配置
func (s *SystemSettingService) UpdateByGroup(groupKey string, req *SettingGroupUpdateRequest, userID uint) (*UpdateResult, error) {
	if !validGroups[groupKey] {
		return nil, fmt.Errorf("invalid settings group: %s", groupKey)
	}

	// 获取该分组的有效 key
	validKeys, err := s.repo.GetKeysByGroup(groupKey)
	if err != nil {
		return nil, err
	}

	var updates []model.SystemSetting
	for key, val := range req.Settings {
		if !validKeys[key] {
			return nil, fmt.Errorf("invalid settings key in group %s: %s", groupKey, key)
		}

		// 敏感字段：值为 "******" 则跳过
		strVal := fmt.Sprintf("%v", val)
		if strVal == "******" {
			continue
		}

		// 获取当前配置检查类型
		setting, err := s.repo.GetByKey(groupKey, key)
		if err != nil {
			return nil, err
		}

		jsonVal, err := toJSONValue(val, setting.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s.%s: %v", groupKey, key, err)
		}

		updates = append(updates, model.SystemSetting{
			GroupKey: groupKey,
			Key:      key,
			Value:    jsonVal,
		})
	}

	if len(updates) == 0 {
		return &UpdateResult{Updated: 0}, nil
	}

	updated, err := s.repo.BatchUpdate(updates, userID)
	if err != nil {
		return nil, err
	}

	// 如果更新了 captcha 或 risk_score 配置，刷新缓存
	if groupKey == "captcha" {
		captcha.RefreshConfig()
	}
	if groupKey == "risk_score" {
		RefreshRiskConfig()
	}
	if groupKey == "storage" {
		if err := storage.RefreshStorage(); err != nil {
			return nil, fmt.Errorf("刷新存储配置失败: %w", err)
		}
		// 同步默认桶到存储桶管理
		if err := SyncDefaultBuckets(); err != nil {
			return nil, fmt.Errorf("同步默认存储桶失败: %w", err)
		}
	}

	return &UpdateResult{
		Updated:         updated,
		RestartRequired: false,
		RestartItems:    []string{},
	}, nil
}

// TestEmail 测试邮件发送
func (s *SystemSettingService) TestEmail(to string) error {
	// 从数据库获取邮件配置
	getVal := func(groupKey, key string) string {
		setting, err := s.repo.GetByKey(groupKey, key)
		if err != nil || setting.Value == "" {
			return ""
		}
		// 去除 JSON 引号
		val := setting.Value
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		return val
	}

	enabled := getVal("email", "smtp_enabled")
	if enabled != "true" {
		return errors.New("email service is not enabled")
	}

	host := getVal("email", "smtp_host")
	port := getVal("email", "smtp_port")
	username := getVal("email", "smtp_username")
	password := getVal("email", "smtp_password")
	from := getVal("email", "smtp_from")
	fromName := getVal("email", "smtp_from_name")

	if host == "" || port == "" || username == "" || password == "" {
		return errors.New("SMTP configuration is incomplete, please check host/port/username/password")
	}

	// 构建邮件
	subject := "Test Email from Backend Server"
	body := "This is a test email sent from the backend management system."
	mime := "MIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n"
	msg := "From: " + fromName + " <" + from + ">\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		mime + "\r\n" +
		body

	// 发送
	addr := host + ":" + port

	if port == "465" {
		// SSL 隐式加密（端口 465）
		return sendMailViaSSL(addr, host, username, password, from, to, msg)
	}

	// STARTTLS（端口 587 等）
	auth := smtp.PlainAuth("", username, password, host)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}

// sendMailViaSSL 通过 SSL/TLS 连接发送邮件（端口 465）
func sendMailViaSSL(addr, host, username, password, from, to, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server via SSL: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", username, password, host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %v", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %v", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO failed: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %v", err)
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email body: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close email writer: %v", err)
	}

	return client.Quit()
}

// TestSMS 测试短信发送
func (s *SystemSettingService) TestSMS(phone string) error {
	getVal := func(groupKey, key string) string {
		setting, err := s.repo.GetByKey(groupKey, key)
		if err != nil || setting.Value == "" {
			return ""
		}
		val := setting.Value
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		return val
	}

	enabled := getVal("sms", "sms_enabled")
	if enabled != "true" {
		return errors.New("SMS service is not enabled")
	}

	accessKey := getVal("sms", "sms_access_key")
	secretKey := getVal("sms", "sms_secret_key")
	signName := getVal("sms", "sms_sign_name")
	templateCode := getVal("sms", "sms_template_code")

	if accessKey == "" || secretKey == "" || signName == "" || templateCode == "" {
		return errors.New("SMS configuration is incomplete, please check access_key/secret_key/sign_name/template_code")
	}

	// TODO: 实际对接短信服务商 API
	// 这里仅验证配置完整性，实际发送需要集成阿里云/腾讯云 SDK
	return errors.New("SMS sending is not yet implemented")
}

// toSettingItem 将模型转为响应项
func toSettingItem(s model.SystemSetting) SettingItem {
	item := SettingItem{
		Key:         s.Key,
		Value:       parseValue(s.Value, s.Type),
		Label:       s.Label,
		Type:        s.Type,
		Tip:         s.Tip,
		Sort:        s.Sort,
		IsPublic:    s.IsPublic == 1,
		IsSensitive: s.IsSensitive == 1,
	}

	// 解析 options
	if s.Options != "" {
		var opts interface{}
		if err := json.Unmarshal([]byte(s.Options), &opts); err == nil {
			item.Options = opts
		}
	}

	// 敏感字段脱敏
	if s.IsSensitive == 1 {
		if strVal, ok := item.Value.(string); ok && strVal != "" {
			item.Value = "******"
		} else if strVal == "" {
			item.Value = ""
		}
	}

	return item
}

// parseValue 将 JSON 字符串解析为对应类型的值
func parseValue(val string, typ string) interface{} {
	if val == "" {
		switch typ {
		case "number":
			return 0
		case "boolean":
			return false
		case "json":
			return nil
		default:
			return ""
		}
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// 解析失败返回原始字符串
		return val
	}
	return result
}

// toJSONValue 将任意值转为 JSON 字符串用于存储
func toJSONValue(val interface{}, typ string) (string, error) {
	switch typ {
	case "string", "select":
		// 字符串类型：确保以 JSON 字符串格式存储
		if str, ok := val.(string); ok {
			b, err := json.Marshal(str)
			if err != nil {
				return "", err
			}
			return string(b), nil
		}
		b, err := json.Marshal(val)
		if err != nil {
			return "", err
		}
		return string(b), nil

	case "number":
		// 数字类型
		switch v := val.(type) {
		case float64:
			return fmt.Sprintf("%g", v), nil
		case int:
			return fmt.Sprintf("%d", v), nil
		case int64:
			return fmt.Sprintf("%d", v), nil
		case string:
			// 验证是否为数字
			var num json.Number
			if err := json.Unmarshal([]byte(v), &num); err == nil {
				return v, nil
			}
			return "", fmt.Errorf("not a valid number: %v", v)
		default:
			b, err := json.Marshal(val)
			if err != nil {
				return "", err
			}
			return string(b), nil
		}

	case "boolean":
		// 布尔类型
		switch v := val.(type) {
		case bool:
			if v {
				return "true", nil
			}
			return "false", nil
		case string:
			lower := strings.ToLower(v)
			if lower == "true" || lower == "false" {
				return lower, nil
			}
			return "", fmt.Errorf("not a valid boolean: %v", v)
		default:
			return "", fmt.Errorf("not a valid boolean: %v", v)
		}

	case "json", "array":
		// JSON/Array 类型：直接存储
		switch v := val.(type) {
		case string:
			// 验证是否为合法 JSON
			var js interface{}
			if err := json.Unmarshal([]byte(v), &js); err != nil {
				return "", fmt.Errorf("not valid JSON: %v", v)
			}
			return v, nil
		default:
			b, err := json.Marshal(val)
			if err != nil {
				return "", err
			}
			return string(b), nil
		}

	default:
		b, err := json.Marshal(val)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}
