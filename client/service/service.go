package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/senma231/p3/client/config"
)

// Install 安装为系统服务
func Install(cfg *config.Config) error {
	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 根据操作系统类型安装服务
	switch runtime.GOOS {
	case "windows":
		return installWindowsService(exePath, cfg)
	case "linux":
		return installLinuxService(exePath, cfg)
	case "darwin":
		return installMacOSService(exePath, cfg)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// Uninstall 卸载系统服务
func Uninstall() error {
	// 根据操作系统类型卸载服务
	switch runtime.GOOS {
	case "windows":
		return uninstallWindowsService()
	case "linux":
		return uninstallLinuxService()
	case "darwin":
		return uninstallMacOSService()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// installWindowsService 安装 Windows 服务
func installWindowsService(exePath string, cfg *config.Config) error {
	// 创建安装目录
	installDir := `C:\Program Files\P3`
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}

	// 复制可执行文件
	destPath := filepath.Join(installDir, "p3-client.exe")
	if err := copyFile(exePath, destPath); err != nil {
		return fmt.Errorf("复制可执行文件失败: %w", err)
	}

	// 保存配置文件
	configPath := filepath.Join(installDir, "config.yaml")
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	// 创建服务
	cmd := exec.Command("sc", "create", "P3Client", "binPath=", fmt.Sprintf(`"%s" -config "%s"`, destPath, configPath), "start=", "auto", "displayname=", "P3 Client")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("创建服务失败: %w", err)
	}

	// 启动服务
	cmd = exec.Command("sc", "start", "P3Client")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启动服务失败: %w", err)
	}

	return nil
}

// uninstallWindowsService 卸载 Windows 服务
func uninstallWindowsService() error {
	// 停止服务
	cmd := exec.Command("sc", "stop", "P3Client")
	_ = cmd.Run() // 忽略错误，服务可能已经停止

	// 删除服务
	cmd = exec.Command("sc", "delete", "P3Client")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("删除服务失败: %w", err)
	}

	// 删除安装目录
	installDir := `C:\Program Files\P3`
	if err := os.RemoveAll(installDir); err != nil {
		return fmt.Errorf("删除安装目录失败: %w", err)
	}

	return nil
}

// installLinuxService 安装 Linux 服务
func installLinuxService(exePath string, cfg *config.Config) error {
	// 创建安装目录
	installDir := "/usr/local/p3"
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}

	// 复制可执行文件
	destPath := filepath.Join(installDir, "p3-client")
	if err := copyFile(exePath, destPath); err != nil {
		return fmt.Errorf("复制可执行文件失败: %w", err)
	}
	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("设置可执行权限失败: %w", err)
	}

	// 保存配置文件
	configPath := filepath.Join(installDir, "config.yaml")
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	// 创建 systemd 服务文件
	serviceContent := fmt.Sprintf(`[Unit]
Description=P3 Client
After=network.target

[Service]
Type=simple
ExecStart=%s -config %s
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
`, destPath, configPath)

	servicePath := "/etc/systemd/system/p3-client.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("创建服务文件失败: %w", err)
	}

	// 重新加载 systemd 配置
	cmd := exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("重新加载 systemd 配置失败: %w", err)
	}

	// 启用服务
	cmd = exec.Command("systemctl", "enable", "p3-client")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启用服务失败: %w", err)
	}

	// 启动服务
	cmd = exec.Command("systemctl", "start", "p3-client")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启动服务失败: %w", err)
	}

	return nil
}

// uninstallLinuxService 卸载 Linux 服务
func uninstallLinuxService() error {
	// 停止服务
	cmd := exec.Command("systemctl", "stop", "p3-client")
	_ = cmd.Run() // 忽略错误，服务可能已经停止

	// 禁用服务
	cmd = exec.Command("systemctl", "disable", "p3-client")
	_ = cmd.Run() // 忽略错误，服务可能已经禁用

	// 删除服务文件
	servicePath := "/etc/systemd/system/p3-client.service"
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除服务文件失败: %w", err)
	}

	// 重新加载 systemd 配置
	cmd = exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("重新加载 systemd 配置失败: %w", err)
	}

	// 删除安装目录
	installDir := "/usr/local/p3"
	if err := os.RemoveAll(installDir); err != nil {
		return fmt.Errorf("删除安装目录失败: %w", err)
	}

	return nil
}

// installMacOSService 安装 macOS 服务
func installMacOSService(exePath string, cfg *config.Config) error {
	// 创建安装目录
	installDir := "/usr/local/p3"
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}

	// 复制可执行文件
	destPath := filepath.Join(installDir, "p3-client")
	if err := copyFile(exePath, destPath); err != nil {
		return fmt.Errorf("复制可执行文件失败: %w", err)
	}
	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("设置可执行权限失败: %w", err)
	}

	// 保存配置文件
	configPath := filepath.Join(installDir, "config.yaml")
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	// 创建 launchd plist 文件
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>cn.p3.client</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>-config</string>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>
`, destPath, configPath)

	plistPath := "/Library/LaunchDaemons/cn.p3.client.plist"
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("创建 plist 文件失败: %w", err)
	}

	// 加载服务
	cmd := exec.Command("launchctl", "load", plistPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("加载服务失败: %w", err)
	}

	return nil
}

// uninstallMacOSService 卸载 macOS 服务
func uninstallMacOSService() error {
	// 卸载服务
	plistPath := "/Library/LaunchDaemons/cn.p3.client.plist"
	cmd := exec.Command("launchctl", "unload", plistPath)
	_ = cmd.Run() // 忽略错误，服务可能已经卸载

	// 删除 plist 文件
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除 plist 文件失败: %w", err)
	}

	// 删除安装目录
	installDir := "/usr/local/p3"
	if err := os.RemoveAll(installDir); err != nil {
		return fmt.Errorf("删除安装目录失败: %w", err)
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	// 读取源文件
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// 写入目标文件
	return os.WriteFile(dst, data, 0644)
}
