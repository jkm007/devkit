package fileutil

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// AllowedExts 默认允许的文件扩展名白名单
// 按类别组织，方便维护
var AllowedExts = map[string]bool{
	// 图片
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
	".svg":  true,
	".ico":  true,
	".tiff": true,
	".tif":  true,
	".avif": true,
	".heic": true,
	".heif": true,
	".raw":  true,
	".cr2":  true,
	".nef":  true,

	// 视频
	".mp4":  true,
	".avi":  true,
	".mov":  true,
	".wmv":  true,
	".flv":  true,
	".mkv":  true,
	".webm": true,
	".m4v":  true,
	".3gp":  true,
	".mts":  true,
	".m2ts": true,

	// 音频
	".mp3":  true,
	".wav":  true,
	".flac": true,
	".aac":  true,
	".ogg":  true,
	".wma":  true,
	".m4a":  true,
	".opus": true,
	".amr":  true,

	// 文档
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".odt":  true,
	".ods":  true,
	".odp":  true,
	".csv":  true,
	".rtf":  true,
	".txt":  true,
	".md":   true,
	".json": true,
	".xml":  true,
	".yaml": true,
	".yml":  true,
	".toml": true,
	".ini":  true,
	".log":  true,

	// 压缩包
	".zip": true,
	".rar": true,
	".7z":  true,
	".tar": true,
	".gz":  true,
	".bz2": true,
	".xz":  true,
	".tgz": true,
	".zst": true,

	// 代码/开发
	".html":  true,
	".htm":   true,
	".css":   true,
	".js":    true,
	".ts":    true,
	".jsx":   true,
	".tsx":   true,
	".vue":   true,
	".go":    true,
	".py":    true,
	".java":  true,
	".c":     true,
	".cpp":   true,
	".h":     true,
	".rs":    true,
	".rb":    true,
	".sh":    true,
	".sql":   true,
	".swift": true,
	".kt":    true,

	// 字体
	".ttf":   true,
	".otf":   true,
	".woff":  true,
	".woff2": true,
	".eot":   true,

	// 3D / 设计
	".psd":  true,
	".ai":   true,
	".sketch": true,
	".fig":  true,
	".obj":  true,
	".stl":  true,
	".step": true,
	".iges": true,

	// 其他
	".apk":  true,
	".ipa":  true,
	".dmg":  true,
	".deb":  true,
	".rpm":  true,
}

// DangerousExts 危险扩展名黑名单
// 这些扩展名无论白名单中是否存在，都应被拒绝
var DangerousExts = map[string]bool{
	".php":   true,
	".php3":  true,
	".php4":  true,
	".php5":  true,
	".phtml": true,
	".jsp":   true,
	".jspx":  true,
	".asp":   true,
	".aspx":  true,
	".exe":   true,
	".bat":   true,
	".cmd":   true,
	".com":   true,
	".msi":   true,
	".scr":   true,
	".pif":   true,
	".vbs":   true,
	".vbe":   true,
	".ws":    true,
	".wsf":   true,
	".wsc":   true,
	".ps1":   true,
	".psm1":  true,
	".reg":   true,
	".inf":   true,
	".lnk":   true,
	".hta":   true,
	".cpl":   true,
	".msc":   true,
	".gadget": true,
	".application": true,
}

// mimeExtMapping 扩展名到预期 MIME 类型的映射
// 用于交叉验证：扩展名声称的类型与 Magic Bytes 检测的类型是否一致
var mimeExtMapping = map[string][]string{
	".jpg":  {"image/jpeg"},
	".jpeg": {"image/jpeg"},
	".png":  {"image/png"},
	".gif":  {"image/gif"},
	".bmp":  {"image/bmp"},
	".webp": {"image/webp"},
	".pdf":  {"application/pdf"},
	".zip":  {"application/zip", "application/x-zip-compressed"},
	".rar":  {"application/x-rar-compressed", "application/vnd.rar"},
	".7z":   {"application/x-7z-compressed"},
	".gz":   {"application/gzip", "application/x-gzip"},
	".tar":  {"application/x-tar"},
	".mp3":  {"audio/mpeg"},
	".mp4":  {"video/mp4"},
	".avi":  {"video/x-msvideo", "video/avi"},
	".mov":  {"video/quicktime"},
	".wmv":  {"video/x-ms-wmv"},
	".flv":  {"video/x-flv"},
	".mkv":  {"video/x-matroska"},
	".webm": {"video/webm"},
	".doc":  {"application/msword"},
	".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	".xls":  {"application/vnd.ms-excel"},
	".xlsx": {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	".ppt":  {"application/vnd.ms-powerpoint"},
	".pptx": {"application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	".ico":  {"image/x-icon", "image/vnd.microsoft.icon"},
	".tiff": {"image/tiff"},
	".tif":  {"image/tiff"},
	".rtf":  {"application/rtf", "text/rtf"},
	".csv":  {"text/csv", "text/plain"},
	".txt":  {"text/plain"},
	".html": {"text/html"},
	".htm":  {"text/html"},
	".xml":  {"text/xml", "application/xml"},
	".json": {"application/json"},
	".svg":  {"image/svg+xml"},
	".wav":  {"audio/wav", "audio/x-wav"},
	".ogg":  {"audio/ogg", "video/ogg"},
	".flac": {"audio/flac"},
	".aac":  {"audio/aac"},
	".m4a":  {"audio/mp4", "audio/x-m4a"},
	".opus": {"audio/opus"},
	".otf":  {"font/otf", "application/x-font-otf"},
	".ttf":  {"font/ttf", "application/x-font-ttf"},
	".woff": {"font/woff", "application/font-woff"},
	".woff2": {"font/woff2"},
	".md":   {"text/markdown", "text/plain", "text/x-markdown"},
}

// ValidateExtension 校验文件扩展名是否在白名单中
// fileName: 原始文件名
// 返回: (allowed bool, reason string)
func ValidateExtension(fileName string) (bool, string) {
	ext := strings.ToLower(filepath.Ext(fileName))

	if ext == "" {
		return false, "文件缺少扩展名"
	}

	// 检查危险扩展名黑名单
	if DangerousExts[ext] {
		return false, fmt.Sprintf("不支持的文件类型: %s", ext)
	}

	// 检查白名单
	if !AllowedExts[ext] {
		return false, fmt.Sprintf("不支持的文件类型: %s", ext)
	}

	return true, ""
}

// ValidateMagicBytes 校验文件内容的 Magic Bytes 是否与扩展名一致
// data: 文件开头的数据（建议至少读取 512 字节）
// fileName: 原始文件名
// 返回: (valid bool, detectedMIME string, reason string)
func ValidateMagicBytes(data []byte, fileName string) (bool, string, string) {
	if len(data) == 0 {
		return false, "", "文件内容为空"
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	detectedMIME := http.DetectContentType(data)

	// 如果检测结果为 application/octet-stream（未知），不强制拒绝
	// 因为很多合法文件格式无法通过 Magic Bytes 识别（如纯文本、SVG 等）
	if detectedMIME == "application/octet-stream" {
		// 对于无法识别的类型，允许通过（依赖扩展名白名单）
		return true, detectedMIME, ""
	}

	// 对于常见格式，进行交叉验证
	expectedMIMEs, ok := mimeExtMapping[ext]
	if !ok {
		// 扩展名不在映射表中，跳过交叉验证
		return true, detectedMIME, ""
	}

	// 检查检测到的 MIME 是否在预期列表中
	for _, expected := range expectedMIMEs {
		if detectedMIME == expected {
			return true, detectedMIME, ""
		}
	}

	// 特殊处理：text/* 类型之间允许互相匹配
	// 因为 http.DetectContentType 对文本文件的检测不太精确
	detectedBase := strings.SplitN(detectedMIME, ";", 2)[0]
	for _, expected := range expectedMIMEs {
		expectedBase := strings.SplitN(expected, ";", 2)[0]
		if strings.HasPrefix(detectedBase, "text/") && strings.HasPrefix(expectedBase, "text/") {
			return true, detectedMIME, ""
		}
	}

	// 特殊处理：application/vnd.openxmlformats-* 文件实际是 ZIP 格式
	// .docx/.xlsx/.pptx 等 Office Open XML 文件在底层就是 ZIP
	if detectedMIME == "application/zip" || detectedMIME == "application/x-zip-compressed" {
		for _, expected := range expectedMIMEs {
			if strings.HasPrefix(expected, "application/vnd.openxmlformats") ||
				strings.HasPrefix(expected, "application/vnd.ms-") {
				return true, detectedMIME, ""
			}
		}
	}

	reason := fmt.Sprintf("文件内容与扩展名不匹配: 扩展名 %s 预期 %s，实际检测为 %s",
		ext, strings.Join(expectedMIMEs, " | "), detectedMIME)
	return false, detectedMIME, reason
}

// ValidateFile 综合校验文件
// fileName: 原始文件名
// headerData: 文件开头的数据（建议至少 512 字节）
// 返回: (valid bool, reason string)
func ValidateFile(fileName string, headerData []byte) (bool, string) {
	// 1. 扩展名白名单校验
	if ok, reason := ValidateExtension(fileName); !ok {
		return false, reason
	}

	// 2. Magic Bytes 校验
	if ok, _, reason := ValidateMagicBytes(headerData, fileName); !ok {
		return false, reason
	}

	return true, ""
}
