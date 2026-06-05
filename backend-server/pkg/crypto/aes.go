package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"backend-server/config"
)

// Encrypt AES-GCM 加密（认证加密，防 Padding Oracle 攻击）
func Encrypt(plaintext string) (string, error) {
	cfg := config.Get().Crypto
	key := []byte(cfg.AESKey)

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", fmt.Errorf("AES 密钥长度必须为 16、24 或 32 字节，当前为 %d 字节", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建 GCM 失败: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成 nonce 失败: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt AES-GCM 解密
func Decrypt(ciphertextStr string) (string, error) {
	cfg := config.Get().Crypto
	key := []byte(cfg.AESKey)

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", fmt.Errorf("AES 密钥长度必须为 16、24 或 32 字节，当前为 %d 字节", len(key))
	}

	data, err := base64.StdEncoding.DecodeString(ciphertextStr)
	if err != nil {
		return "", fmt.Errorf("Base64 解码失败: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建 GCM 失败: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文太短")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败（密文或密钥错误）: %w", err)
	}

	return string(plaintext), nil
}
