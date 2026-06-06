package sms

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// AliyunSender 阿里云短信发送器
type AliyunSender struct {
	accessKeyID     string
	accessKeySecret string
	signName        string
	templateCode    string
}

// NewAliyunSender 创建阿里云短信发送器
func NewAliyunSender(cfg *Config) *AliyunSender {
	return &AliyunSender{
		accessKeyID:     cfg.AliyunAccessKeyID,
		accessKeySecret: cfg.AliyunAccessKeySecret,
		signName:        cfg.AliyunSignName,
		templateCode:    cfg.AliyunTemplateCode,
	}
}

// Send 发送短信验证码
func (s *AliyunSender) Send(phone, code string) error {
	params := map[string]string{
		"AccessKeyId":      s.accessKeyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phone,
		"RegionId":         "cn-hangzhou",
		"SignName":         s.signName,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"SignatureVersion": "1.0",
		"TemplateCode":     s.templateCode,
		"TemplateParam":    fmt.Sprintf(`{"code":"%s"}`, code),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}

	// 计算签名
	params["Signature"] = s.signRequest(params)

	// 构建请求 URL
	apiURL := "https://dysmsapi.aliyuncs.com/"
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	reqURL := apiURL + "?" + query.Encode()

	// 发送请求
	resp, err := http.Get(reqURL)
	if err != nil {
		return fmt.Errorf("阿里云短信请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析阿里云短信响应失败: %w", err)
	}

	if code, ok := result["Code"].(string); ok && code != "OK" {
		msg, _ := result["Message"].(string)
		return fmt.Errorf("阿里云短信发送失败: %s - %s", code, msg)
	}

	return nil
}

// signRequest 计算阿里云 API 签名（HMAC-SHA1）
func (s *AliyunSender) signRequest(params map[string]string) string {
	// 1. 按参数名排序
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "Signature" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 构建规范化查询字符串
	var queryParts []string
	for _, k := range keys {
		queryParts = append(queryParts, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}
	canonicalQuery := strings.Join(queryParts, "&")

	// 3. 构建待签名字符串
	stringToSign := "GET&" + url.QueryEscape("/") + "&" + url.QueryEscape(canonicalQuery)

	// 4. HMAC-SHA1 签名
	mac := hmac.New(sha1.New, []byte(s.accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature
}
