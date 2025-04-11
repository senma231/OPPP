package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPConfig TOTP 配置
type TOTPConfig struct {
	// 密钥长度
	SecretSize uint
	// 发行者
	Issuer string
	// 有效期（秒）
	Period uint
	// 位数
	Digits otp.Digits
	// 算法
	Algorithm otp.Algorithm
}

// DefaultTOTPConfig 默认 TOTP 配置
var DefaultTOTPConfig = TOTPConfig{
	SecretSize: 20,
	Issuer:     "P3",
	Period:     30,
	Digits:     otp.DigitsSix,
	Algorithm:  otp.AlgorithmSHA1,
}

// GenerateTOTPSecret 生成 TOTP 密钥
func GenerateTOTPSecret(username string, config TOTPConfig) (string, string, error) {
	// 生成随机密钥
	secret := make([]byte, config.SecretSize)
	_, err := rand.Read(secret)
	if err != nil {
		return "", "", fmt.Errorf("生成随机密钥失败: %w", err)
	}

	// 编码为 Base32
	secretBase32 := base32.StdEncoding.EncodeToString(secret)
	// 移除填充字符
	secretBase32 = strings.TrimRight(secretBase32, "=")

	// 生成 TOTP URI
	uri := url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   fmt.Sprintf("%s:%s", url.PathEscape(config.Issuer), url.PathEscape(username)),
	}

	params := url.Values{}
	params.Add("secret", secretBase32)
	params.Add("issuer", config.Issuer)
	params.Add("algorithm", algorithmToString(config.Algorithm))
	params.Add("digits", digitsToString(config.Digits))
	params.Add("period", fmt.Sprintf("%d", config.Period))

	uri.RawQuery = params.Encode()
	return secretBase32, uri.String(), nil
}

// VerifyTOTP 验证 TOTP 代码
func VerifyTOTP(secret string, passcode string, config TOTPConfig) (bool, error) {
	// 添加填充字符
	paddingCount := len(secret) % 8
	if paddingCount > 0 {
		secret = secret + strings.Repeat("=", 8-paddingCount)
	}

	// 验证 TOTP 代码
	valid, err := totp.ValidateCustom(
		passcode,
		secret,
		time.Now(),
		totp.ValidateOpts{
			Period:    config.Period,
			Skew:      1,
			Digits:    config.Digits,
			Algorithm: config.Algorithm,
		},
	)
	if err != nil {
		return false, fmt.Errorf("验证 TOTP 代码失败: %w", err)
	}

	return valid, nil
}

// GenerateTOTP 生成 TOTP 代码
func GenerateTOTP(secret string, config TOTPConfig) (string, error) {
	// 添加填充字符
	paddingCount := len(secret) % 8
	if paddingCount > 0 {
		secret = secret + strings.Repeat("=", 8-paddingCount)
	}

	// 生成 TOTP 代码
	passcode, err := totp.GenerateCodeCustom(
		secret,
		time.Now(),
		totp.ValidateOpts{
			Period:    config.Period,
			Skew:      1,
			Digits:    config.Digits,
			Algorithm: config.Algorithm,
		},
	)
	if err != nil {
		return "", fmt.Errorf("生成 TOTP 代码失败: %w", err)
	}

	return passcode, nil
}

// algorithmToString 将算法转换为字符串
func algorithmToString(algorithm otp.Algorithm) string {
	switch algorithm {
	case otp.AlgorithmSHA1:
		return "SHA1"
	case otp.AlgorithmSHA256:
		return "SHA256"
	case otp.AlgorithmSHA512:
		return "SHA512"
	default:
		return "SHA1"
	}
}

// digitsToString 将位数转换为字符串
func digitsToString(digits otp.Digits) string {
	switch digits {
	case otp.DigitsSix:
		return "6"
	case otp.DigitsEight:
		return "8"
	default:
		return "6"
	}
}
