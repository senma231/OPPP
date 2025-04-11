package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestErrorCreation(t *testing.T) {
	// 测试 New 函数
	err := New(ErrInvalidParam, "无效参数")
	if err.Code != ErrInvalidParam {
		t.Errorf("错误码错误，期望 %d，实际 %d", ErrInvalidParam, err.Code)
	}
	if err.Message != "无效参数" {
		t.Errorf("错误消息错误，期望 '无效参数'，实际 '%s'", err.Message)
	}
	if err.Cause != nil {
		t.Errorf("错误原因应为 nil，实际 %v", err.Cause)
	}

	// 测试 Wrap 函数
	cause := errors.New("原始错误")
	wrappedErr := Wrap(ErrDatabase, "数据库错误", cause)
	if wrappedErr.Code != ErrDatabase {
		t.Errorf("包装错误码错误，期望 %d，实际 %d", ErrDatabase, wrappedErr.Code)
	}
	if wrappedErr.Message != "数据库错误" {
		t.Errorf("包装错误消息错误，期望 '数据库错误'，实际 '%s'", wrappedErr.Message)
	}
	if wrappedErr.Cause != cause {
		t.Errorf("包装错误原因错误，期望 %v，实际 %v", cause, wrappedErr.Cause)
	}
}

func TestErrorString(t *testing.T) {
	// 测试无原因错误的字符串表示
	err := New(ErrNotFound, "资源不存在")
	if err.Error() != "资源不存在" {
		t.Errorf("错误字符串错误，期望 '资源不存在'，实际 '%s'", err.Error())
	}

	// 测试有原因错误的字符串表示
	cause := errors.New("原始错误")
	wrappedErr := Wrap(ErrDatabase, "数据库错误", cause)
	expected := "数据库错误: 原始错误"
	if wrappedErr.Error() != expected {
		t.Errorf("包装错误字符串错误，期望 '%s'，实际 '%s'", expected, wrappedErr.Error())
	}
}

func TestErrorUnwrap(t *testing.T) {
	// 测试 Unwrap 函数
	cause := errors.New("原始错误")
	wrappedErr := Wrap(ErrDatabase, "数据库错误", cause)
	unwrapped := wrappedErr.Unwrap()
	if unwrapped != cause {
		t.Errorf("解包错误错误，期望 %v，实际 %v", cause, unwrapped)
	}
}

func TestErrorStatusCode(t *testing.T) {
	// 测试不同错误码的 HTTP 状态码
	testCases := []struct {
		code           ErrorCode
		expectedStatus int
	}{
		{ErrInvalidParam, http.StatusBadRequest},
		{ErrUnauthorized, http.StatusUnauthorized},
		{ErrForbidden, http.StatusForbidden},
		{ErrNotFound, http.StatusNotFound},
		{ErrUserNotFound, http.StatusNotFound},
		{ErrDeviceNotFound, http.StatusNotFound},
		{ErrAppNotFound, http.StatusNotFound},
		{ErrForwardNotFound, http.StatusNotFound},
		{ErrPeerNotFound, http.StatusNotFound},
		{ErrConflict, http.StatusConflict},
		{ErrUserAlreadyExists, http.StatusConflict},
		{ErrDeviceAlreadyExists, http.StatusConflict},
		{ErrAppAlreadyExists, http.StatusConflict},
		{ErrForwardAlreadyExists, http.StatusConflict},
		{ErrPortInUse, http.StatusConflict},
		{ErrTooManyRequests, http.StatusTooManyRequests},
		{ErrNotImplemented, http.StatusNotImplemented},
		{ErrServiceUnavailable, http.StatusServiceUnavailable},
		{ErrBadGateway, http.StatusBadGateway},
		{ErrGatewayTimeout, http.StatusGatewayTimeout},
		{ErrUnknown, http.StatusInternalServerError},
		{ErrInternal, http.StatusInternalServerError},
		{ErrDatabase, http.StatusInternalServerError},
		{ErrNetwork, http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		err := New(tc.code, "测试错误")
		status := err.StatusCode()
		if status != tc.expectedStatus {
			t.Errorf("错误码 %d 的 HTTP 状态码错误，期望 %d，实际 %d", tc.code, tc.expectedStatus, status)
		}
	}
}

func TestErrorIs(t *testing.T) {
	// 测试 Is 函数
	err := New(ErrNotFound, "资源不存在")
	if !Is(err, ErrNotFound) {
		t.Error("Is 函数应该返回 true")
	}
	if Is(err, ErrInvalidParam) {
		t.Error("Is 函数应该返回 false")
	}
	if Is(nil, ErrNotFound) {
		t.Error("Is 函数对 nil 错误应该返回 false")
	}
	if Is(errors.New("普通错误"), ErrNotFound) {
		t.Error("Is 函数对非 Error 类型应该返回 false")
	}
}

func TestErrorAsError(t *testing.T) {
	// 测试 AsError 函数
	err := New(ErrNotFound, "资源不存在")
	asErr := AsError(err)
	if asErr != err {
		t.Errorf("AsError 函数对 Error 类型错误，期望 %v，实际 %v", err, asErr)
	}

	stdErr := errors.New("普通错误")
	asStdErr := AsError(stdErr)
	if asStdErr.Code != ErrUnknown {
		t.Errorf("AsError 函数对普通错误，错误码期望 %d，实际 %d", ErrUnknown, asStdErr.Code)
	}
	if asStdErr.Message != stdErr.Error() {
		t.Errorf("AsError 函数对普通错误，错误消息期望 '%s'，实际 '%s'", stdErr.Error(), asStdErr.Message)
	}
	if asStdErr.Cause != stdErr {
		t.Errorf("AsError 函数对普通错误，错误原因期望 %v，实际 %v", stdErr, asStdErr.Cause)
	}

	if AsError(nil) != nil {
		t.Error("AsError 函数对 nil 错误应该返回 nil")
	}
}

func TestErrorHelperFunctions(t *testing.T) {
	// 测试辅助函数
	testCases := []struct {
		fn       func(string) *Error
		code     ErrorCode
		message  string
		expected string
	}{
		{InvalidParam, ErrInvalidParam, "无效参数", "无效参数"},
		{Unauthorized, ErrUnauthorized, "未授权", "未授权"},
		{Forbidden, ErrForbidden, "禁止访问", "禁止访问"},
		{NotFound, ErrNotFound, "资源不存在", "资源不存在"},
		{Conflict, ErrConflict, "资源冲突", "资源冲突"},
		{Internal, ErrInternal, "内部错误", "内部错误"},
		{Timeout, ErrTimeout, "请求超时", "请求超时"},
		{NotImplemented, ErrNotImplemented, "未实现", "未实现"},
		{ServiceUnavailable, ErrServiceUnavailable, "服务不可用", "服务不可用"},
		{TooManyRequests, ErrTooManyRequests, "请求过多", "请求过多"},
		{BadGateway, ErrBadGateway, "网关错误", "网关错误"},
		{GatewayTimeout, ErrGatewayTimeout, "网关超时", "网关超时"},
	}

	for _, tc := range testCases {
		err := tc.fn(tc.message)
		if err.Code != tc.code {
			t.Errorf("%T 函数错误码错误，期望 %d，实际 %d", tc.fn, tc.code, err.Code)
		}
		if err.Message != tc.expected {
			t.Errorf("%T 函数错误消息错误，期望 '%s'，实际 '%s'", tc.fn, tc.expected, err.Message)
		}
	}

	// 测试带原因的辅助函数
	cause := errors.New("原始错误")
	dbErr := Database("数据库错误", cause)
	if dbErr.Code != ErrDatabase {
		t.Errorf("Database 函数错误码错误，期望 %d，实际 %d", ErrDatabase, dbErr.Code)
	}
	if dbErr.Message != "数据库错误" {
		t.Errorf("Database 函数错误消息错误，期望 '数据库错误'，实际 '%s'", dbErr.Message)
	}
	if dbErr.Cause != cause {
		t.Errorf("Database 函数错误原因错误，期望 %v，实际 %v", cause, dbErr.Cause)
	}

	netErr := Network("网络错误", cause)
	if netErr.Code != ErrNetwork {
		t.Errorf("Network 函数错误码错误，期望 %d，实际 %d", ErrNetwork, netErr.Code)
	}
	if netErr.Message != "网络错误" {
		t.Errorf("Network 函数错误消息错误，期望 '网络错误'，实际 '%s'", netErr.Message)
	}
	if netErr.Cause != cause {
		t.Errorf("Network 函数错误原因错误，期望 %v，实际 %v", cause, netErr.Cause)
	}
}
