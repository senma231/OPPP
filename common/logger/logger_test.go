package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLogLevels(t *testing.T) {
	// 创建缓冲区用于捕获日志输出
	var buf bytes.Buffer

	// 创建日志记录器
	logger := NewLogger(InfoLevel, &buf)

	// 测试不同级别的日志
	logger.Debug("这是一条调试日志")
	if buf.Len() > 0 {
		t.Error("调试日志不应该被记录")
	}
	buf.Reset()

	logger.Info("这是一条信息日志")
	if !strings.Contains(buf.String(), "INFO") || !strings.Contains(buf.String(), "这是一条信息日志") {
		t.Errorf("信息日志记录错误: %s", buf.String())
	}
	buf.Reset()

	logger.Warn("这是一条警告日志")
	if !strings.Contains(buf.String(), "WARN") || !strings.Contains(buf.String(), "这是一条警告日志") {
		t.Errorf("警告日志记录错误: %s", buf.String())
	}
	buf.Reset()

	logger.Error("这是一条错误日志")
	if !strings.Contains(buf.String(), "ERROR") || !strings.Contains(buf.String(), "这是一条错误日志") {
		t.Errorf("错误日志记录错误: %s", buf.String())
	}
	buf.Reset()

	// 测试设置日志级别
	logger.SetLevel(ErrorLevel)
	logger.Info("这条信息日志不应该被记录")
	logger.Warn("这条警告日志不应该被记录")
	if buf.Len() > 0 {
		t.Errorf("信息和警告日志不应该被记录: %s", buf.String())
	}
	buf.Reset()

	logger.Error("这是一条错误日志")
	if !strings.Contains(buf.String(), "ERROR") || !strings.Contains(buf.String(), "这是一条错误日志") {
		t.Errorf("错误日志记录错误: %s", buf.String())
	}
}

func TestLogFormat(t *testing.T) {
	// 创建缓冲区用于捕获日志输出
	var buf bytes.Buffer

	// 创建日志记录器
	logger := NewLogger(InfoLevel, &buf)

	// 测试日志格式
	logger.Info("测试日志格式")
	logOutput := buf.String()

	// 检查日期时间格式
	if !strings.Contains(logOutput, "-") || !strings.Contains(logOutput, ":") {
		t.Errorf("日志缺少日期时间: %s", logOutput)
	}

	// 检查日志级别
	if !strings.Contains(logOutput, "[INFO]") {
		t.Errorf("日志缺少级别: %s", logOutput)
	}

	// 检查文件名和行号
	if !strings.Contains(logOutput, "logger_test.go") {
		t.Errorf("日志缺少文件名: %s", logOutput)
	}

	// 检查日志消息
	if !strings.Contains(logOutput, "测试日志格式") {
		t.Errorf("日志缺少消息: %s", logOutput)
	}
}

func TestLogPrefix(t *testing.T) {
	// 创建缓冲区用于捕获日志输出
	var buf bytes.Buffer

	// 创建日志记录器
	logger := NewLogger(InfoLevel, &buf)

	// 设置前缀
	logger.SetPrefix("[测试] ")

	// 记录日志
	logger.Info("带前缀的日志")
	logOutput := buf.String()

	// 检查前缀
	if !strings.Contains(logOutput, "[测试] 带前缀的日志") {
		t.Errorf("日志缺少前缀: %s", logOutput)
	}
}

func TestLogOutput(t *testing.T) {
	// 创建临时文件
	tmpfile, err := os.CreateTemp("", "logger_test")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// 创建日志记录器
	logger := NewLogger(InfoLevel, tmpfile)

	// 记录日志
	logger.Info("写入文件的日志")

	// 关闭文件
	tmpfile.Close()

	// 读取文件内容
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("读取临时文件失败: %v", err)
	}

	// 检查日志内容
	if !strings.Contains(string(content), "写入文件的日志") {
		t.Errorf("日志未正确写入文件: %s", string(content))
	}
}

func TestParseLevel(t *testing.T) {
	testCases := []struct {
		input    string
		expected Level
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"info", InfoLevel},
		{"INFO", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"WARN", WarnLevel},
		{"error", ErrorLevel},
		{"ERROR", ErrorLevel},
		{"invalid", InfoLevel}, // 默认为 InfoLevel
	}

	for _, tc := range testCases {
		level := ParseLevel(tc.input)
		if level != tc.expected {
			t.Errorf("解析日志级别错误，输入 %s，期望 %v，实际 %v", tc.input, tc.expected, level)
		}
	}
}
