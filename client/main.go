package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/senma231/p3/client/config"
	"github.com/senma231/p3/client/core"
	"github.com/senma231/p3/common/logger"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	logLevel := flag.String("log-level", "info", "日志级别 (debug, info, warn, error)")
	flag.Parse()

	// 初始化日志
	level := logger.InfoLevel
	switch *logLevel {
	case "debug":
		level = logger.DebugLevel
	case "info":
		level = logger.InfoLevel
	case "warn":
		level = logger.WarnLevel
	case "error":
		level = logger.ErrorLevel
	}
	logger.Init(level, os.Stdout)

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("加载配置失败: %v", err)
	}

	// 创建引擎
	engine := core.NewEngine(cfg)

	// 启动引擎
	if err := engine.Start(); err != nil {
		logger.Fatal("启动引擎失败: %v", err)
	}

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 打印启动信息
	fmt.Println("P3 客户端已启动")
	fmt.Println("按 Ctrl+C 退出")

	// 等待退出信号
	<-sigCh
	fmt.Println("正在关闭...")

	// 停止引擎
	if err := engine.Stop(); err != nil {
		logger.Error("停止引擎失败: %v", err)
	}

	fmt.Println("已关闭")
}
