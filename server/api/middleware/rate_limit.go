package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/common/logger"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	// 每个 IP 的请求记录
	requests map[string][]time.Time
	// 时间窗口（秒）
	window time.Duration
	// 窗口内允许的最大请求数
	limit int
	// 互斥锁
	mu sync.Mutex
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(window time.Duration, limit int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		limit:    limit,
	}
}

// RateLimit 速率限制中间件
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端 IP
		ip := getClientIP(c.Request)

		// 检查是否超过速率限制
		if rl.isLimited(ip) {
			logger.Warn("IP %s 请求过于频繁，已被限制", ip)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// isLimited 检查 IP 是否超过速率限制
func (rl *RateLimiter) isLimited(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// 清理过期的请求记录
	if times, exists := rl.requests[ip]; exists {
		var validTimes []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				validTimes = append(validTimes, t)
			}
		}
		rl.requests[ip] = validTimes
	}

	// 获取当前窗口内的请求数
	count := len(rl.requests[ip])

	// 如果请求数超过限制，则拒绝请求
	if count >= rl.limit {
		return true
	}

	// 记录本次请求
	rl.requests[ip] = append(rl.requests[ip], now)
	return false
}

// getClientIP 获取客户端真实 IP
func getClientIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 头获取
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For 可能包含多个 IP，取第一个
		ips := splitAndTrim(xForwardedFor, ",")
		if len(ips) > 0 && ips[0] != "" {
			return ips[0]
		}
	}

	// 尝试从 X-Real-IP 头获取
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// 从 RemoteAddr 获取
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// splitAndTrim 分割字符串并去除空格
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range strings.Split(s, sep) {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

// 为不同的 API 路径设置不同的速率限制
var (
	// 全局限制器：每分钟 60 个请求
	GlobalLimiter = NewRateLimiter(time.Minute, 60)
	// 认证限制器：每分钟 10 个请求
	AuthLimiter = NewRateLimiter(time.Minute, 10)
	// 注册限制器：每小时 5 个请求
	RegisterLimiter = NewRateLimiter(time.Hour, 5)
)

// SetupRateLimits 设置速率限制
func SetupRateLimits(r *gin.Engine) {
	// 全局限制
	r.Use(GlobalLimiter.RateLimit())

	// 认证路由限制
	auth := r.Group("/api/v1/auth")
	{
		auth.Use(AuthLimiter.RateLimit())
		// 注册路由额外限制
		auth.POST("/register", RegisterLimiter.RateLimit())
	}
}
