package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码
type ErrorCode int

const (
	// ErrUnknown 未知错误
	ErrUnknown ErrorCode = iota + 1000
	// ErrInvalidParam 无效参数
	ErrInvalidParam
	// ErrUnauthorized 未授权
	ErrUnauthorized
	// ErrForbidden 禁止访问
	ErrForbidden
	// ErrNotFound 未找到
	ErrNotFound
	// ErrConflict 冲突
	ErrConflict
	// ErrInternal 内部错误
	ErrInternal
	// ErrDatabase 数据库错误
	ErrDatabase
	// ErrNetwork 网络错误
	ErrNetwork
	// ErrTimeout 超时
	ErrTimeout
	// ErrNotImplemented 未实现
	ErrNotImplemented
	// ErrServiceUnavailable 服务不可用
	ErrServiceUnavailable
	// ErrTooManyRequests 请求过多
	ErrTooManyRequests
	// ErrBadGateway 网关错误
	ErrBadGateway
	// ErrGatewayTimeout 网关超时
	ErrGatewayTimeout
	// ErrInvalidToken 无效令牌
	ErrInvalidToken
	// ErrTokenExpired 令牌过期
	ErrTokenExpired
	// ErrUserNotFound 用户不存在
	ErrUserNotFound
	// ErrUserAlreadyExists 用户已存在
	ErrUserAlreadyExists
	// ErrInvalidPassword 密码错误
	ErrInvalidPassword
	// ErrDeviceNotFound 设备不存在
	ErrDeviceNotFound
	// ErrDeviceAlreadyExists 设备已存在
	ErrDeviceAlreadyExists
	// ErrDeviceOffline 设备离线
	ErrDeviceOffline
	// ErrAppNotFound 应用不存在
	ErrAppNotFound
	// ErrAppAlreadyExists 应用已存在
	ErrAppAlreadyExists
	// ErrAppNotRunning 应用未运行
	ErrAppNotRunning
	// ErrAppAlreadyRunning 应用已运行
	ErrAppAlreadyRunning
	// ErrForwardNotFound 转发规则不存在
	ErrForwardNotFound
	// ErrForwardAlreadyExists 转发规则已存在
	ErrForwardAlreadyExists
	// ErrForwardNotEnabled 转发规则未启用
	ErrForwardNotEnabled
	// ErrForwardAlreadyEnabled 转发规则已启用
	ErrForwardAlreadyEnabled
	// ErrPortInUse 端口已被占用
	ErrPortInUse
	// ErrConnectionFailed 连接失败
	ErrConnectionFailed
	// ErrPeerNotFound 对等节点不存在
	ErrPeerNotFound
	// ErrPeerOffline 对等节点离线
	ErrPeerOffline
	// ErrNATTraversalFailed NAT 穿透失败
	ErrNATTraversalFailed
	// ErrRelayFailed 中继失败
	ErrRelayFailed
	// ErrTURNFailed TURN 失败
	ErrTURNFailed
	// ErrSTUNFailed STUN 失败
	ErrSTUNFailed
	// ErrUPnPFailed UPnP 失败
	ErrUPnPFailed
	// ErrNATPMPFailed NAT-PMP 失败
	ErrNATPMPFailed
	// ErrEncryptionFailed 加密失败
	ErrEncryptionFailed
	// ErrDecryptionFailed 解密失败
	ErrDecryptionFailed
	// ErrAuthenticationFailed 认证失败
	ErrAuthenticationFailed
	// ErrAuthorizationFailed 授权失败
	ErrAuthorizationFailed
)

// Error 错误
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Cause   error     `json:"cause,omitempty"`
}

// Error 返回错误信息
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap 解包错误
func (e *Error) Unwrap() error {
	return e.Cause
}

// StatusCode 返回 HTTP 状态码
func (e *Error) StatusCode() int {
	switch e.Code {
	case ErrInvalidParam:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound, ErrUserNotFound, ErrDeviceNotFound, ErrAppNotFound, ErrForwardNotFound, ErrPeerNotFound:
		return http.StatusNotFound
	case ErrConflict, ErrUserAlreadyExists, ErrDeviceAlreadyExists, ErrAppAlreadyExists, ErrForwardAlreadyExists, ErrPortInUse:
		return http.StatusConflict
	case ErrTooManyRequests:
		return http.StatusTooManyRequests
	case ErrNotImplemented:
		return http.StatusNotImplemented
	case ErrServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrBadGateway:
		return http.StatusBadGateway
	case ErrGatewayTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// New 创建错误
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装错误
func Wrap(code ErrorCode, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Is 检查错误类型
func Is(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// AsError 将错误转换为 Error
func AsError(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return Wrap(ErrUnknown, err.Error(), err)
}

// InvalidParam 创建无效参数错误
func InvalidParam(message string) *Error {
	return New(ErrInvalidParam, message)
}

// Unauthorized 创建未授权错误
func Unauthorized(message string) *Error {
	return New(ErrUnauthorized, message)
}

// Forbidden 创建禁止访问错误
func Forbidden(message string) *Error {
	return New(ErrForbidden, message)
}

// NotFound 创建未找到错误
func NotFound(message string) *Error {
	return New(ErrNotFound, message)
}

// Conflict 创建冲突错误
func Conflict(message string) *Error {
	return New(ErrConflict, message)
}

// Internal 创建内部错误
func Internal(message string) *Error {
	return New(ErrInternal, message)
}

// Database 创建数据库错误
func Database(message string, cause error) *Error {
	return Wrap(ErrDatabase, message, cause)
}

// Network 创建网络错误
func Network(message string, cause error) *Error {
	return Wrap(ErrNetwork, message, cause)
}

// Timeout 创建超时错误
func Timeout(message string) *Error {
	return New(ErrTimeout, message)
}

// NotImplemented 创建未实现错误
func NotImplemented(message string) *Error {
	return New(ErrNotImplemented, message)
}

// ServiceUnavailable 创建服务不可用错误
func ServiceUnavailable(message string) *Error {
	return New(ErrServiceUnavailable, message)
}

// TooManyRequests 创建请求过多错误
func TooManyRequests(message string) *Error {
	return New(ErrTooManyRequests, message)
}

// BadGateway 创建网关错误
func BadGateway(message string) *Error {
	return New(ErrBadGateway, message)
}

// GatewayTimeout 创建网关超时错误
func GatewayTimeout(message string) *Error {
	return New(ErrGatewayTimeout, message)
}
