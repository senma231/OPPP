package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/senma231/p3/common/logger"
)

// JWT 相关常量
const (
	// AccessTokenExpiry 访问令牌过期时间（小时）
	AccessTokenExpiry = 1
	// RefreshTokenExpiry 刷新令牌过期时间（天）
	RefreshTokenExpiry = 7
)

// TokenType 令牌类型
type TokenType string

const (
	// AccessToken 访问令牌
	AccessToken TokenType = "access"
	// RefreshToken 刷新令牌
	RefreshToken TokenType = "refresh"
)

// CustomClaims 自定义 JWT 声明
type CustomClaims struct {
	UserID uint      `json:"user_id"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

// JWTService JWT 服务
type JWTService struct {
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTService 创建 JWT 服务
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey:     secretKey,
		accessExpiry:  time.Hour * AccessTokenExpiry,
		refreshExpiry: time.Hour * 24 * RefreshTokenExpiry,
	}
}

// GenerateTokens 生成访问令牌和刷新令牌
func (s *JWTService) GenerateTokens(userID uint, role string) (accessToken, refreshToken string, err error) {
	// 生成访问令牌
	accessToken, err = s.generateToken(userID, role, AccessToken, s.accessExpiry)
	if err != nil {
		return "", "", fmt.Errorf("生成访问令牌失败: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err = s.generateToken(userID, role, RefreshToken, s.refreshExpiry)
	if err != nil {
		return "", "", fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return accessToken, refreshToken, nil
}

// generateToken 生成 JWT 令牌
func (s *JWTService) generateToken(userID uint, role string, tokenType TokenType, expiry time.Duration) (string, error) {
	// 创建声明
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "p3-server",
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证 JWT 令牌
func (s *JWTService) ValidateToken(tokenString string) (*CustomClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	// 验证令牌有效性
	if !token.Valid {
		return nil, errors.New("无效的令牌")
	}

	// 提取声明
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("无效的令牌声明")
	}

	return claims, nil
}

// RefreshAccessToken 使用刷新令牌生成新的访问令牌
func (s *JWTService) RefreshAccessToken(refreshTokenString string) (string, error) {
	// 验证刷新令牌
	claims, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("验证刷新令牌失败: %w", err)
	}

	// 确保是刷新令牌
	if claims.Type != RefreshToken {
		return "", errors.New("无效的刷新令牌类型")
	}

	// 生成新的访问令牌
	accessToken, err := s.generateToken(claims.UserID, claims.Role, AccessToken, s.accessExpiry)
	if err != nil {
		return "", fmt.Errorf("生成新的访问令牌失败: %w", err)
	}

	return accessToken, nil
}

// BlacklistToken 将令牌加入黑名单
// 注意：实际实现应该使用 Redis 或数据库来存储黑名单
func (s *JWTService) BlacklistToken(tokenString string) error {
	// 解析令牌以获取过期时间
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("解析令牌失败: %w", err)
	}

	// 计算令牌剩余有效期
	expiresAt := claims.ExpiresAt.Time
	ttl := time.Until(expiresAt)

	// 将令牌添加到黑名单
	// 这里应该使用 Redis 或数据库实现
	logger.Info("令牌已加入黑名单，有效期: %v", ttl)

	return nil
}
