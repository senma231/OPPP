package nat

import (
	"fmt"
	"net"
	"time"
)

// NATType 表示 NAT 类型
type NATType int

const (
	NATUnknown        NATType = iota
	NATNone                   // 无 NAT（公网 IP）
	NATFull                   // 完全锥形 NAT（Full Cone）
	NATRestricted             // 受限锥形 NAT（Restricted Cone）
	NATPortRestricted         // 端口受限锥形 NAT（Port Restricted Cone）
	NATSymmetric              // 对称型 NAT（Symmetric）
)

// String 返回 NAT 类型的字符串表示
func (t NATType) String() string {
	switch t {
	case NATNone:
		return "No NAT (Public IP)"
	case NATFull:
		return "Full Cone NAT"
	case NATRestricted:
		return "Restricted Cone NAT"
	case NATPortRestricted:
		return "Port Restricted Cone NAT"
	case NATSymmetric:
		return "Symmetric NAT"
	default:
		return "Unknown NAT Type"
	}
}

// NATInfo 存储 NAT 相关信息
type NATInfo struct {
	Type          NATType
	ExternalIP    net.IP
	ExternalPort  int
	LocalIP       net.IP
	LocalPort     int
	UPnPAvailable bool
}

// Detector NAT 类型检测器
type Detector struct {
	STUNServers []string
	Timeout     time.Duration
}

// NewDetector 创建一个新的 NAT 类型检测器
func NewDetector(stunServers []string, timeout time.Duration) *Detector {
	if len(stunServers) == 0 {
		// 默认 STUN 服务器
		stunServers = []string{
			"stun.l.google.com:19302",
			"stun1.l.google.com:19302",
			"stun2.l.google.com:19302",
			"stun3.l.google.com:19302",
			"stun4.l.google.com:19302",
		}
	}

	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &Detector{
		STUNServers: stunServers,
		Timeout:     timeout,
	}
}

// Detect 检测 NAT 类型
func (d *Detector) Detect() (*NATInfo, error) {
	// 创建 STUN 客户端
	stunClient := NewSTUNClient(d.STUNServers, d.Timeout)

	// 检测 NAT 类型
	natType, err := stunClient.DetectNATType()
	if err != nil {
		return nil, fmt.Errorf("NAT 类型检测失败: %w", err)
	}

	// 获取外部 IP 和端口
	externalIP, externalPort, err := stunClient.Discover()
	if err != nil {
		return nil, fmt.Errorf("获取外部地址失败: %w", err)
	}

	// 获取本地 IP
	localIP, err := getLocalIP()
	if err != nil {
		return nil, fmt.Errorf("获取本地 IP 失败: %w", err)
	}

	// 检测是否支持 UPnP
	upnpAvailable := false
	if natType != NATNone {
		// 尝试映射一个测试端口
		available, _ := UPnPMapping(12345, "UDP", "P3 NAT Test")
		upnpAvailable = available
		// 如果成功映射，删除映射
		if upnpAvailable {
			_ = UPnPRemoveMapping(12345, "UDP")
		}
	}

	return &NATInfo{
		Type:          natType,
		ExternalIP:    externalIP,
		ExternalPort:  externalPort,
		LocalIP:       localIP,
		LocalPort:     0, // 当前未知，需要在实际使用时设置
		UPnPAvailable: upnpAvailable,
	}, nil
}

// UPnPMapping 尝试通过 UPnP 映射端口
func UPnPMapping(port int, protocol string, description string) (bool, error) {
	// 创建 UPnP 客户端
	upnpClient := NewUPnPClient(5 * time.Second)

	// 检查 UPnP 是否可用
	if !upnpClient.IsUPnPAvailable() {
		return false, fmt.Errorf("UPnP 不可用")
	}

	// 添加端口映射
	success, _, err := upnpClient.AddPortMapping(port, port, protocol, description)
	if err != nil {
		return false, fmt.Errorf("添加端口映射失败: %w", err)
	}

	return success, nil
}

// UPnPRemoveMapping 移除 UPnP 端口映射
func UPnPRemoveMapping(port int, protocol string) error {
	// 创建 UPnP 客户端
	upnpClient := NewUPnPClient(5 * time.Second)

	// 删除端口映射
	err := upnpClient.DeletePortMapping(port, protocol)
	if err != nil {
		return fmt.Errorf("删除端口映射失败: %w", err)
	}

	return nil
}
