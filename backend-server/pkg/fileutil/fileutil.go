package fileutil

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Common MIME types and their magic bytes
var magicSignatures = map[string][][]byte{
	"image/jpeg": {{0xFF, 0xD8, 0xFF}},
	"image/png":  {{0x89, 0x50, 0x4E, 0x47}, {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}},
	"image/gif":  {{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}, {0x47, 0x49, 0x46, 0x38, 0x39, 0x61}},
	"image/webp": {{0x52, 0x49, 0x46, 0x46}},
	"application/pdf": {{0x25, 0x50, 0x44, 0x46}},
	"application/zip": {{0x50, 0x4B, 0x03, 0x04}, {0x50, 0x4B, 0x05, 0x06}, {0x50, 0x4B, 0x07, 0x08}},
	"application/x-rar": {{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07}},
	"application/x-7z-compressed": {{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}},
	"application/msword": {{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}},
	"video/mp4": {{0x66, 0x74, 0x79, 0x70}},
	"audio/mpeg": {{0x49, 0x44, 0x33}, {0xFF, 0xFB}},
}

// AllowedExtensions defines allowed file extensions by category
var AllowedExtensions = map[string][]string{
	"image": {".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"},
	"document": {".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".csv"},
	"archive": {".zip", ".rar", ".7z", ".tar", ".gz"},
	"video": {".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv"},
	"audio": {".mp3", ".wav", ".flac", ".aac", ".ogg"},
}

// ValidateExtension checks if the file extension is allowed
// Returns (isValid, reason)
func ValidateExtension(fileName string) (bool, string) {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		return false, "文件没有扩展名"
	}

	// Check all allowed extensions
	for _, exts := range AllowedExtensions {
		for _, allowedExt := range exts {
			if ext == allowedExt {
				return true, ""
			}
		}
	}

	return false, fmt.Sprintf("不支持的文件类型: %s", ext)
}

// DetectMimeType detects MIME type from magic bytes
func DetectMagicBytes(header []byte) (string, bool) {
	for mimeType, signatures := range magicSignatures {
		for _, sig := range signatures {
			if len(header) >= len(sig) && bytes.HasPrefix(header, sig) {
				return mimeType, true
			}
			// Special case for MP4 which has offset signature
			if mimeType == "video/mp4" && len(header) > 8 {
				if bytes.Contains(header[4:12], sig) {
					return mimeType, true
				}
			}
		}
	}
	return "", false
}

// ValidateMagicBytes validates file content matches its extension
// Returns (isValid, detectedMIME, reason)
func ValidateMagicBytes(header []byte, fileName string) (bool, string, string) {
	ext := strings.ToLower(filepath.Ext(fileName))

	// Map extensions to expected MIME types
	extToMIME := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".rar":  "application/x-rar",
		".7z":   "application/x-7z-compressed",
		".doc":  "application/msword",
		".docx": "application/zip", // Office Open XML is ZIP-based
		".xlsx": "application/zip",
		".pptx": "application/zip",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
	}

	expectedMIME, ok := extToMIME[ext]
	if !ok {
		// Unknown extension, skip magic byte check
		return true, "", ""
	}

	detectedMIME, detected := DetectMagicBytes(header)
	if !detected {
		return false, "", fmt.Sprintf("无法识别文件内容格式")
	}

	// Special handling for Office Open XML formats
	if ext == ".docx" || ext == ".xlsx" || ext == ".pptx" {
		if detectedMIME == "application/zip" {
			return true, detectedMIME, ""
		}
		return false, detectedMIME, fmt.Sprintf("文件格式不匹配: 期望 Office 文档, 实际为 %s", detectedMIME)
	}

	if detectedMIME != expectedMIME {
		return false, detectedMIME, fmt.Sprintf("文件格式不匹配: 期望 %s, 实际为 %s", expectedMIME, detectedMIME)
	}

	return true, detectedMIME, ""
}

// SanitizeFileName sanitizes a file name to prevent path traversal
func SanitizeFileName(fileName string) string {
	// Remove path separators to prevent directory traversal
	fileName = filepath.Base(fileName)
	// Remove null bytes
	fileName = strings.ReplaceAll(fileName, "\x00", "")
	return fileName
}

// GenerateUniqueFileName generates a unique file name with timestamp
func GenerateUniqueFileName(prefix, ext string) string {
	now := time.Now()
	return fmt.Sprintf("%s_%s%s", prefix, now.Format("20060102_150405"), ext)
}
