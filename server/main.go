package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/senma231/p3/common/logger"
	"github.com/senma231/p3/server/api"
	"github.com/senma231/p3/server/app"
	"github.com/senma231/p3/server/auth"
	"github.com/senma231/p3/server/config"
	"github.com/senma231/p3/server/db"
	"github.com/senma231/p3/server/device"
	"github.com/senma231/p3/server/forward"
)

// 服务器启动时间
var startTime = time.Now()

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	initDB := flag.Bool("init-db", false, "初始化数据库")
	flag.Parse()

	// 初始化日志
	logger.Init(logger.InfoLevel, os.Stdout)
	logger.Info("服务器启动中...")

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("加载配置失败: %v", err)
	}
	logger.Info("加载配置成功")

	// 初始化数据库
	if err := db.InitDB(cfg); err != nil {
		logger.Fatal("初始化数据库失败: %v", err)
	}
	logger.Info("初始化数据库成功")

	// 如果只是初始化数据库，则退出
	if *initDB {
		logger.Info("数据库初始化完成，退出")
		return
	}

	// 初始化服务
	authService := auth.NewService(cfg)
	deviceService := device.NewService()
	appService := app.NewService()
	forwardService := forward.NewService()
	logger.Info("初始化服务成功")

	// 设置路由
	router := api.SetupRouter(authService, deviceService, appService, forwardService)

	// 将服务注入到上下文中
	router.Use(func(c *gin.Context) {
		c.Set("authService", authService)
		c.Set("deviceService", deviceService)
		c.Set("appService", appService)
		c.Set("forwardService", forwardService)
		c.Next()
	})

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("服务器已启动，监听地址: %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("启动服务器失败: %v", err)
		}
	}()

	<-quit
	logger.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭数据库连接
	db.CloseDB()

	// 关闭 HTTP 服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("关闭服务器失败: %v", err)
	}

	logger.Info("服务器已关闭")
}
