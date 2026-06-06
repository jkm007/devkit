package sms

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// TencentSender 腾讯云短信发送器
type TencentSender struct {
	secretID    string
	secretKey   string
	appID       string
	signName    string
	templateID  string
}

// NewTencentSender 创建腾讯云短信发送器
func NewTencentSender(cfg *Config) *TencentSender {
	return &TencentSender{
		secretID:   cfg.TencentSecretID,
		secretKey:  cfg.TencentSecretKey,
		appID:      cfg.TencentAppID,
		signName:   cfg.TencentSignName,
		templateID: cfg.TencentTemplateID,
	}
}

// tencentSendRequest 腾讯云短信发送请求体
type tencentSendRequest struct {
	PhoneNumberSet []string `json:"PhoneNumberSet"`
	SmsSdkAppID    string   `json:"SmsSdkAppId"`
	Sign           string   `json:"Sign"`
	TemplateID     string   `json:"TemplateId"`
	TemplateParamSet []string `json:"TemplateParamSet"`
}

// tencentSendResponse 腾讯云短信发送响应
type tencentSendResponse struct {
	Response struct {
		RequestID  string `json:"RequestId"`
		StatusCode int    `json:"StatusCode"`
		StatusMessage string `json:"StatusMessage"`
		SendStatusSet []struct {
			SerialNo    string `json:"SerialNo"`
			PhoneNumber string `json:"PhoneNumber"`
			Fee         int    `json:"Fee"`
			Code        string `json:"Code"`
			Message     string `json:"Message"`
		} `json:"SendStatusSet"`
		Error *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error"`
	} `json:"Response"`
}

// Send 发送短信验证码
func (s *TencentSender) Send(phone, code string) error {
	// 构建请求体
	reqBody := tencentSendRequest{
		PhoneNumberSet:   []string{phone},
		SmsSdkAppID:      s.appID,
		Sign:             s.signName,
		TemplateID:       s.templateID,
		TemplateParamSet: []string{code},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构建签名
	now := time.Now()
	timestamp := now.Unix()
	authorization := s.buildAuthorization("POST", "sms.tencentcloudapi.com", "/", string(bodyBytes), timestamp)

	// 发送请求
	req, err := http.NewRequest("POST", "https://sms.tencentcloudapi.com/", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-TC-Action", "SendSms")
	req.Header.Set("X-TC-Version", "2021-01-11")
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Region", "ap-guangzhou")
	req.Header.Set("Authorization", authorization)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("腾讯云短信请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result tencentSendResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析腾讯云短信响应失败: %w", err)
	}

	if result.Response.Error != nil {
		return fmt.Errorf("腾讯云短信发送失败: %s - %s", result.Response.Error.Code, result.Response.Error.Message)
	}

	if len(result.Response.SendStatusSet) > 0 {
		status := result.Response.SendStatusSet[0]
		if status.Code != "Ok" {
			return fmt.Errorf("腾讯云短信发送失败: %s - %s", status.Code, status.Message)
		}
	}

	return nil
}

// buildAuthorization 构建腾讯云 API v3 签名
func (s *TencentSender) buildAuthorization(method, host, uri, payload string, timestamp int64) string {
	// 1. 拼接规范请求串
	httpRequestMethod := method
	canonicalURI := uri
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:application/json; charset=utf-8\nhost:%s\n", host)
	signedHeaders := "content-type;host"
	hashedPayload := sha256Hex(payload)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod, canonicalURI, canonicalQueryString,
		canonicalHeaders, signedHeaders, hashedPayload)

	// 2. 拼接待签名字符串
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/sms/tc3_request", date)
	hashedCanonicalRequest := sha256Hex(canonicalRequest)
	stringToSign := fmt.Sprintf("TC3-HMAC-SHA256\n%d\n%s\n%s",
		timestamp, credentialScope, hashedCanonicalRequest)

	// 3. 计算签名
	secretDate := hmacSHA256("TC3"+s.secretKey, date)
	secretService := hmacSHA256(secretDate, "sms")
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString([]byte(hmacSHA256(secretSigning, stringToSign)))

	// 4. 拼接 Authorization
	return fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.secretID, credentialScope, signedHeaders, signature)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, message string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return string(mac.Sum(nil))
}

// GetSortedKeys 排序 map 的 key（导出供测试用）
func GetSortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// MapToQueryString 将 map 转换为查询字符串
func MapToQueryString(m map[string]string) string {
	keys := GetSortedKeys(m)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+m[k])
	}
	return strings.Join(parts, "&")
}
