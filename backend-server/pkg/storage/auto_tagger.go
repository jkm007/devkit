package storage

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// AutoTagger 自动打标签器
type AutoTagger struct {
	imageExts    map[string]bool
	videoExts    map[string]bool
	audioExts    map[string]bool
	documentExts map[string]bool
	archiveExts  map[string]bool
}

// NewAutoTagger 创建自动打标签器
func NewAutoTagger() *AutoTagger {
	return &AutoTagger{
		imageExts: map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
			".bmp": true, ".webp": true, ".svg": true, ".ico": true,
			".tiff": true, ".tif": true, ".heic": true, ".heif": true,
		},
		videoExts: map[string]bool{
			".mp4": true, ".avi": true, ".mov": true, ".wmv": true,
			".flv": true, ".mkv": true, ".webm": true, ".m4v": true,
			".mpeg": true, ".mpg": true, ".3gp": true,
		},
		audioExts: map[string]bool{
			".mp3": true, ".wav": true, ".flac": true, ".aac": true,
			".ogg": true, ".wma": true, ".m4a": true, ".opus": true,
		},
		documentExts: map[string]bool{
			".pdf": true, ".doc": true, ".docx": true, ".xls": true,
			".xlsx": true, ".ppt": true, ".pptx": true, ".txt": true,
			".md": true, ".csv": true, ".rtf": true,
		},
		archiveExts: map[string]bool{
			".zip": true, ".rar": true, ".7z": true, ".tar": true,
			".gz": true, ".bz2": true, ".xz": true,
		},
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
	// 优先根据扩展名判断
	if t.imageExts[ext] {
		return "image"
	}
	if t.videoExts[ext] {
		return "video"
	}
	if t.audioExts[ext] {
		return "audio"
	}
	if t.documentExts[ext] {
		return "document"
	}
	if t.archiveExts[ext] {
		return "archive"
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

// GenerateObjectKey 生成存储对象键
func (t *AutoTagger) GenerateObjectKey(pathPrefix, filename string) string {
	// 格式: {pathPrefix}{yyyy/MM/dd}/{timestamp}{ext}
	now := time.Now()
	ext := filepath.Ext(filename)
	timestamp := now.UnixNano() / int64(time.Millisecond)

	return fmt.Sprintf("%s%s/%s/%d%s",
		pathPrefix,
		now.Format("2006"),
		now.Format("01/02"),
		timestamp,
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
