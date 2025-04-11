package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// 定义 Argon2 参数
const (
	// 时间成本
	argon2Time uint32 = 1
	// 内存成本
	argon2Memory uint32 = 64 * 1024
	// 并行度
	argon2Threads uint8 = 4
	// 密钥长度
	argon2KeyLen uint32 = 32
)

var (
	// ErrInvalidHash 表示提供的哈希格式无效
	ErrInvalidHash = errors.New("提供的密码哈希格式无效")
	// ErrIncompatibleVersion 表示哈希版本不兼容
	ErrIncompatibleVersion = errors.New("不兼容的 Argon2 版本")
)

// HashPassword 使用 Argon2id 算法对密码进行哈希处理
func HashPassword(password string) (string, error) {
	// 生成随机盐值
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 使用 Argon2id 算法计算哈希
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// 编码为 Base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// 格式化哈希字符串
	// 格式: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	encodedHash := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2Memory, argon2Time, argon2Threads, b64Salt, b64Hash)

	return encodedHash, nil
}

// VerifyPassword 验证密码是否匹配存储的哈希
func VerifyPassword(password, encodedHash string) (bool, error) {
	// 解析哈希字符串
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// 使用相同的参数计算哈希
	otherHash := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, params.keyLen)

	// 比较哈希值（使用恒定时间比较以防止计时攻击）
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// argon2Params 存储 Argon2 参数
type argon2Params struct {
	memory  uint32
	time    uint32
	threads uint8
	keyLen  uint32
}

// decodeHash 解析 Argon2 哈希字符串
func decodeHash(encodedHash string) (params *argon2Params, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	// 检查算法和版本
	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrIncompatibleVersion
	}
	if parts[2] != "v=19" {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	// 解析参数
	var memory, time uint32
	var threads uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return nil, nil, nil, err
	}

	// 解码盐值
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}

	// 解码哈希
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}

	params = &argon2Params{
		memory:  memory,
		time:    time,
		threads: threads,
		keyLen:  uint32(len(hash)),
	}

	return params, salt, hash, nil
}

// NeedsRehash 检查密码哈希是否需要重新计算
// 当哈希参数变更时，可以使用此函数来确定是否需要更新哈希
func NeedsRehash(encodedHash string) (bool, error) {
	params, _, _, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// 检查参数是否匹配当前设置
	return params.memory != argon2Memory ||
		params.time != argon2Time ||
		params.threads != argon2Threads ||
		params.keyLen != argon2KeyLen, nil
}
