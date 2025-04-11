package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	// DebugLevel 调试级别
	DebugLevel Level = iota
	// InfoLevel 信息级别
	InfoLevel
	// WarnLevel 警告级别
	WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel
)

// String 返回日志级别的字符串表示
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel 解析日志级别
func ParseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// Logger 日志记录器
type Logger struct {
	level     Level
	output    io.Writer
	mu        sync.Mutex
	prefix    string
	callDepth int
}

var (
	// DefaultLogger 默认日志记录器
	DefaultLogger = NewLogger(InfoLevel, os.Stdout)
)

// NewLogger 创建日志记录器
func NewLogger(level Level, output io.Writer) *Logger {
	return &Logger{
		level:     level,
		output:    output,
		callDepth: 2,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput 设置输出
func (l *Logger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

// SetPrefix 设置前缀
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// SetCallDepth 设置调用深度
func (l *Logger) SetCallDepth(depth int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.callDepth = depth
}

// log 记录日志
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().Format("2006-01-02 15:04:05.000")
	var file string
	var line int
	var ok bool

	_, file, line, ok = runtime.Caller(l.callDepth)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}

	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	fmt.Fprintf(l.output, "%s [%s] %s:%d %s%s\n", now, level.String(), file, line, l.prefix, msg)
}

// Debug 记录调试级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

// Info 记录信息级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

// Warn 记录警告级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

// Error 记录错误级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

// Debug 记录调试级别日志
func Debug(format string, args ...interface{}) {
	DefaultLogger.Debug(format, args...)
}

// Info 记录信息级别日志
func Info(format string, args ...interface{}) {
	DefaultLogger.Info(format, args...)
}

// Warn 记录警告级别日志
func Warn(format string, args ...interface{}) {
	DefaultLogger.Warn(format, args...)
}

// Error 记录错误级别日志
func Error(format string, args ...interface{}) {
	DefaultLogger.Error(format, args...)
}

// SetLevel 设置默认日志记录器的日志级别
func SetLevel(level Level) {
	DefaultLogger.SetLevel(level)
}

// SetOutput 设置默认日志记录器的输出
func SetOutput(output io.Writer) {
	DefaultLogger.SetOutput(output)
}

// SetPrefix 设置默认日志记录器的前缀
func SetPrefix(prefix string) {
	DefaultLogger.SetPrefix(prefix)
}

// InitLogger 初始化日志记录器
func InitLogger(level, output, file string) error {
	// 设置日志级别
	logLevel := ParseLevel(level)
	SetLevel(logLevel)

	// 设置日志输出
	switch strings.ToLower(output) {
	case "stdout":
		SetOutput(os.Stdout)
	case "file":
		if file == "" {
			return fmt.Errorf("日志文件路径不能为空")
		}
		// 创建日志目录
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %w", err)
		}
		// 打开日志文件
		f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %w", err)
		}
		SetOutput(f)
	default:
		return fmt.Errorf("不支持的日志输出类型: %s", output)
	}

	Info("日志系统初始化完成，级别: %s, 输出: %s", level, output)
	return nil
}
