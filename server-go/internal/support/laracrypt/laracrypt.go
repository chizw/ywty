// Package laracrypt 与 Laravel Crypt::encryptString / decryptString（aes-256-cbc）双向兼容，
// 用于 settings 表中的加密字段（如 app.license_key），保证 PHP 版与 Go 版可互读数据。
package laracrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type payload struct {
	IV    string `json:"iv"`
	Value string `json:"value"`
	MAC   string `json:"mac"`
}

// EncryptString 等价 Laravel Crypt::encryptString（aes-256-cbc + PKCS7 + HMAC-SHA256）。
func EncryptString(key, value string) (string, error) {
	raw, err := decodeKey(key)
	if err != nil {
		return "", err
	}
	if len(raw) != 32 {
		return "", errors.New("laracrypt: aes-256-cbc 需要 32 字节密钥")
	}

	block, err := aes.NewCipher(raw)
	if err != nil {
		return "", err
	}
	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}
	plain := padPKCS7([]byte(value), aes.BlockSize)
	cipherText := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(cipherText, plain)

	ivB64 := base64.StdEncoding.EncodeToString(iv)
	valB64 := base64.StdEncoding.EncodeToString(cipherText)
	mac := hashHMAC(raw, ivB64, valB64)

	data, err := json.Marshal(payload{IV: ivB64, Value: valB64, MAC: mac})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// DecryptString 等价 Laravel Crypt::decryptString。
func DecryptString(key, value string) (string, error) {
	raw, err := decodeKey(key)
	if err != nil {
		return "", err
	}
	if len(raw) != 32 {
		return "", errors.New("laracrypt: aes-256-cbc 需要 32 字节密钥")
	}

	jsonBytes, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("laracrypt: 载荷不是合法 base64: %w", err)
	}
	var p payload
	if err := json.Unmarshal(jsonBytes, &p); err != nil {
		return "", fmt.Errorf("laracrypt: 载荷不是合法 JSON: %w", err)
	}
	if !hmac.Equal([]byte(p.MAC), []byte(hashHMAC(raw, p.IV, p.Value))) {
		return "", errors.New("laracrypt: MAC 校验失败")
	}

	iv, err := base64.StdEncoding.DecodeString(p.IV)
	if err != nil || len(iv) != 16 {
		return "", errors.New("laracrypt: 非法 IV")
	}
	cipherText, err := base64.StdEncoding.DecodeString(p.Value)
	if err != nil || len(cipherText) == 0 || len(cipherText)%aes.BlockSize != 0 {
		return "", errors.New("laracrypt: 非法密文长度")
	}

	block, err := aes.NewCipher(raw)
	if err != nil {
		return "", err
	}
	plain := make([]byte, len(cipherText))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, cipherText)
	plain, err = unpadPKCS7(plain, aes.BlockSize)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func hashHMAC(key []byte, ivB64, valB64 string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(ivB64 + valB64))
	return hex.EncodeToString(mac.Sum(nil))
}

func decodeKey(key string) ([]byte, error) {
	if v, ok := strings.CutPrefix(key, "base64:"); ok {
		key = v
	}
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("laracrypt: 密钥不是合法 base64: %w", err)
	}
	return decoded, nil
}

func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}
	return padded
}

func unpadPKCS7(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("laracrypt: 空密文")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, errors.New("laracrypt: PKCS7 填充非法")
	}
	for _, b := range data[len(data)-padding:] {
		if int(b) != padding {
			return nil, errors.New("laracrypt: PKCS7 填充非法")
		}
	}
	return data[:len(data)-padding], nil
}
