package storage

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// defaultImageExts 默认图片扩展名
var defaultImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".bmp": true, ".webp": true, ".svg": true, ".ico": true,
	".tiff": true, ".tif": true, ".heic": true, ".heif": true,
}

// defaultVideoExts 默认视频扩展名
var defaultVideoExts = map[string]bool{
	".mp4": true, ".avi": true, ".mov": true, ".wmv": true,
	".flv": true, ".mkv": true, ".webm": true, ".m4v": true,
	".mpeg": true, ".mpg": true, ".3gp": true,
}

// defaultAudioExts 默认音频扩展名
var defaultAudioExts = map[string]bool{
	".mp3": true, ".wav": true, ".flac": true, ".aac": true,
	".ogg": true, ".wma": true, ".m4a": true, ".opus": true,
}

// defaultDocumentExts 默认文档扩展名
var defaultDocumentExts = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true, ".xls": true,
	".xlsx": true, ".ppt": true, ".pptx": true, ".txt": true,
	".md": true, ".csv": true, ".rtf": true,
}

// defaultArchiveExts 默认压缩包扩展名
var defaultArchiveExts = map[string]bool{
	".zip": true, ".rar": true, ".7z": true, ".tar": true,
	".gz": true, ".bz2": true, ".xz": true,
}

// AutoTagger 自动打标签器
type AutoTagger struct {
	mu           sync.RWMutex
	extTypeMap   map[string]string // extension -> file_type（包含所有类型）
	imageExts    map[string]bool
	videoExts    map[string]bool
	audioExts    map[string]bool
	documentExts map[string]bool
	archiveExts  map[string]bool
	customExts   map[string]string // 自定义类型的扩展名 -> 类型
}

// 全局 AutoTagger 实例
var globalAutoTagger = NewAutoTagger()

// GetGlobalAutoTagger 获取全局 AutoTagger 实例
func GetGlobalAutoTagger() *AutoTagger {
	return globalAutoTagger
}

// NewAutoTagger 创建自动打标签器（使用默认硬编码扩展名）
func NewAutoTagger() *AutoTagger {
	t := &AutoTagger{
		extTypeMap:   make(map[string]string),
		imageExts:    copyMap(defaultImageExts),
		videoExts:    copyMap(defaultVideoExts),
		audioExts:    copyMap(defaultAudioExts),
		documentExts: copyMap(defaultDocumentExts),
		archiveExts:  copyMap(defaultArchiveExts),
		customExts:   make(map[string]string),
	}
	t.buildExtTypeMap()
	return t
}

// copyMap 复制 map
func copyMap(src map[string]bool) map[string]bool {
	dst := make(map[string]bool, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// FileTypeRuleData 文件类型规则数据（用于从数据库加载）
type FileTypeRuleData struct {
	Extension string
	FileType  string
}

// LoadFromDB 从数据库加载文件类型规则，替换当前的扩展名映射
func (t *AutoTagger) LoadFromDB(rules []FileTypeRuleData) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 重置为默认值
	t.imageExts = copyMap(defaultImageExts)
	t.videoExts = copyMap(defaultVideoExts)
	t.audioExts = copyMap(defaultAudioExts)
	t.documentExts = copyMap(defaultDocumentExts)
	t.archiveExts = copyMap(defaultArchiveExts)
	t.customExts = make(map[string]string)

	// 用数据库规则覆盖
	for _, rule := range rules {
		ext := strings.ToLower(strings.TrimSpace(rule.Extension))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}

		switch rule.FileType {
		case "image":
			t.imageExts[ext] = true
		case "video":
			t.videoExts[ext] = true
		case "audio":
			t.audioExts[ext] = true
		case "document":
			t.documentExts[ext] = true
		case "archive":
			t.archiveExts[ext] = true
		default:
			// 自定义类型
			t.customExts[ext] = rule.FileType
		}
	}

	t.buildExtTypeMap()
	log.Printf("[INFO] AutoTagger 已从数据库加载 %d 条文件类型规则（含 %d 个自定义类型）",
		len(rules), len(t.customExts))
}

// buildExtTypeMap 构建扩展名到类型的快速查找映射
func (t *AutoTagger) buildExtTypeMap() {
	t.extTypeMap = make(map[string]string)
	for ext := range t.imageExts {
		t.extTypeMap[ext] = "image"
	}
	for ext := range t.videoExts {
		t.extTypeMap[ext] = "video"
	}
	for ext := range t.audioExts {
		t.extTypeMap[ext] = "audio"
	}
	for ext := range t.documentExts {
		t.extTypeMap[ext] = "document"
	}
	for ext := range t.archiveExts {
		t.extTypeMap[ext] = "archive"
	}
	// 添加自定义类型
	for ext, fileType := range t.customExts {
		t.extTypeMap[ext] = fileType
	}
}


// GenerateTags 根据文件信息生成标签
func (t *AutoTagger) GenerateTags(filename, contentType, source string) []RoutingTag {
	tags := []RoutingTag{}

	// 1. 文件类型标签
	ext := strings.ToLower(filepath.Ext(filename))
	typeTag := t.getTypeTag(ext, contentType)
	tags = append(tags, RoutingTag{Key: "type", Value: typeTag})

	// 2. 来源标签
	if source != "" {
		tags = append(tags, RoutingTag{Key: "source", Value: source})
	} else {
		tags = append(tags, RoutingTag{Key: "source", Value: "user"})
	}

	// 3. 默认敏感度标签
	tags = append(tags, RoutingTag{Key: "sensitivity", Value: "public"})

	return tags
}

// GenerateTagsWithPurpose 根据文件信息和用途生成标签
func (t *AutoTagger) GenerateTagsWithPurpose(filename, contentType, source, purpose string) []RoutingTag {
	tags := t.GenerateTags(filename, contentType, source)

	// 添加用途标签
	if purpose != "" {
		tags = append(tags, RoutingTag{Key: "purpose", Value: purpose})
	}

	return tags
}

// getTypeTag 根据扩展名和MIME类型判断文件类型
func (t *AutoTagger) getTypeTag(ext, contentType string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 优先根据扩展名判断（使用快速查找映射）
	if fileType, ok := t.extTypeMap[ext]; ok {
		return fileType
	}

	// 根据MIME类型判断
	if strings.HasPrefix(contentType, "image/") {
		return "image"
	}
	if strings.HasPrefix(contentType, "video/") {
		return "video"
	}
	if strings.HasPrefix(contentType, "audio/") {
		return "audio"
	}
	if strings.HasPrefix(contentType, "application/pdf") {
		return "document"
	}

	return "other"
}

// GenerateObjectKey 生成存储对象键（使用 UUID 随机名称）
func (t *AutoTagger) GenerateObjectKey(pathPrefix, filename string) string {
	// 格式: {pathPrefix}{yyyy/MM/dd}/{uuid}{ext}
	now := time.Now()
	ext := filepath.Ext(filename)
	randomName := uuid.New().String()

	return fmt.Sprintf("%s%s/%s/%s%s",
		pathPrefix,
		now.Format("2006"),
		now.Format("01/02"),
		randomName,
		ext,
	)
}

// GenerateObjectKeyWithHash 生成带哈希的存储对象键
func (t *AutoTagger) GenerateObjectKeyWithHash(pathPrefix, filename, hash string) string {
	// 格式: {pathPrefix}{yyyy/MM/dd}/{hash}{ext}
	now := time.Now()
	ext := filepath.Ext(filename)

	// 取哈希的前16位
	hashPrefix := hash
	if len(hash) > 16 {
		hashPrefix = hash[:16]
	}

	return fmt.Sprintf("%s%s/%s/%s%s",
		pathPrefix,
		now.Format("2006"),
		now.Format("01/02"),
		hashPrefix,
		ext,
	)
}
