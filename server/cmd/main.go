ackage main

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
	"github.com/senma231/p3/server/forward"
	"github.com/senma231/p3/server/p2p"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	logLevel := flag.String("log-level", "info", "日志级别 (debug, info, warn, error)")
	flag.Parse()

	// 设置日志级别
	switch *logLevel {
	case "debug":
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		log.Println("日志级别设置为: DEBUG")
	case "info":
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		log.Println("日志级别设置为: INFO")
	case "warn":
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		log.Println("日志级别设置为: WARN")
	case "error":
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		log.Println("日志级别设置为: ERROR")
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 打印启动信息
	log.Println("P3 服务端启动中...")
	log.Printf("版本: %s", cfg.Version)
	log.Printf("监听端口: %d", cfg.Server.Port)

	// 初始化数据库连接
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.CloseDB()

	// 初始化服务
	authService := auth.NewService(cfg)
	deviceService := device.NewService(cfg)
	appService := app.NewService(cfg)
	forwardService := forward.NewService()

	// 初始化 P2P 协调器
	coordinator := p2p.NewCoordinator(cfg, deviceService)

	// 初始化中继服务器
	relayServer := p2p.NewRelayServer(cfg, coordinator)
	if err := relayServer.Start(); err != nil {
		log.Printf("启动中继服务器失败: %v", err)
	}

	// 初始化信令服务器
	signalingServer := p2p.NewSignalingServer(cfg, coordinator, authService, deviceService)
	signalingServer.Start()

	// 设置路由
	router := api.SetupRouter(authService, deviceService, appService, forwardService)

	// 注册信令服务路由
	signalingServer.RegisterRoutes(router.Group("/api/v1"))

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// 启动 HTTP 服务器
	go func() {
		log.Printf("HTTP 服务器已启动，监听地址: %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动 HTTP 服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	log.Println("正在关闭服务...")

	// 停止信令服务器
	signalingServer.Stop()

	// 停止中继服务器
	if err := relayServer.Stop(); err != nil {
		log.Printf("停止中继服务器失败: %v", err)
	}

	// 关闭 HTTP 服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("关闭 HTTP 服务器失败: %v", err)
	}

	log.Println("服务已关闭")
}
