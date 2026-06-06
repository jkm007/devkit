package wechat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// doWeChatRequest 发起微信 API 请求并解析 JSON 响应
func doWeChatRequest(reqURL string, result interface{}) error {
	resp, err := http.Get(reqURL)
	if err != nil {
		return fmt.Errorf("微信 API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取微信响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("微信 API 返回 HTTP %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("解析微信响应失败: %w, body: %s", err, string(body))
	}

	return nil
}
