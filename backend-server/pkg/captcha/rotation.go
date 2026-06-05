package captcha

import (
	"encoding/json"
	"math"
)

// RotationAnswer 已在 gocaptcha.go 中定义

// generateRotation 生成旋转验证码（旧版本，保留兼容）
// 返回头像图片 URL，前端 SliderRotateCaptcha 组件处理旋转
func generateRotation() (*CaptchaData, error) {
	captchaID, err := generateID()
	if err != nil {
		return nil, err
	}

	// 随机旋转角度 (30° ~ 330°)
	angle := float64(randInt(30, 331))

	answer, _ := json.Marshal(RotationAnswer{Angle: angle})
	startTime, err := saveToRedis(captchaID, "rotation", string(answer))
	if err != nil {
		return nil, err
	}

	return &CaptchaData{
		CaptchaID: captchaID,
		Image:     defaultAvatarURL, // 使用 vue-vben-admin 默认头像
		Type:      "rotation",
		StartTime: startTime,
	}, nil
}

// verifyRotation 验证旋转答案，容差 ±3°
// 前端 SliderRotateCaptcha 组件验证通过后，将角度提交到后端
func verifyRotation(answerStr, value string) (bool, string) {
	var answer RotationAnswer
	decrypted, err := decryptAnswer(answerStr)
	if err != nil {
		return false, "验证码数据异常"
	}
	if err := json.Unmarshal([]byte(decrypted), &answer); err != nil {
		return false, "验证码数据异常"
	}

	var userAnswer struct {
		Angle float64 `json:"angle"`
	}
	if err := json.Unmarshal([]byte(value), &userAnswer); err != nil {
		return false, "提交数据格式错误"
	}

	// 计算角度差（考虑 360° 周期）
	diff := math.Abs(userAnswer.Angle - answer.Angle)
	if diff > 180 {
		diff = 360 - diff
	}

	// 角度差 <= 3° 视为通过
	if diff <= 3 {
		return true, "验证通过"
	}
	return false, "角度偏差过大，请重试"
}
