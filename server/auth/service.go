package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/senma231/p3/common/errors"
	"github.com/senma231/p3/common/logger"
	"github.com/senma231/p3/server/config"
	"github.com/senma231/p3/server/db"
	"gorm.io/gorm"
)

// Service 认证服务
type Service struct {
	cfg        *config.Config
	jwtService *JWTService
}

// NewService 创建认证服务
func NewService(cfg *config.Config) *Service {
	return &Service{
		cfg:        cfg,
		jwtService: NewJWTService(cfg.JWT.Secret),
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=100"`
	Email    string `json:"email" binding:"required,email"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	TOTPCode string `json:"totp_code"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Register 注册用户
func (s *Service) Register(req *RegisterRequest) (*db.User, error) {
	// 检查用户名是否已存在
	var existingUser db.User
	if result := db.DB.Where("username = ?", req.Username).First(&existingUser); result.Error == nil {
		return nil, errors.Conflict("用户名已存在")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.Database("查询用户失败", result.Error)
	}

	// 检查邮箱是否已存在
	if result := db.DB.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		return nil, errors.Conflict("邮箱已存在")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.Database("查询用户失败", result.Error)
	}

	// 哈希密码
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, errors.Internal("密码哈希失败")
	}

	// 创建用户
	user := &db.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
	}

	if result := db.DB.Create(user); result.Error != nil {
		return nil, errors.Database("创建用户失败", result.Error)
	}

	return user, nil
}

// Login 用户登录
func (s *Service) Login(req *LoginRequest, userAgent, ip string) (*TokenResponse, error) {
	// 查找用户
	var user db.User
	if result := db.DB.Where("username = ?", req.Username).First(&user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Unauthorized("用户名或密码错误")
		}
		return nil, errors.Database("查询用户失败", result.Error)
	}

	// 验证密码
	if !VerifyPassword(req.Password, user.Password) {
		return nil, errors.Unauthorized("用户名或密码错误")
	}

	// 检查是否启用了双因素认证
	var totp db.TOTP
	if result := db.DB.Where("user_id = ? AND enabled = ?", user.ID, true).First(&totp); result.Error == nil {
		// 如果启用了双因素认证，验证 TOTP 代码
		if req.TOTPCode == "" {
			return nil, errors.Unauthorized("需要双因素认证代码")
		}

		// 验证 TOTP 代码
		valid, err := VerifyTOTP(totp.Secret, req.TOTPCode)
		if err != nil || !valid {
			return nil, errors.Unauthorized("双因素认证代码无效")
		}

		// 更新最后使用时间
		totp.LastUsedAt = time.Now()
		db.DB.Save(&totp)
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.Database("查询 TOTP 失败", result.Error)
	}

	// 生成令牌
	accessToken, refreshToken, err := s.jwtService.GenerateTokens(user.ID, "user")
	if err != nil {
		return nil, errors.Internal("生成令牌失败")
	}

	// 创建会话
	session := &db.Session{
		UserID:       user.ID,
		Token:        accessToken,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IP:           ip,
		ExpiresAt:    time.Now().Add(time.Hour * time.Duration(s.cfg.JWT.AccessExpireTime)),
		LastActiveAt: time.Now(),
	}

	if result := db.DB.Create(session); result.Error != nil {
		return nil, errors.Database("创建会话失败", result.Error)
	}

	// 更新用户最后登录时间
	user.LastLoginAt = time.Now()
	if result := db.DB.Save(&user); result.Error != nil {
		logger.Warn("更新用户最后登录时间失败: %v", result.Error)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.JWT.AccessExpireTime * 3600),
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken 刷新令牌
func (s *Service) RefreshToken(req *RefreshTokenRequest) (*TokenResponse, error) {
	// 验证刷新令牌
	claims, err := s.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, errors.Unauthorized("无效的刷新令牌")
	}

	// 确保是刷新令牌
	if claims.Type != RefreshToken {
		return nil, errors.Unauthorized("无效的令牌类型")
	}

	// 查找会话
	var session db.Session
	if result := db.DB.Where("refresh_token = ? AND revoked = ?", req.RefreshToken, false).First(&session); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Unauthorized("会话不存在或已被撤销")
		}
		return nil, errors.Database("查询会话失败", result.Error)
	}

	// 检查会话是否过期
	if session.ExpiresAt.Before(time.Now()) {
		return nil, errors.Unauthorized("会话已过期")
	}

	// 生成新的访问令牌
	accessToken, err := s.jwtService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		return nil, errors.Internal("生成访问令牌失败")
	}

	// 更新会话
	session.Token = accessToken
	session.LastActiveAt = time.Now()
	if result := db.DB.Save(&session); result.Error != nil {
		return nil, errors.Database("更新会话失败", result.Error)
	}

	return &TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(s.cfg.JWT.AccessExpireTime * 3600),
		TokenType:   "Bearer",
	}, nil
}

// Logout 用户登出
func (s *Service) Logout(token string) error {
	// 查找会话
	var session db.Session
	if result := db.DB.Where("token = ? AND revoked = ?", token, false).First(&session); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil // 会话不存在，视为已登出
		}
		return errors.Database("查询会话失败", result.Error)
	}

	// 撤销会话
	session.Revoked = true
	if result := db.DB.Save(&session); result.Error != nil {
		return errors.Database("撤销会话失败", result.Error)
	}

	// 将令牌加入黑名单
	if err := s.jwtService.BlacklistToken(session.Token); err != nil {
		logger.Warn("将令牌加入黑名单失败: %v", err)
	}

	return nil
}

// GetUserByID 根据 ID 获取用户
func (s *Service) GetUserByID(id uint) (*db.User, error) {
	var user db.User
	if result := db.DB.First(&user, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("用户不存在")
		}
		return nil, errors.Database("查询用户失败", result.Error)
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (s *Service) UpdateUser(id uint, email string) (*db.User, error) {
	var user db.User
	if result := db.DB.First(&user, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("用户不存在")
		}
		return nil, errors.Database("查询用户失败", result.Error)
	}

	// 更新邮箱
	if email != "" && email != user.Email {
		// 检查邮箱是否已存在
		var existingUser db.User
		if result := db.DB.Where("email = ? AND id != ?", email, id).First(&existingUser); result.Error == nil {
			return nil, errors.Conflict("邮箱已存在")
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Database("查询用户失败", result.Error)
		}

		user.Email = email
	}

	if result := db.DB.Save(&user); result.Error != nil {
		return nil, errors.Database("更新用户失败", result.Error)
	}

	return &user, nil
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(id uint, oldPassword, newPassword string) error {
	var user db.User
	if result := db.DB.First(&user, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.NotFound("用户不存在")
		}
		return errors.Database("查询用户失败", result.Error)
	}

	// 验证旧密码
	if !VerifyPassword(oldPassword, user.Password) {
		return errors.Unauthorized("旧密码错误")
	}

	// 哈希新密码
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return errors.Internal("密码哈希失败")
	}

	// 更新密码
	user.Password = hashedPassword
	if result := db.DB.Save(&user); result.Error != nil {
		return errors.Database("更新密码失败", result.Error)
	}

	// 撤销所有会话
	if result := db.DB.Model(&db.Session{}).Where("user_id = ?", id).Update("revoked", true); result.Error != nil {
		logger.Warn("撤销会话失败: %v", result.Error)
	}

	return nil
}

// EnableTOTP 启用双因素认证
func (s *Service) EnableTOTP(userID uint) (string, string, error) {
	var user db.User
	if result := db.DB.First(&user, userID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", "", errors.NotFound("用户不存在")
		}
		return "", "", errors.Database("查询用户失败", result.Error)
	}

	// 检查是否已启用
	var totp db.TOTP
	if result := db.DB.Where("user_id = ?", userID).First(&totp); result.Error == nil {
		if totp.Enabled {
			return "", "", errors.Conflict("双因素认证已启用")
		}
		// 如果存在但未启用，则重新生成
		db.DB.Delete(&totp)
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", "", errors.Database("查询 TOTP 失败", result.Error)
	}

	// 生成 TOTP 密钥
	secret, uri, err := GenerateTOTPSecret(user.Username, DefaultTOTPConfig)
	if err != nil {
		return "", "", errors.Internal("生成 TOTP 密钥失败")
	}

	// 创建 TOTP 记录
	totp = db.TOTP{
		UserID:   userID,
		Secret:   secret,
		Enabled:  false,
		Verified: false,
	}

	if result := db.DB.Create(&totp); result.Error != nil {
		return "", "", errors.Database("创建 TOTP 记录失败", result.Error)
	}

	return secret, uri, nil
}

// VerifyAndEnableTOTP 验证并启用双因素认证
func (s *Service) VerifyAndEnableTOTP(userID uint, code string) error {
	var totp db.TOTP
	if result := db.DB.Where("user_id = ?", userID).First(&totp); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.NotFound("未找到 TOTP 记录")
		}
		return errors.Database("查询 TOTP 失败", result.Error)
	}

	// 验证 TOTP 代码
	valid, err := VerifyTOTP(totp.Secret, code)
	if err != nil || !valid {
		return errors.Unauthorized("TOTP 代码无效")
	}

	// 启用 TOTP
	totp.Enabled = true
	totp.Verified = true
	totp.LastUsedAt = time.Now()

	if result := db.DB.Save(&totp); result.Error != nil {
		return errors.Database("更新 TOTP 记录失败", result.Error)
	}

	return nil
}

// DisableTOTP 禁用双因素认证
func (s *Service) DisableTOTP(userID uint, code string) error {
	var totp db.TOTP
	if result := db.DB.Where("user_id = ? AND enabled = ?", userID, true).First(&totp); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.NotFound("未找到已启用的 TOTP 记录")
		}
		return errors.Database("查询 TOTP 失败", result.Error)
	}

	// 验证 TOTP 代码
	valid, err := VerifyTOTP(totp.Secret, code)
	if err != nil || !valid {
		return errors.Unauthorized("TOTP 代码无效")
	}

	// 删除 TOTP 记录
	if result := db.DB.Delete(&totp); result.Error != nil {
		return errors.Database("删除 TOTP 记录失败", result.Error)
	}

	return nil
}

// VerifyPassword 验证密码
func VerifyPassword(password, hashedPassword string) bool {
	valid, err := auth.VerifyPassword(password, hashedPassword)
	if err != nil {
		logger.Error("验证密码失败: %v", err)
		return false
	}
	return valid
}

// HashPassword 哈希密码
func HashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}

// GetUserFromRequest 从请求中获取用户
func (s *Service) GetUserFromRequest(r *http.Request) (*db.User, error) {
	// 从请求头获取令牌
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.Unauthorized("未提供认证令牌")
	}

	// 解析令牌
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		return nil, errors.Unauthorized("无效的认证头")
	}

	// 验证令牌
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		return nil, errors.Unauthorized("无效的认证令牌")
	}

	// 确保是访问令牌
	if claims.Type != AccessToken {
		return nil, errors.Unauthorized("无效的令牌类型")
	}

	// 查找会话
	var session db.Session
	if result := db.DB.Where("token = ? AND revoked = ?", tokenString, false).First(&session); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Unauthorized("会话不存在或已被撤销")
		}
		return nil, errors.Database("查询会话失败", result.Error)
	}

	// 检查会话是否过期
	if session.ExpiresAt.Before(time.Now()) {
		return nil, errors.Unauthorized("会话已过期")
	}

	// 更新会话最后活动时间
	session.LastActiveAt = time.Now()
	if result := db.DB.Save(&session); result.Error != nil {
		logger.Warn("更新会话最后活动时间失败: %v", result.Error)
	}

	// 获取用户
	return s.GetUserByID(claims.UserID)
}
