package laracrypt_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/chizw/ywty/server-go/internal/support/laracrypt"
)

// testKey 与 Laravel APP_KEY=base64:xxx 同构的 32 字节密钥。
var testKey = "base64:" + base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))

func TestEncryptDecryptRoundtrip(t *testing.T) {
	cases := []string{
		"",
		"hello",
		"LICENSE-KEY-2XNZ-0001",
		"包含中文的授权密钥测试",
		strings.Repeat("long-value-", 30),
	}
	for _, want := range cases {
		enc, err := laracrypt.EncryptString(testKey, want)
		if err != nil {
			t.Fatalf("EncryptString(%q): %v", want, err)
		}
		got, err := laracrypt.DecryptString(testKey, enc)
		if err != nil {
			t.Fatalf("DecryptString(%q): %v", want, err)
		}
		if got != want {
			t.Fatalf("roundtrip mismatch: want %q got %q", want, got)
		}
	}
}

// TestPayloadShape 校验密文结构与 Laravel Encrypter 完全一致：
// base64(json{iv,value,mac})，mac 为 HMAC-SHA256 hex(iv_b64+value_b64, key)。
func TestPayloadShape(t *testing.T) {
	enc, err := laracrypt.EncryptString(testKey, "abc")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		t.Fatalf("外层不是合法 base64: %v", err)
	}
	var p struct {
		IV    string `json:"iv"`
		Value string `json:"value"`
		MAC   string `json:"mac"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("载荷不是合法 JSON: %v", err)
	}
	if _, err := base64.StdEncoding.DecodeString(p.IV); err != nil {
		t.Fatalf("iv 不是 base64: %v", err)
	}
	if _, err := base64.StdEncoding.DecodeString(p.Value); err != nil {
		t.Fatalf("value 不是 base64: %v", err)
	}
	if len(p.IV) != 24 { // 16 字节 → base64 24 字符
		t.Fatalf("iv 长度异常: %d", len(p.IV))
	}
	if len(p.MAC) != 64 { // sha256 hex
		t.Fatalf("mac 长度异常: %d", len(p.MAC))
	}
}

func TestDecryptTamperedFails(t *testing.T) {
	enc, err := laracrypt.EncryptString(testKey, "secret")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := base64.StdEncoding.DecodeString(enc)
	var p struct {
		IV    string `json:"iv"`
		Value string `json:"value"`
		MAC   string `json:"mac"`
	}
	_ = json.Unmarshal(raw, &p)
	p.Value = "AAAA" + p.Value[4:]
	tampered, _ := json.Marshal(p)
	if _, err := laracrypt.DecryptString(testKey, base64.StdEncoding.EncodeToString(tampered)); err == nil {
		t.Fatal("篡改后的密文应解密失败")
	}
}
