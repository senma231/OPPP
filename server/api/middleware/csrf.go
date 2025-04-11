package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/common/logger"
)

const (
	// CSRFTokenLength CSRF 令牌长度
	CSRFTokenLength = 32
	// CSRFTokenCookieName CSRF 令牌 Cookie 名称
	CSRFTokenCookieName = "csrf_token"
	// CSRFTokenHeaderName CSRF 令牌请求头名称
	CSRFTokenHeaderName = "X-CSRF-Token"
	// CSRFTokenFormName CSRF 令牌表单字段名称
	CSRFTokenFormName = "csrf_token"
	// CSRFCookieMaxAge CSRF Cookie 最大有效期（秒）
	CSRFCookieMaxAge = 3600
)

var (
	// ErrInvalidCSRFToken 无效的 CSRF 令牌
	ErrInvalidCSRFToken = errors.New("无效的 CSRF 令牌")
	// ErrMissingCSRFToken 缺少 CSRF 令牌
	ErrMissingCSRFToken = errors.New("缺少 CSRF 令牌")
)

// CSRFProtection CSRF 保护中间件
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于安全的 HTTP 方法（GET, HEAD, OPTIONS, TRACE），不需要 CSRF 保护
		if isSafeMethod(c.Request.Method) {
			// 为安全方法生成 CSRF 令牌
			token, err := generateCSRFToken()
			if err != nil {
				logger.Error("生成 CSRF 令牌失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
				c.Abort()
				return
			}

			// 设置 CSRF 令牌 Cookie
			setCSRFCookie(c, token)
			
			// 将令牌存储在上下文中，以便视图可以访问
			c.Set(CSRFTokenFormName, token)
			
			c.Next()
			return
		}

		// 对于不安全的方法（POST, PUT, DELETE, PATCH），需要验证 CSRF 令牌
		// 从 Cookie 中获取令牌
		cookieToken, err := c.Cookie(CSRFTokenCookieName)
		if err != nil || cookieToken == "" {
			logger.Warn("缺少 CSRF Cookie 令牌")
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF 验证失败"})
			c.Abort()
			return
		}

		// 从请求中获取令牌（优先从请求头获取，然后从表单获取）
		requestToken := c.GetHeader(CSRFTokenHeaderName)
		if requestToken == "" {
			requestToken = c.PostForm(CSRFTokenFormName)
		}

		if requestToken == "" {
			logger.Warn("缺少 CSRF 请求令牌")
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF 验证失败"})
			c.Abort()
			return
		}

		// 验证令牌
		if requestToken != cookieToken {
			logger.Warn("CSRF 令牌不匹配: cookie=%s, request=%s", cookieToken, requestToken)
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF 验证失败"})
			c.Abort()
			return
		}

		// 生成新的 CSRF 令牌
		newToken, err := generateCSRFToken()
		if err != nil {
			logger.Error("生成 CSRF 令牌失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
			c.Abort()
			return
		}

		// 设置新的 CSRF 令牌 Cookie
		setCSRFCookie(c, newToken)
		
		// 将新令牌存储在上下文中
		c.Set(CSRFTokenFormName, newToken)

		c.Next()
	}
}

// generateCSRFToken 生成 CSRF 令牌
func generateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// setCSRFCookie 设置 CSRF Cookie
func setCSRFCookie(c *gin.Context, token string) {
	c.SetCookie(
		CSRFTokenCookieName,
		token,
		CSRFCookieMaxAge,
		"/",
		"",
		c.Request.TLS != nil, // 如果是 HTTPS，则设置 Secure
		true,                 // HttpOnly
	)
}

// isSafeMethod 检查 HTTP 方法是否安全
func isSafeMethod(method string) bool {
	return method == http.MethodGet ||
		method == http.MethodHead ||
		method == http.MethodOptions ||
		method == http.MethodTrace
}

// GetCSRFToken 从上下文中获取 CSRF 令牌
func GetCSRFToken(c *gin.Context) string {
	token, exists := c.Get(CSRFTokenFormName)
	if !exists {
		return ""
	}
	return token.(string)
}
