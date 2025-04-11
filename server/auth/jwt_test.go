package auth

import (
	"testing"
	"time"
)

func TestJWTService(t *testing.T) {
	// 创建 JWT 服务
	jwtService := NewJWTService("test-secret-key")

	// 测试生成令牌
	userID := uint(123)
	role := "user"
	accessToken, refreshToken, err := jwtService.GenerateTokens(userID, role)
	if err != nil {
		t.Fatalf("生成令牌失败: %v", err)
	}

	// 检查令牌不为空
	if accessToken == "" {
		t.Error("访问令牌不应为空")
	}
	if refreshToken == "" {
		t.Error("刷新令牌不应为空")
	}

	// 验证访问令牌
	accessClaims, err := jwtService.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("验证访问令牌失败: %v", err)
	}

	// 检查访问令牌声明
	if accessClaims.UserID != userID {
		t.Errorf("访问令牌用户 ID 错误，期望 %d，实际 %d", userID, accessClaims.UserID)
	}
	if accessClaims.Role != role {
		t.Errorf("访问令牌角色错误，期望 '%s'，实际 '%s'", role, accessClaims.Role)
	}
	if accessClaims.Type != AccessToken {
		t.Errorf("访问令牌类型错误，期望 '%s'，实际 '%s'", AccessToken, accessClaims.Type)
	}

	// 验证刷新令牌
	refreshClaims, err := jwtService.ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("验证刷新令牌失败: %v", err)
	}

	// 检查刷新令牌声明
	if refreshClaims.UserID != userID {
		t.Errorf("刷新令牌用户 ID 错误，期望 %d，实际 %d", userID, refreshClaims.UserID)
	}
	if refreshClaims.Role != role {
		t.Errorf("刷新令牌角色错误，期望 '%s'，实际 '%s'", role, refreshClaims.Role)
	}
	if refreshClaims.Type != RefreshToken {
		t.Errorf("刷新令牌类型错误，期望 '%s'，实际 '%s'", RefreshToken, refreshClaims.Type)
	}
}

func TestRefreshAccessToken(t *testing.T) {
	// 创建 JWT 服务
	jwtService := NewJWTService("test-secret-key")

	// 生成令牌
	userID := uint(123)
	role := "user"
	_, refreshToken, err := jwtService.GenerateTokens(userID, role)
	if err != nil {
		t.Fatalf("生成令牌失败: %v", err)
	}

	// 使用刷新令牌生成新的访问令牌
	newAccessToken, err := jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		t.Fatalf("刷新访问令牌失败: %v", err)
	}

	// 检查新的访问令牌不为空
	if newAccessToken == "" {
		t.Error("新的访问令牌不应为空")
	}

	// 验证新的访问令牌
	accessClaims, err := jwtService.ValidateToken(newAccessToken)
	if err != nil {
		t.Fatalf("验证新的访问令牌失败: %v", err)
	}

	// 检查新的访问令牌声明
	if accessClaims.UserID != userID {
		t.Errorf("新的访问令牌用户 ID 错误，期望 %d，实际 %d", userID, accessClaims.UserID)
	}
	if accessClaims.Role != role {
		t.Errorf("新的访问令牌角色错误，期望 '%s'，实际 '%s'", role, accessClaims.Role)
	}
	if accessClaims.Type != AccessToken {
		t.Errorf("新的访问令牌类型错误，期望 '%s'，实际 '%s'", AccessToken, accessClaims.Type)
	}
}

func TestInvalidToken(t *testing.T) {
	// 创建 JWT 服务
	jwtService := NewJWTService("test-secret-key")

	// 测试无效令牌
	invalidToken := "invalid.token.string"
	_, err := jwtService.ValidateToken(invalidToken)
	if err == nil {
		t.Error("验证无效令牌应该返回错误")
	}

	// 测试使用无效令牌刷新访问令牌
	_, err = jwtService.RefreshAccessToken(invalidToken)
	if err == nil {
		t.Error("使用无效令牌刷新访问令牌应该返回错误")
	}

	// 测试使用访问令牌刷新访问令牌
	userID := uint(123)
	role := "user"
	accessToken, _, err := jwtService.GenerateTokens(userID, role)
	if err != nil {
		t.Fatalf("生成令牌失败: %v", err)
	}

	_, err = jwtService.RefreshAccessToken(accessToken)
	if err == nil {
		t.Error("使用访问令牌刷新访问令牌应该返回错误")
	}
}

func TestTokenExpiry(t *testing.T) {
	// 创建短期 JWT 服务（1 秒过期）
	shortJWTService := &JWTService{
		secretKey:     "test-secret-key",
		accessExpiry:  time.Second,
		refreshExpiry: time.Second,
	}

	// 生成令牌
	userID := uint(123)
	role := "user"
	accessToken, _, err := shortJWTService.GenerateTokens(userID, role)
	if err != nil {
		t.Fatalf("生成令牌失败: %v", err)
	}

	// 等待令牌过期
	time.Sleep(2 * time.Second)

	// 验证过期令牌
	_, err = shortJWTService.ValidateToken(accessToken)
	if err == nil {
		t.Error("验证过期令牌应该返回错误")
	}
}
