package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"backend-server/pkg/database"
	"backend-server/pkg/email"
	"backend-server/pkg/logger"
	"backend-server/pkg/sms"
)

const (
	verifyCodeKeyPrefix       = "verify_code:"
	verifyCodeCooldownPrefix  = "verify_code_cooldown:"
	verifyCodeFailPrefix      = "verify_code_fail:"

	smsCodeKeyPrefix      = "sms_code:"
	smsCodeCooldownPrefix = "sms_code_cooldown:"
	smsCodeFailPrefix     = "sms_code_fail:"

	verifyCodeExpireMinutes   = 5
	verifyCodeCooldownSeconds = 60
	verifyCodeMaxFail         = 5
	verifyCodeLockMinutes     = 15
)

// VerifyCodeService 邮箱验证码服务
type VerifyCodeService struct{}

// NewVerifyCodeService 创建验证码服务
func NewVerifyCodeService() *VerifyCodeService {
	return &VerifyCodeService{}
}

// SendCode 发送邮箱验证码
func (s *VerifyCodeService) SendCode(to, purpose string) error {
	rdb := database.GetRedis()
	ctx := context.Background()

	// 检查发送频率限制
	cooldownKey := verifyCodeCooldownPrefix + purpose + ":" + to
	exists, _ := rdb.Exists(ctx, cooldownKey).Result()
	if exists > 0 {
		return fmt.Errorf("验证码发送过于频繁，请 %d 秒后再试", verifyCodeCooldownSeconds)
	}

	// 检查失败锁定
	failKey := verifyCodeFailPrefix + purpose + ":" + to
	failCount, _ := rdb.Get(ctx, failKey).Int()
	if failCount >= verifyCodeMaxFail {
		ttl := rdb.TTL(ctx, failKey).Val()
		if ttl > 0 {
			return fmt.Errorf("验证码验证失败次数过多，请 %d 分钟后再试", int(ttl.Minutes())+1)
		}
	}

	// 生成6位随机数字验证码
	code, err := generateVerifyCode(6)
	if err != nil {
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	// 存入 Redis
	codeKey := verifyCodeKeyPrefix + purpose + ":" + to
	rdb.Set(ctx, codeKey, code, verifyCodeExpireMinutes*time.Minute)

	// 设置发送频率限制
	rdb.Set(ctx, cooldownKey, "1", verifyCodeCooldownSeconds*time.Second)

	// 获取站点名称
	siteName := getSiteName()

	// 渲染邮件模板
	purposeText := PurposeText(purpose)

	subject := fmt.Sprintf("【%s】%s验证码", siteName, purposeText)
	htmlBody := email.RenderVerifyCode(code, purposeText, verifyCodeExpireMinutes, siteName)

	// 发送邮件
	if err := email.SendHTMLEmail(to, subject, htmlBody); err != nil {
		// 发送失败，删除已存储的验证码
		rdb.Del(ctx, codeKey)
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	logger.Info(fmt.Sprintf("验证码已发送: %s, purpose: %s", to, purpose))
	return nil
}

// VerifyCode 验证邮箱验证码
func (s *VerifyCodeService) VerifyCode(to, code, purpose string) (bool, error) {
	rdb := database.GetRedis()
	ctx := context.Background()

	// 检查失败锁定
	failKey := verifyCodeFailPrefix + purpose + ":" + to
	failCount, _ := rdb.Get(ctx, failKey).Int()
	if failCount >= verifyCodeMaxFail {
		return false, fmt.Errorf("验证码验证失败次数过多，请稍后再试")
	}

	// 从 Redis 获取验证码
	codeKey := verifyCodeKeyPrefix + purpose + ":" + to
	storedCode, err := rdb.Get(ctx, codeKey).Result()
	if err != nil {
		return false, fmt.Errorf("验证码已过期或不存在")
	}

	// 比对验证码（不区分大小写）
	if storedCode != code {
		// 记录失败次数
		rdb.Incr(ctx, failKey)
		rdb.Expire(ctx, failKey, verifyCodeLockMinutes*time.Minute)
		return false, nil
	}

	// 验证成功，删除验证码和失败记录
	rdb.Del(ctx, codeKey)
	rdb.Del(ctx, failKey)

	logger.Info(fmt.Sprintf("验证码验证成功: %s, purpose: %s", to, purpose))
	return true, nil
}

// generateVerifyCode 生成指定长度的数字验证码
func generateVerifyCode(length int) (string, error) {
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = byte('0' + n.Int64())
	}
	return string(code), nil
}

// PurposeText 将 purpose 标识转为中文说明，供邮件模板等场景复用
func PurposeText(purpose string) string {
	switch purpose {
	case "register":
		return "注册账号"
	case "reset_password":
		return "重置密码"
	case "login":
		return "登录账号"
	default:
		return purpose
	}
}

// SendSMSCode 发送短信验证码
func (s *VerifyCodeService) SendSMSCode(phone, purpose string) error {
	rdb := database.GetRedis()
	ctx := context.Background()

	// 检查发送频率限制
	cooldownKey := smsCodeCooldownPrefix + purpose + ":" + phone
	exists, _ := rdb.Exists(ctx, cooldownKey).Result()
	if exists > 0 {
		return fmt.Errorf("验证码发送过于频繁，请 %d 秒后再试", verifyCodeCooldownSeconds)
	}

	// 检查失败锁定
	failKey := smsCodeFailPrefix + purpose + ":" + phone
	failCount, _ := rdb.Get(ctx, failKey).Int()
	if failCount >= verifyCodeMaxFail {
		ttl := rdb.TTL(ctx, failKey).Val()
		if ttl > 0 {
			return fmt.Errorf("验证码验证失败次数过多，请 %d 分钟后再试", int(ttl.Minutes())+1)
		}
	}

	// 生成6位随机数字验证码
	code, err := generateVerifyCode(6)
	if err != nil {
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	// 存入 Redis
	codeKey := smsCodeKeyPrefix + purpose + ":" + phone
	rdb.Set(ctx, codeKey, code, verifyCodeExpireMinutes*time.Minute)

	// 设置发送频率限制
	rdb.Set(ctx, cooldownKey, "1", verifyCodeCooldownSeconds*time.Second)

	// 获取短信发送器
	sender, err := sms.GetSender()
	if err != nil {
		rdb.Del(ctx, codeKey)
		return fmt.Errorf("短信服务不可用: %w", err)
	}

	// 发送短信
	if err := sender.Send(phone, code); err != nil {
		// 发送失败，删除已存储的验证码
		rdb.Del(ctx, codeKey)
		return fmt.Errorf("发送短信失败: %w", err)
	}

	logger.Info(fmt.Sprintf("短信验证码已发送: %s, purpose: %s", phone, purpose))
	return nil
}

// VerifySMSCode 验证短信验证码
func (s *VerifyCodeService) VerifySMSCode(phone, code, purpose string) (bool, error) {
	rdb := database.GetRedis()
	ctx := context.Background()

	// 检查失败锁定
	failKey := smsCodeFailPrefix + purpose + ":" + phone
	failCount, _ := rdb.Get(ctx, failKey).Int()
	if failCount >= verifyCodeMaxFail {
		return false, fmt.Errorf("验证码验证失败次数过多，请稍后再试")
	}

	// 从 Redis 获取验证码
	codeKey := smsCodeKeyPrefix + purpose + ":" + phone
	storedCode, err := rdb.Get(ctx, codeKey).Result()
	if err != nil {
		return false, fmt.Errorf("验证码已过期或不存在")
	}

	// 比对验证码
	if storedCode != code {
		// 记录失败次数
		rdb.Incr(ctx, failKey)
		rdb.Expire(ctx, failKey, verifyCodeLockMinutes*time.Minute)
		return false, nil
	}

	// 验证成功，删除验证码和失败记录
	rdb.Del(ctx, codeKey)
	rdb.Del(ctx, failKey)

	logger.Info(fmt.Sprintf("短信验证码验证成功: %s, purpose: %s", phone, purpose))
	return true, nil
}

// getSiteName 从数据库获取站点名称
func getSiteName() string {
	db := database.GetMySQL()
	var value string
	err := db.Raw("SELECT value FROM sys_system_settings WHERE group_key = 'basic' AND `key` = 'site_name' AND deleted_at IS NULL").Scan(&value).Error
	if err != nil || value == "" {
		return "管理系统"
	}
	// 去除 JSON 引号
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	return value
}
