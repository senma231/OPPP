package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/senma231/p3/server/config"
	"github.com/senma231/p3/server/db"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Claims JWT 声明
type Claims struct {
	UserID uint   `json:"userId"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// Service 认证服务
type Service struct {
	config *config.Config
}

// NewService 创建认证服务
func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
	}
}

// Register 注册用户
func (s *Service) Register(username, password, email string) (*db.User, error) {
	// 检查用户名是否已存在
	var existingUser db.User
	if err := db.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查邮箱是否已存在
	if email != "" {
		if err := db.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
			return nil, errors.New("邮箱已存在")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("查询用户失败: %w", err)
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("加密密码失败: %w", err)
	}

	// 创建用户
	user := &db.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	}

	if err := db.DB.Create(user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// Login 用户登录
func (s *Service) Login(username, password string) (*db.User, string, error) {
	// 查询用户
	var user db.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("用户不存在")
		}
		return nil, "", fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", errors.New("密码错误")
	}

	// 更新最后登录时间
	user.LastLoginAt = time.Now()
	if err := db.DB.Save(&user).Error; err != nil {
		return nil, "", fmt.Errorf("更新用户失败: %w", err)
	}

	// 生成 JWT Token
	token, err := s.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("生成 Token 失败: %w", err)
	}

	return &user, token, nil
}

// GenerateToken 生成 JWT Token
func (s *Service) GenerateToken(userID uint, username string) (string, error) {
	// 设置过期时间
	expireTime := time.Now().Add(time.Duration(s.config.JWT.ExpireTime) * time.Hour)

	// 创建声明
	claims := &Claims{
		UserID:   userID,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "p3-server",
		},
	}

	// 创建 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名 Token
	tokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析 JWT Token
func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	// 解析 Token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	// 验证 Token
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的 Token")
}

// VerifyPassword 验证密码
func (s *Service) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// HashPassword 加密密码
func (s *Service) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// GetUserByID 根据 ID 获取用户
func (s *Service) GetUserByID(userID uint) (*db.User, error) {
	var user db.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}
