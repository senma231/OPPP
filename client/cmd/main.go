package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/senma231/p3/client/config"
	"github.com/senma231/p3/client/core"
	"github.com/senma231/p3/client/nat"
	"github.com/senma231/p3/client/p2p"
	"github.com/senma231/p3/client/service"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	node := flag.String("node", "", "节点名称")
	token := flag.String("token", "", "认证令牌")
	daemon := flag.Bool("d", false, "以守护进程模式运行")
	install := flag.Bool("install", false, "安装为系统服务")
	uninstall := flag.Bool("uninstall", false, "卸载系统服务")
	shareBandwidth := flag.Int("sharebandwidth", 10, "共享带宽（Mbps），0表示不共享")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		// 如果配置文件不存在，使用默认配置
		cfg = config.DefaultConfig()
	}

	// 命令行参数覆盖配置文件
	if *node != "" {
		cfg.Node.ID = *node
	}
	if *token != "" {
		cfg.Node.Token = *token
	}
	if *shareBandwidth >= 0 {
		// 保存共享带宽设置
		cfg.Performance.BandwidthLimit.Upload = *shareBandwidth
	}

	// 检查必要参数
	if cfg.Node.ID == "" {
		log.Fatal("节点名称不能为空，请使用 -node 参数指定")
	}
	if cfg.Node.Token == "" {
		log.Fatal("认证令牌不能为空，请使用 -token 参数指定")
	}

	// 处理安装/卸载命令
	if *install {
		fmt.Println("正在安装系统服务...")
		if err := service.Install(cfg); err != nil {
			log.Fatalf("安装系统服务失败: %v", err)
		}
		fmt.Println("系统服务安装成功")
		return
	}
	if *uninstall {
		fmt.Println("正在卸载系统服务...")
		if err := service.Uninstall(); err != nil {
			log.Fatalf("卸载系统服务失败: %v", err)
		}
		fmt.Println("系统服务卸载成功")
		return
	}

	// 打印启动信息
	fmt.Println("P3 客户端启动中...")
	fmt.Printf("节点 ID: %s\n", cfg.Node.ID)
	fmt.Printf("服务器地址: %s\n", cfg.Server.Address)
	fmt.Printf("共享带宽: %d Mbps\n", cfg.Performance.BandwidthLimit.Upload)

	// 检测 NAT 类型
	detector := nat.NewDetector(cfg.Network.STUNServers, 5*time.Second)
	natInfo, err := detector.Detect()
	if err != nil {
		log.Printf("NAT 类型检测失败: %v", err)
		// 创建一个默认的 NAT 信息
		natInfo = &nat.NATInfo{
			Type:          nat.NATUnknown,
			ExternalIP:    nil,
			ExternalPort:  0,
			UPnPAvailable: false,
		}
	} else {
		fmt.Printf("NAT 类型: %s\n", natInfo.Type)
		fmt.Printf("外部 IP: %s\n", natInfo.ExternalIP)
		fmt.Printf("外部端口: %d\n", natInfo.ExternalPort)
		fmt.Printf("UPnP 可用: %t\n", natInfo.UPnPAvailable)
	}

	// 创建信令客户端
	signalingClient := p2p.NewSignalingClient(cfg, natInfo)

	// 连接到信令服务器
	if err := signalingClient.Connect(); err != nil {
		log.Printf("连接到信令服务器失败: %v", err)
	} else {
		fmt.Println("已连接到信令服务器")
	}

	// 创建 P2P 连接器
	connector := p2p.NewConnector(cfg, natInfo, signalingClient)

	// 创建引擎
	engine := core.NewEngine(cfg)

	// 设置 P2P 连接器
	engine.SetConnector(connector)

	// 启动引擎
	if err := engine.Start(); err != nil {
		log.Fatalf("启动引擎失败: %v", err)
	}

	// 如果是守护进程模式，启动监控
	if *daemon {
		fmt.Println("以守护进程模式运行")
		// TODO: 实现守护进程逻辑
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	fmt.Println("正在关闭客户端...")

	// 断开与信令服务器的连接
	if err := signalingClient.Disconnect(); err != nil {
		log.Printf("断开与信令服务器的连接失败: %v", err)
	}

	// 关闭引擎
	if err := engine.Stop(); err != nil {
		log.Printf("关闭引擎失败: %v", err)
	}

	fmt.Println("客户端已关闭")
}
