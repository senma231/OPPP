package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
)

// TLSConfig 返回 TLS 1.3 配置
func TLSConfig(isServer bool) *tls.Config {
	config := &tls.Config{
		MinVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}

	if isServer {
		// 服务端需要证书
		// TODO: 实现证书生成或加载
		config.Certificates = []tls.Certificate{}
	}

	return config
}

// AESEncrypt 使用 AES-GCM 加密数据
func AESEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建 GCM 模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 创建随机数
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESDecrypt 使用 AES-GCM 解密数据
func AESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建 GCM 模式
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 获取随机数
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("密文长度不足")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateAESKey 生成 AES 密钥
func GenerateAESKey(bits int) ([]byte, error) {
	if bits != 128 && bits != 192 && bits != 256 {
		return nil, fmt.Errorf("不支持的密钥长度: %d", bits)
	}

	key := make([]byte, bits/8)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return key, nil
}

// TOTPGenerator TOTP 一次性密码生成器
type TOTPGenerator struct {
	Secret string
	Digits int
	Period int
}

// NewTOTPGenerator 创建一个新的 TOTP 生成器
func NewTOTPGenerator(secret string, digits, period int) *TOTPGenerator {
	if digits == 0 {
		digits = 6
	}
	if period == 0 {
		period = 30
	}

	return &TOTPGenerator{
		Secret: secret,
		Digits: digits,
		Period: period,
	}
}

// GenerateCode 生成 TOTP 代码
func (g *TOTPGenerator) GenerateCode() (string, error) {
	// TODO: 实现 TOTP 代码生成
	return "123456", fmt.Errorf("TOTP 代码生成尚未实现")
}

// VerifyCode 验证 TOTP 代码
func (g *TOTPGenerator) VerifyCode(code string) (bool, error) {
	// TODO: 实现 TOTP 代码验证
	return false, fmt.Errorf("TOTP 代码验证尚未实现")
}
