package fileutil

import (
	"os"
	"testing"
)

func TestValidateExtension(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantOK   bool
	}{
		// 正常扩展名
		{"jpg image", "photo.jpg", true},
		{"png image", "screenshot.png", true},
		{"pdf document", "report.pdf", true},
		{"zip archive", "backup.zip", true},
		{"mp4 video", "video.mp4", true},
		{"mp3 audio", "song.mp3", true},
		{"xlsx spreadsheet", "data.xlsx", true},
		{"vue component", "App.vue", true},
		{"go source", "main.go", true},
		{"tar gz archive", "archive.tar.gz", true},

		// 大小写不敏感
		{"uppercase JPG", "PHOTO.JPG", true},
		{"mixed case Pdf", "Report.PDF", true},

		// 危险扩展名
		{"php file", "shell.php", false},
		{"jsp file", "exploit.jsp", false},
		{"exe file", "virus.exe", false},
		{"bat file", "script.bat", false},
		{"vbs file", "macro.vbs", false},
		{"hta file", "payload.hta", false},
		{"ps1 file", "script.ps1", false},

		// 缺少扩展名
		{"no extension", "Makefile", false},
		{"just dot", "file.", false},

		// 不在白名单中的扩展名
		{"unknown ext", "file.xyz123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateExtension(tt.fileName)
			if ok != tt.wantOK {
				t.Errorf("ValidateExtension(%q) = %v, want %v", tt.fileName, ok, tt.wantOK)
			}
		})
	}
}

func TestValidateMagicBytes(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		data     []byte
		wantOK   bool
	}{
		// PNG Magic Bytes
		{
			name:     "valid png",
			fileName: "image.png",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			wantOK:   true,
		},
		// JPEG Magic Bytes
		{
			name:     "valid jpeg",
			fileName: "photo.jpg",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			wantOK:   true,
		},
		// GIF Magic Bytes
		{
			name:     "valid gif",
			fileName: "animation.gif",
			data:     []byte("GIF89a\x00\x01\x00\x01"),
			wantOK:   true,
		},
		// PDF Magic Bytes
		{
			name:     "valid pdf",
			fileName: "document.pdf",
			data:     []byte("%PDF-1.4"),
			wantOK:   true,
		},
		// ZIP Magic Bytes with .zip extension
		{
			name:     "valid zip",
			fileName: "archive.zip",
			data:     []byte{0x50, 0x4B, 0x03, 0x04, 0x14, 0x00},
			wantOK:   true,
		},
		// Office Open XML (.docx is ZIP format)
		{
			name:     "docx is zip-based",
			fileName: "report.docx",
			data:     []byte{0x50, 0x4B, 0x03, 0x04, 0x14, 0x00},
			wantOK:   true,
		},
		// xlsx is also ZIP-based
		{
			name:     "xlsx is zip-based",
			fileName: "data.xlsx",
			data:     []byte{0x50, 0x4B, 0x03, 0x04},
			wantOK:   true,
		},
		// Magic bytes mismatch: PNG data in a .jpg file
		{
			name:     "png data in jpg ext",
			fileName: "fake.jpg",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			wantOK:   false,
		},
		// Empty data
		{
			name:     "empty data",
			fileName: "empty.png",
			data:     []byte{},
			wantOK:   false,
		},
		// Plain text file
		{
			name:     "plain text",
			fileName: "readme.txt",
			data:     []byte("Hello, World!\nThis is a text file."),
			wantOK:   true,
		},
		// CSV file (text-based)
		{
			name:     "csv text",
			fileName: "data.csv",
			data:     []byte("name,age,city\nJohn,30,NYC"),
			wantOK:   true,
		},
		// Unknown binary data in unknown extension (should pass - relies on whitelist)
		{
			name:     "unknown binary in unknown ext",
			fileName: "data.psd",
			data:     []byte{0x38, 0x42, 0x50, 0x53, 0x00, 0x01},
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _, _ := ValidateMagicBytes(tt.data, tt.fileName)
			if ok != tt.wantOK {
				t.Errorf("ValidateMagicBytes(%q, %d bytes) = %v, want %v", tt.fileName, len(tt.data), ok, tt.wantOK)
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		data     []byte
		wantOK   bool
	}{
		{
			name:     "valid png file",
			fileName: "image.png",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			wantOK:   true,
		},
		{
			name:     "dangerous extension rejected",
			fileName: "shell.php",
			data:     []byte("<?php echo 'hacked'; ?>"),
			wantOK:   false,
		},
		{
			name:     "extension mismatch rejected",
			fileName: "image.jpg",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateFile(tt.fileName, tt.data)
			if ok != tt.wantOK {
				t.Errorf("ValidateFile(%q) = %v, want %v", tt.fileName, ok, tt.wantOK)
			}
		})
	}
}

// TestValidateMagicBytesWithRealPng 使用真实 PNG 文件测试
func TestValidateMagicBytesWithRealPng(t *testing.T) {
	// 创建一个最小的有效 PNG 文件
	pngHeader := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length
		0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, 0x02,             // 8-bit RGB
		0x00, 0x00, 0x00,       // compression, filter, interlace
		0x90, 0x77, 0x53, 0xDE, // CRC
	}

	ok, mime, reason := ValidateMagicBytes(pngHeader, "test.png")
	if !ok {
		t.Errorf("expected valid PNG, got invalid: %s", reason)
	}
	if mime != "image/png" {
		t.Errorf("expected image/png, got %s", mime)
	}
}

// TestValidateMagicBytesWithRealJpeg 使用真实 JPEG 文件测试
func TestValidateMagicBytesWithRealJpeg(t *testing.T) {
	// 创建临时 JPEG 文件
	tmpFile, err := os.CreateTemp("", "test-*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入最小 JPEG 数据
	jpegData := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46,
		0x49, 0x46, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01,
		0x00, 0x01, 0x00, 0x00, 0xFF, 0xD9,
	}
	tmpFile.Write(jpegData)
	tmpFile.Close()

	// 读取文件头
	data, _ := os.ReadFile(tmpFile.Name())
	ok, mime, reason := ValidateMagicBytes(data[:16], "photo.jpg")
	if !ok {
		t.Errorf("expected valid JPEG, got invalid: %s", reason)
	}
	if mime != "image/jpeg" {
		t.Errorf("expected image/jpeg, got %s", mime)
	}
}
