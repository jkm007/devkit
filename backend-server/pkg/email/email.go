package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
	"sync"

	"backend-server/pkg/database"
)

// Config 邮件配置
type Config struct {
	Enabled  bool
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// 配置缓存
var configCache *Config
var configMutex sync.RWMutex

// GetConfig 获取邮件配置（带缓存）
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

// RefreshConfig 刷新邮件配置缓存
func RefreshConfig() {
	configMutex.Lock()
	configCache = loadConfigFromDB()
	configMutex.Unlock()
}

// loadConfigFromDB 从数据库加载邮件配置
func loadConfigFromDB() *Config {
	db := database.GetMySQL()

	rows, err := db.Raw("SELECT `key`, value FROM sys_system_settings WHERE group_key = 'email' AND deleted_at IS NULL").Rows()
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

	port, _ := strconv.Atoi(kv["smtp_port"])
	if port == 0 {
		port = 465
	}

	return &Config{
		Enabled:  kv["smtp_enabled"] == "true",
		Host:     kv["smtp_host"],
		Port:     port,
		Username: kv["smtp_username"],
		Password: kv["smtp_password"],
		From:     kv["smtp_from"],
		FromName: kv["smtp_from_name"],
	}
}

// SendEmail 发送纯文本邮件
func SendEmail(to, subject, body string) error {
	return SendHTMLEmail(to, subject, buildPlainTextBody(body))
}

// SendHTMLEmail 发送 HTML 邮件
func SendHTMLEmail(to, subject, htmlBody string) error {
	cfg := GetConfig()

	if !cfg.Enabled {
		return errors.New("邮件服务未启用")
	}
	if cfg.Host == "" || cfg.Username == "" || cfg.Password == "" {
		return errors.New("SMTP 配置不完整")
	}

	// 构建邮件头
	boundary := "=_boundary_"
	msg := buildMessage(cfg, to, subject, htmlBody, boundary)

	// 发送
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	if cfg.Port == 465 {
		return sendViaSSL(addr, cfg.Host, cfg.Username, cfg.Password, cfg.From, to, msg)
	}

	return sendViaSTARTTLS(addr, cfg.Host, cfg.Username, cfg.Password, cfg.From, to, msg)
}

// buildMessage 构建邮件消息
func buildMessage(cfg *Config, to, subject, htmlBody, boundary string) string {
	var sb strings.Builder

	sb.WriteString("From: ")
	if cfg.FromName != "" {
		sb.WriteString(cfg.FromName)
		sb.WriteString(" <")
		sb.WriteString(cfg.From)
		sb.WriteString(">")
	} else {
		sb.WriteString(cfg.From)
	}
	sb.WriteString("\r\n")

	sb.WriteString("To: ")
	sb.WriteString(to)
	sb.WriteString("\r\n")

	sb.WriteString("Subject: ")
	sb.WriteString(subject)
	sb.WriteString("\r\n")

	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(htmlBody)

	return sb.String()
}

// buildPlainTextBody 简单包装纯文本为 HTML
func buildPlainTextBody(body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
<p>%s</p>
</body>
</html>`, body)
}

// sendViaSSL 通过 SSL/TLS 发送（端口 465）
func sendViaSSL(addr, host, username, password, from, to, msg string) error {
	tlsConfig := &tls.Config{ServerName: host}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("SSL 连接失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}
	defer client.Close()

	if err = client.Auth(smtp.PlainAuth("", username, password, host)); err != nil {
		return fmt.Errorf("SMTP 认证失败: %w", err)
	}
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM 失败: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO 失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA 失败: %w", err)
	}
	if _, err = w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("关闭邮件写入失败: %w", err)
	}

	return client.Quit()
}

// sendViaSTARTTLS 通过 STARTTLS 发送（端口 587 等）
func sendViaSTARTTLS(addr, host, username, password, from, to, msg string) error {
	auth := smtp.PlainAuth("", username, password, host)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}
	return nil
}
