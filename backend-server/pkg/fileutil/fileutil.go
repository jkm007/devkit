package fileutil

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

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
