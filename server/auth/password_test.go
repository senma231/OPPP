package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	// 测试密码哈希
	password := "P@ssw0rd123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("密码哈希失败: %v", err)
	}

	// 检查哈希格式
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Errorf("哈希格式错误: %s", hash)
	}

	// 检查哈希长度
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("哈希部分数量错误，期望 6，实际 %d", len(parts))
	}
}

func TestVerifyPassword(t *testing.T) {
	// 测试密码验证
	password := "P@ssw0rd123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("密码哈希失败: %v", err)
	}

	// 验证正确密码
	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("密码验证失败: %v", err)
	}
	if !valid {
		t.Error("正确密码验证应该返回 true")
	}

	// 验证错误密码
	valid, err = VerifyPassword("WrongPassword", hash)
	if err != nil {
		t.Fatalf("密码验证失败: %v", err)
	}
	if valid {
		t.Error("错误密码验证应该返回 false")
	}
}

func TestDecodeHash(t *testing.T) {
	// 测试哈希解码
	hash := "$argon2id$v=19$m=65536,t=1,p=4$c29tZXNhbHQ$c29tZWhhc2g"
	params, salt, hashBytes, err := decodeHash(hash)
	if err != nil {
		t.Fatalf("哈希解码失败: %v", err)
	}

	// 检查参数
	if params.memory != 65536 {
		t.Errorf("内存参数错误，期望 65536，实际 %d", params.memory)
	}
	if params.time != 1 {
		t.Errorf("时间参数错误，期望 1，实际 %d", params.time)
	}
	if params.threads != 4 {
		t.Errorf("线程参数错误，期望 4，实际 %d", params.threads)
	}

	// 检查盐值和哈希
	if string(salt) != "somesalt" {
		t.Errorf("盐值错误，期望 'somesalt'，实际 '%s'", string(salt))
	}
	if string(hashBytes) != "somehash" {
		t.Errorf("哈希错误，期望 'somehash'，实际 '%s'", string(hashBytes))
	}
}

func TestInvalidHash(t *testing.T) {
	// 测试无效哈希
	invalidHashes := []string{
		"invalid",
		"$invalid$v=19$m=65536,t=1,p=4$c29tZXNhbHQ$c29tZWhhc2g",
		"$argon2id$invalid$m=65536,t=1,p=4$c29tZXNhbHQ$c29tZWhhc2g",
		"$argon2id$v=19$invalid$c29tZXNhbHQ$c29tZWhhc2g",
	}

	for _, hash := range invalidHashes {
		_, err := VerifyPassword("password", hash)
		if err == nil {
			t.Errorf("无效哈希 '%s' 应该返回错误", hash)
		}
	}
}

func TestNeedsRehash(t *testing.T) {
	// 测试需要重新哈希
	password := "P@ssw0rd123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("密码哈希失败: %v", err)
	}

	// 当前参数不需要重新哈希
	needsRehash, err := NeedsRehash(hash)
	if err != nil {
		t.Fatalf("检查重新哈希失败: %v", err)
	}
	if needsRehash {
		t.Error("当前参数不应该需要重新哈希")
	}

	// 修改哈希参数
	parts := strings.Split(hash, "$")
	parts[3] = "m=32768,t=2,p=4" // 修改内存和时间参数
	modifiedHash := strings.Join(parts, "$")

	// 修改后的参数需要重新哈希
	needsRehash, err = NeedsRehash(modifiedHash)
	if err != nil {
		t.Fatalf("检查重新哈希失败: %v", err)
	}
	if !needsRehash {
		t.Error("修改后的参数应该需要重新哈希")
	}
}
