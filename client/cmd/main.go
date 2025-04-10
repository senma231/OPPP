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
	"github.com/senma231/p3/client/forward"
	"github.com/senma231/p3/client/nat"
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
		cfg.Network.Node = *node
	}
	if *token != "" {
		cfg.Network.Token = *token
	}
	if *shareBandwidth >= 0 {
		cfg.Network.ShareBandwidth = *shareBandwidth
	}

	// 检查必要参数
	if cfg.Network.Node == "" {
		log.Fatal("节点名称不能为空，请使用 -node 参数指定")
	}
	if cfg.Network.Token == "" {
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
	fmt.Printf("版本: %s\n", cfg.Version)
	fmt.Printf("节点名称: %s\n", cfg.Network.Node)
	fmt.Printf("共享带宽: %d Mbps\n", cfg.Network.ShareBandwidth)

	// 检测 NAT 类型
	detector := nat.NewDetector(nil, 5*time.Second)
	natInfo, err := detector.Detect()
	if err != nil {
		log.Printf("NAT 类型检测失败: %v", err)
	} else {
		fmt.Printf("NAT 类型: %s\n", natInfo.Type)
		fmt.Printf("外部 IP: %s\n", natInfo.ExternalIP)
		fmt.Printf("外部端口: %d\n", natInfo.ExternalPort)
		fmt.Printf("UPnP 可用: %t\n", natInfo.UPnPAvailable)
	}

	// 初始化 P2P 引擎
	engine := core.NewEngine(cfg)
	if err := engine.Start(); err != nil {
		log.Fatalf("启动 P2P 引擎失败: %v", err)
	}

	// 初始化端口转发器
	forwarder := forward.NewForwarder()

	// 加载转发规则
	for _, appConfig := range cfg.Apps {
		rule := &forward.ForwardRule{
			ID:          appConfig.AppName,
			Protocol:    appConfig.Protocol,
			SrcPort:     appConfig.SrcPort,
			DstHost:     appConfig.DstHost,
			DstPort:     appConfig.DstPort,
			Description: appConfig.AppName,
			Enabled:     true,
		}

		if err := forwarder.AddRule(rule); err != nil {
			log.Printf("添加转发规则失败: %v", err)
		} else {
			fmt.Printf("添加转发规则: %s -> %s:%d\n", rule.ID, rule.DstHost, rule.DstPort)
		}
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

	// 关闭端口转发器
	if err := forwarder.Close(); err != nil {
		log.Printf("关闭端口转发器失败: %v", err)
	}

	// 关闭 P2P 引擎
	if err := engine.Stop(); err != nil {
		log.Printf("关闭 P2P 引擎失败: %v", err)
	}

	fmt.Println("客户端已关闭")
}
