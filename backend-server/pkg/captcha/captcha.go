package captcha

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend-server/pkg/database"
)

// CaptchaConfig 验证码配置（从数据库加载）
type CaptchaConfig struct {
	// 通用配置
	Enabled      bool
	Type         string
	Expire       int // 秒
	MaxFail      int
	LoginTrigger int
	MinDuration  int // 毫秒

	// 数字验证码
	NumericLength  int
	NumericWidth   int
	NumericHeight  int

	// 滑块验证码
	SliderWidth     int
	SliderHeight    int
	SliderTolerance int

	// 拼图验证码
	PuzzleWidth           int
	PuzzleHeight          int
	PuzzleTolerance       int
	PuzzleVerticalRandom  bool

	// 旋转验证码
	RotationSize        int
	RotationThumbSize   int
	RotationTolerance   int

	// 点选验证码
	PointWidth     int
	PointHeight    int
	PointCount     int
	PointTolerance int
}

// CaptchaData 验证码数据（通用返回）
type CaptchaData struct {
	CaptchaID string   `json:"captcha_id"`
	Image     string   `json:"image"`               // 主图片（base64 或 URL）
	Thumb     string   `json:"thumb,omitempty"`      // 缩略图（base64）
	ThumbX    int      `json:"thumb_x,omitempty"`    // 缩略图初始 X 位置（滑块/拼图）
	ThumbY    int      `json:"thumb_y,omitempty"`    // 缩略图初始 Y 位置（滑块/拼图）
	Type      string   `json:"type"`                 // 验证码类型
	HintText  string   `json:"hint_text,omitempty"`  // 点选验证码的提示文字
	Chars     []string `json:"chars,omitempty"`      // 点选验证码的字符列表
	Width     int      `json:"width,omitempty"`      // 图片宽度（前端可能需要）
	Height    int      `json:"height,omitempty"`     // 图片高度（前端可能需要）
	Length    int      `json:"length,omitempty"`     // 数字验证码长度
	StartTime int64    `json:"start_time,omitempty"` // 验证码生成时间戳（毫秒，用于时间检测）
}

// CaptchaStore 存储在 Redis 中的数据
type CaptchaStore struct {
	Type      string `json:"type"`
	Answer    string `json:"answer"`     // JSON 格式的答案（加密）
	StartTime int64  `json:"start_time"` // 生成时间戳（毫秒）
	Used      bool   `json:"used"`       // 是否已使用
}

// AES 加密密钥（固定 32 字节，通过 SetSecret 在启动时从配置加载）
var captchaSecret = []byte("captcha_secret_key_32bytes!!!!!1")

// SetSecret 设置验证码加密密钥（应在启动时调用，使用配置文件或环境变量中的值）
func SetSecret(key []byte) {
	if len(key) >= 32 {
		captchaSecret = key[:32]
	}
}

const captchaKeyPrefix = "captcha:"

// vue-vben-admin 静态图片资源
const (
	defaultAvatarURL   = "https://unpkg.com/@vbenjs/static-source@0.1.7/source/avatar-v1.webp"
	defaultPuzzleURL   = "https://unpkg.com/@vbenjs/static-source@0.1.7/source/pro-avatar.webp"
	defaultCaptchaURL  = "https://unpkg.com/@vbenjs/static-source@0.1.7/source/default-captcha-image.jpeg"
	defaultHintImgURL  = "https://unpkg.com/@vbenjs/static-source@0.1.7/source/default-hint-image.png"
)

// 配置缓存
var configCache *CaptchaConfig
var configCacheOnce sync.Once
var configCacheMutex sync.RWMutex

// GetConfig 获取验证码配置（带缓存）
func GetConfig() *CaptchaConfig {
	configCacheMutex.RLock()
	if configCache != nil {
		configCacheMutex.RUnlock()
		return configCache
	}
	configCacheMutex.RUnlock()

	configCacheMutex.Lock()
	defer configCacheMutex.Unlock()

	if configCache != nil {
		return configCache
	}

	configCache = loadConfigFromDB()
	return configCache
}

// loadConfigFromDB 从数据库加载验证码配置
func loadConfigFromDB() *CaptchaConfig {
	db := database.GetMySQL()

	// 查询 captcha 分组的所有配置（key 是 MySQL 保留字，需要反引号）
	rows, err := db.Raw("SELECT `key`, value FROM sys_system_settings WHERE group_key = 'captcha' AND deleted_at IS NULL").Rows()
	if err != nil {
		// 返回默认配置
		return getDefaultConfig()
	}
	defer rows.Close()

	config := getDefaultConfig()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}

		// 解析 JSON 值（去除引号）
		value = strings.Trim(value, "\"")

		switch key {
		// 通用配置
		case "captcha_enabled":
			config.Enabled = value == "true"
		case "captcha_type":
			config.Type = value
		case "captcha_expire":
			config.Expire = parseInt(value, 120)
		case "captcha_max_fail":
			config.MaxFail = parseInt(value, 5)
		case "captcha_login_trigger":
			config.LoginTrigger = parseInt(value, 3)
		case "captcha_min_duration":
			config.MinDuration = parseInt(value, 500)

		// 数字验证码
		case "numeric_length":
			config.NumericLength = parseInt(value, 4)
		case "numeric_width":
			config.NumericWidth = parseInt(value, 160)
		case "numeric_height":
			config.NumericHeight = parseInt(value, 60)

		// 滑块验证码
		case "slider_width":
			config.SliderWidth = parseInt(value, 320)
		case "slider_height":
			config.SliderHeight = parseInt(value, 200)
		case "slider_tolerance":
			config.SliderTolerance = parseInt(value, 5)

		// 拼图验证码
		case "puzzle_width":
			config.PuzzleWidth = parseInt(value, 320)
		case "puzzle_height":
			config.PuzzleHeight = parseInt(value, 200)
		case "puzzle_tolerance":
			config.PuzzleTolerance = parseInt(value, 5)
		case "puzzle_vertical_random":
			config.PuzzleVerticalRandom = value == "true"

		// 旋转验证码
		case "rotation_size":
			config.RotationSize = parseInt(value, 220)
		case "rotation_thumb_size":
			config.RotationThumbSize = parseInt(value, 80)
		case "rotation_tolerance":
			config.RotationTolerance = parseInt(value, 10)

		// 点选验证码
		case "point_width":
			config.PointWidth = parseInt(value, 320)
		case "point_height":
			config.PointHeight = parseInt(value, 220)
		case "point_count":
			config.PointCount = parseInt(value, 4)
		case "point_tolerance":
			config.PointTolerance = parseInt(value, 30)
		}
	}

	return config
}

// getDefaultConfig 返回默认配置
func getDefaultConfig() *CaptchaConfig {
	return &CaptchaConfig{
		Enabled:      true,
		Type:         "slider",
		Expire:       120,
		MaxFail:      5,
		LoginTrigger: 3,
		MinDuration:  500,

		NumericLength:  4,
		NumericWidth:   160,
		NumericHeight:  60,

		SliderWidth:     320,
		SliderHeight:    200,
		SliderTolerance: 5,

		PuzzleWidth:          320,
		PuzzleHeight:         200,
		PuzzleTolerance:      5,
		PuzzleVerticalRandom: true,

		RotationSize:        220,
		RotationThumbSize:   80,
		RotationTolerance:   10,

		PointWidth:     320,
		PointHeight:    220,
		PointCount:     4,
		PointTolerance: 30,
	}
}

// parseInt 解析整数，失败返回默认值
func parseInt(s string, defaultVal int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

// RefreshConfig 刷新配置缓存
func RefreshConfig() {
	configCacheMutex.Lock()
	configCache = loadConfigFromDB()
	configCacheMutex.Unlock()
}

// Generate 根据类型生成验证码（使用配置）
func Generate(captchaType string) (*CaptchaData, error) {
	config := GetConfig()

	var data *CaptchaData
	var err error

	switch captchaType {
	case "slider":
		data, err = generateGoSlide(config)
	case "puzzle":
		data, err = generateGoPuzzle(config)
	case "rotation":
		data, err = generateGoRotation(config)
	case "point":
		data, err = generateGoClick(config)
	default:
		data, err = generateGoNumeric(config)
	}

	if err != nil {
		return nil, err
	}

	// 返回验证码生成时间（与 Redis 中存储的 StartTime 一致，用于前端时间检测）
	data.StartTime = time.Now().UnixMilli()
	return data, nil
}

// Verify 根据类型验证验证码（使用配置）
func Verify(captchaID, captchaCode string, startTime int64, points []Point) (bool, string) {
	if captchaID == "" {
		return false, "验证码ID不能为空"
	}

	config := GetConfig()
	rdb := database.GetRedis()
	ctx := context.Background()
	key := captchaKeyPrefix + captchaID

	// 获取并删除（一次性）
	dataStr, err := rdb.GetDel(ctx, key).Result()
	if err != nil {
		return false, "验证码已过期或不存在"
	}

	var store CaptchaStore
	if err := json.Unmarshal([]byte(dataStr), &store); err != nil {
		return false, "验证码数据异常"
	}

	// 一次性校验
	if store.Used {
		return false, "验证码已使用"
	}

	// 时间检测：使用配置的最短操作时间
	if startTime > 0 && config.MinDuration > 0 {
		elapsed := startTime - store.StartTime
		if elapsed < int64(config.MinDuration) {
			return false, "操作过于迅速，请重试"
		}
	}

	switch store.Type {
	case "slider":
		return verifySliderAnswer(store.Answer, captchaCode, config.SliderTolerance)
	case "puzzle":
		return verifyPuzzleAnswer(store.Answer, captchaCode, config.PuzzleTolerance)
	case "rotation":
		return verifyRotationAnswer(store.Answer, captchaCode, config.RotationTolerance)
	case "point":
		return verifyPointAnswer(store.Answer, points, config.PointTolerance)
	default:
		code := decryptAnswer(store.Answer)
		if strings.EqualFold(code, captchaCode) {
			return true, "验证通过"
		}
		return false, "验证码错误"
	}
}

// saveToRedis 保存验证码到 Redis（使用配置的过期时间）
func saveToRedis(captchaID, captchaType, answer string) error {
	config := GetConfig()
	rdb := database.GetRedis()
	ctx := context.Background()
	key := captchaKeyPrefix + captchaID

	store := CaptchaStore{
		Type:      captchaType,
		Answer:    encryptAnswer(answer),
		StartTime: time.Now().UnixMilli(),
		Used:      false,
	}
	data, _ := json.Marshal(store)

	ttl := time.Duration(config.Expire) * time.Second
	return rdb.Set(ctx, key, string(data), ttl).Err()
}

// encryptAnswer 使用 AES-GCM 加密答案
func encryptAnswer(plaintext string) string {
	block, err := aes.NewCipher(captchaSecret)
	if err != nil {
		return plaintext
	}
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// decryptAnswer 使用 AES-GCM 解密答案
func decryptAnswer(encoded string) string {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encoded
	}
	block, err := aes.NewCipher(captchaSecret)
	if err != nil {
		return encoded
	}
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return encoded
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return encoded
	}
	return string(plaintext)
}

// generateID 生成唯一 ID
func generateID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

// randInt 生成 [min, max) 范围的随机整数
func randInt(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	return min + int(n.Int64())
}

// toBase64 将图片编码为 base64 data URL
func toBase64(img image.Image) (string, error) {
	var buf strings.Builder
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if err := png.Encode(encoder, img); err != nil {
		return "", err
	}
	encoder.Close()
	return "data:image/png;base64," + buf.String(), nil
}

// abs 绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// clamp 将值限制在 0-255
func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
