package init

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/senma231/p3/common/logger"
)

// InitWorkDir 初始化工作目录
func InitWorkDir(dir string) error {
	// 如果目录不存在，则创建
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建工作目录失败: %w", err)
		}
	}

	// 切换到工作目录
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("切换到工作目录失败: %w", err)
	}

	return nil
}

// InitLogger 初始化日志
func InitLogger(level logger.Level, logFile string) error {
	// 如果日志文件为空，则使用标准输出
	if logFile == "" {
		logger.Init(level, os.Stdout)
		return nil
	}

	// 创建日志目录
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 初始化日志
	logger.Init(level, file)
	return nil
}

// GetDefaultWorkDir 获取默认工作目录
func GetDefaultWorkDir() string {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 如果获取失败，则使用当前目录
		return "."
	}

	// 根据操作系统创建不同的工作目录
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", "P3")
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "P3")
	default:
		return filepath.Join(homeDir, ".p3")
	}
}

// GetDefaultConfigFile 获取默认配置文件路径
func GetDefaultConfigFile() string {
	return filepath.Join(GetDefaultWorkDir(), "config.yaml")
}

// GetDefaultLogFile 获取默认日志文件路径
func GetDefaultLogFile() string {
	return filepath.Join(GetDefaultWorkDir(), "logs", "p3.log")
}

// GetDefaultCertDir 获取默认证书目录
func GetDefaultCertDir() string {
	return filepath.Join(GetDefaultWorkDir(), "certs")
}

// GetDefaultDataDir 获取默认数据目录
func GetDefaultDataDir() string {
	return filepath.Join(GetDefaultWorkDir(), "data")
}

// GetDefaultTempDir 获取默认临时目录
func GetDefaultTempDir() string {
	return filepath.Join(GetDefaultWorkDir(), "temp")
}

// InitDirs 初始化所有目录
func InitDirs() error {
	// 初始化工作目录
	if err := InitWorkDir(GetDefaultWorkDir()); err != nil {
		return err
	}

	// 初始化证书目录
	if err := os.MkdirAll(GetDefaultCertDir(), 0755); err != nil {
		return fmt.Errorf("创建证书目录失败: %w", err)
	}

	// 初始化数据目录
	if err := os.MkdirAll(GetDefaultDataDir(), 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 初始化临时目录
	if err := os.MkdirAll(GetDefaultTempDir(), 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 初始化日志目录
	logDir := filepath.Dir(GetDefaultLogFile())
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	return nil
}

// Cleanup 清理资源
func Cleanup() {
	// 清理临时目录
	tempDir := GetDefaultTempDir()
	if _, err := os.Stat(tempDir); err == nil {
		// 删除临时目录中的所有文件
		files, err := os.ReadDir(tempDir)
		if err == nil {
			for _, file := range files {
				os.Remove(filepath.Join(tempDir, file.Name()))
			}
		}
	}
}
