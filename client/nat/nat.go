package nat

import (
	"fmt"
	"net"
	"time"
)

// NATType 表示 NAT 类型
type NATType int

const (
	NATUnknown NATType = iota
	NATNone           // 无 NAT（公网 IP）
	NATFull           // 完全锥形 NAT（Full Cone）
	NATRestricted     // 受限锥形 NAT（Restricted Cone）
	NATPortRestricted // 端口受限锥形 NAT（Port Restricted Cone）
	NATSymmetric      // 对称型 NAT（Symmetric）
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
	Type       NATType
	ExternalIP net.IP
	ExternalPort int
	LocalIP    net.IP
	LocalPort  int
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
	// TODO: 实现 NAT 类型检测
	// 1. 使用 STUN 协议检测 NAT 类型
	// 2. 检测是否支持 UPnP
	// 3. 获取外部 IP 和端口

	// 临时返回一个模拟结果
	return &NATInfo{
		Type:          NATPortRestricted,
		ExternalIP:    net.ParseIP("203.0.113.1"),
		ExternalPort:  12345,
		LocalIP:       net.ParseIP("192.168.1.2"),
		LocalPort:     27182,
		UPnPAvailable: false,
	}, nil
}

// UPnPMapping 尝试通过 UPnP 映射端口
func UPnPMapping(port int, protocol string, description string) (bool, error) {
	// TODO: 实现 UPnP 端口映射
	// 1. 发现 UPnP 设备
	// 2. 获取外部 IP
	// 3. 添加端口映射

	return false, fmt.Errorf("UPnP 端口映射尚未实现")
}

// UPnPRemoveMapping 移除 UPnP 端口映射
func UPnPRemoveMapping(port int, protocol string) error {
	// TODO: 实现移除 UPnP 端口映射

	return fmt.Errorf("移除 UPnP 端口映射尚未实现")
}
