package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/senma231/p3/server/api"
	"github.com/senma231/p3/server/app"
	"github.com/senma231/p3/server/auth"
	"github.com/senma231/p3/server/config"
	"github.com/senma231/p3/server/db"
	"github.com/senma231/p3/server/device"
	"github.com/senma231/p3/server/p2p"
	"github.com/senma231/p3/server/relay"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 打印启动信息
	fmt.Println("P3 服务端启动中...")
	fmt.Printf("版本: %s\n", cfg.Version)
	fmt.Printf("监听端口: %d\n", cfg.Server.Port)

	// 初始化数据库连接
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.CloseDB()

	// 初始化服务
	authService := auth.NewService(cfg)
	deviceService := device.NewService(cfg)
	appService := app.NewService(cfg)
	// 创建 P2P 协调器和中继服务，但暂时不使用
	_ = p2p.NewCoordinator(cfg, deviceService)
	_ = relay.NewService(cfg)

	// 设置路由
	router := api.SetupRouter(authService, deviceService, appService)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// 启动 HTTP 服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动 HTTP 服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	fmt.Println("正在关闭服务...")

	// 关闭 HTTP 服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("关闭 HTTP 服务器失败: %v", err)
	}

	fmt.Println("服务已关闭")
}
