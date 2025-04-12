package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/common/errors"
	"github.com/senma231/p3/common/logger"
	"github.com/senma231/p3/server/auth"
	"github.com/senma231/p3/server/device"
)

// Auth 认证中间件
func Auth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 解析令牌
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证头",
			})
			c.Abort()
			return
		}

		// 获取用户
		user, err := authService.GetUserFromRequest(c.Request)
		if err != nil {
			errObj := errors.AsError(err)
			c.JSON(errObj.StatusCode(), gin.H{
				"error": errObj.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中
		c.Set("user", user)
		c.Set("userID", user.ID)

		c.Next()
	}
}

// DeviceAuth 设备认证中间件
func DeviceAuth(deviceService *device.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取节点 ID 和令牌
		nodeID := c.GetHeader("X-Node-ID")
		token := c.GetHeader("X-Node-Token")

		if nodeID == "" || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未提供节点 ID 或令牌",
			})
			c.Abort()
			return
		}

		// 认证设备
		device, err := deviceService.AuthenticateDevice(nodeID, token)
		if err != nil {
			errObj := errors.AsError(err)
			c.JSON(errObj.StatusCode(), gin.H{
				"error": errObj.Error(),
			})
			c.Abort()
			return
		}

		// 将设备信息存储在上下文中
		c.Set("device", device)
		c.Set("deviceID", device.ID)
		c.Set("userID", device.UserID)

		c.Next()
	}
}

// CORS CORS 中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		// 执行时间
		latency := end.Sub(start)

		// 请求方法
		method := c.Request.Method
		// 请求路由
		path := c.Request.URL.Path
		// 状态码
		statusCode := c.Writer.Status()
		// 客户端 IP
		clientIP := c.ClientIP()

		// 日志格式
		logger.Info("[GIN] %v | %3d | %13v | %15s | %-7s %s",
			end.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

// RateLimit 速率限制中间件
func RateLimit() gin.HandlerFunc {
	// 创建速率限制器
	limiter := NewRateLimiter(time.Minute, 60)
	return limiter.RateLimit()
}

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

// GetCSRFToken 从上下文中获取 CSRF 令牌
func GetCSRFToken(c *gin.Context) string {
	token, exists := c.Get(CSRFTokenFormName)
	if !exists {
		return ""
	}
	return token.(string)
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
