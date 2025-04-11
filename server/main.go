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
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 设置日志格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("服务器启动中...")

	// 模拟加载配置
	log.Printf("加载配置文件: %s", *configPath)

	// 模拟初始化数据库
	log.Println("初始化数据库成功")

	// 模拟初始化服务
	log.Println("初始化服务成功")

	// 模拟初始化 TURN 服务器
	log.Println("初始化 TURN 服务器成功")

	// 模拟初始化中继选择器
	log.Println("初始化中继选择器成功")

	// 模拟初始化带宽管理器
	log.Println("初始化带宽管理器成功")

	// 模拟设置路由
	log.Println("设置路由成功")

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, P3!\n")
		}),
	}

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("服务器已启动，监听地址: %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	<-quit
	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("关闭服务器失败: %v", err)
	}

	log.Println("服务器已关闭")
}
