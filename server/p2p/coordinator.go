package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/senma231/p3/server/config"
	"github.com/senma231/p3/server/db"
	"github.com/senma231/p3/server/device"
)

// NATType NAT 类型
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

// ParseNATType 解析 NAT 类型字符串
func ParseNATType(s string) NATType {
	switch s {
	case "No NAT (Public IP)":
		return NATNone
	case "Full Cone NAT":
		return NATFull
	case "Restricted Cone NAT":
		return NATRestricted
	case "Port Restricted Cone NAT":
		return NATPortRestricted
	case "Symmetric NAT":
		return NATSymmetric
	default:
		return NATUnknown
	}
}

// PeerInfo 对等节点信息
type PeerInfo struct {
	NodeID       string
	NATType      NATType
	ExternalIP   net.IP
	ExternalPort int
	LocalIP      net.IP
	LocalPort    int
	LastSeen     time.Time
}

// ConnectionType 连接类型
type ConnectionType int

const (
	ConnectionUnknown ConnectionType = iota
	ConnectionDirect               // 直接连接
	ConnectionUPnP                 // UPnP 连接
	ConnectionHolePunch            // 打洞连接
	ConnectionRelay                // 中继连接
)

// String 返回连接类型的字符串表示
func (t ConnectionType) String() string {
	switch t {
	case ConnectionDirect:
		return "Direct"
	case ConnectionUPnP:
		return "UPnP"
	case ConnectionHolePunch:
		return "Hole Punch"
	case ConnectionRelay:
		return "Relay"
	default:
		return "Unknown"
	}
}

// ParseConnectionType 解析连接类型字符串
func ParseConnectionType(s string) ConnectionType {
	switch s {
	case "Direct":
		return ConnectionDirect
	case "UPnP":
		return ConnectionUPnP
	case "Hole Punch":
		return ConnectionHolePunch
	case "Relay":
		return ConnectionRelay
	default:
		return ConnectionUnknown
	}
}

// Coordinator P2P 协调器
type Coordinator struct {
	config       *config.Config
	deviceService *device.Service
	peers        map[string]*PeerInfo
	relayNodes   map[string]*PeerInfo
	mu           sync.RWMutex
}

// NewCoordinator 创建 P2P 协调器
func NewCoordinator(cfg *config.Config, deviceService *device.Service) *Coordinator {
	return &Coordinator{
		config:       cfg,
		deviceService: deviceService,
		peers:        make(map[string]*PeerInfo),
		relayNodes:   make(map[string]*PeerInfo),
	}
}

// RegisterPeer 注册对等节点
func (c *Coordinator) RegisterPeer(nodeID string, natType NATType, externalIP net.IP, externalPort int, localIP net.IP, localPort int) error {
	// 验证设备是否存在
	_, err := c.deviceService.GetDeviceByNodeID(nodeID)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 更新或添加对等节点信息
	c.peers[nodeID] = &PeerInfo{
		NodeID:       nodeID,
		NATType:      natType,
		ExternalIP:   externalIP,
		ExternalPort: externalPort,
		LocalIP:      localIP,
		LocalPort:    localPort,
		LastSeen:     time.Now(),
	}

	// 如果是公网 IP 或完全锥形 NAT，可以作为中继节点
	if natType == NATNone || natType == NATFull {
		c.relayNodes[nodeID] = c.peers[nodeID]
	}

	return nil
}

// UnregisterPeer 注销对等节点
func (c *Coordinator) UnregisterPeer(nodeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.peers, nodeID)
	delete(c.relayNodes, nodeID)
}

// GetPeerInfo 获取对等节点信息
func (c *Coordinator) GetPeerInfo(nodeID string) (*PeerInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	peer, ok := c.peers[nodeID]
	if !ok {
		return nil, errors.New("对等节点不存在")
	}

	return peer, nil
}

// GetAllPeers 获取所有对等节点
func (c *Coordinator) GetAllPeers() []*PeerInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	peers := make([]*PeerInfo, 0, len(c.peers))
	for _, peer := range c.peers {
		peers = append(peers, peer)
	}

	return peers
}

// GetRelayNodes 获取所有中继节点
func (c *Coordinator) GetRelayNodes() []*PeerInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	relayNodes := make([]*PeerInfo, 0, len(c.relayNodes))
	for _, node := range c.relayNodes {
		relayNodes = append(relayNodes, node)
	}

	return relayNodes
}

// SelectRelayNode 选择中继节点
func (c *Coordinator) SelectRelayNode(sourceNodeID, targetNodeID string) (*PeerInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 如果没有中继节点，返回错误
	if len(c.relayNodes) == 0 {
		return nil, errors.New("没有可用的中继节点")
	}

	// TODO: 实现更复杂的中继节点选择算法
	// 目前简单地选择第一个中继节点
	for _, node := range c.relayNodes {
		// 不要选择源节点或目标节点作为中继
		if node.NodeID != sourceNodeID && node.NodeID != targetNodeID {
			return node, nil
		}
	}

	return nil, errors.New("没有合适的中继节点")
}

// DetermineConnectionType 确定连接类型
func (c *Coordinator) DetermineConnectionType(sourceNodeID, targetNodeID string) (ConnectionType, error) {
	sourcePeer, err := c.GetPeerInfo(sourceNodeID)
	if err != nil {
		return ConnectionUnknown, err
	}

	targetPeer, err := c.GetPeerInfo(targetNodeID)
	if err != nil {
		return ConnectionUnknown, err
	}

	// 如果两个节点都在同一个局域网，可以直接连接
	if sourcePeer.LocalIP.Equal(targetPeer.LocalIP) {
		return ConnectionDirect, nil
	}

	// 如果目标节点是公网 IP，可以直接连接
	if targetPeer.NATType == NATNone {
		return ConnectionDirect, nil
	}

	// 如果源节点是公网 IP，可以直接连接
	if sourcePeer.NATType == NATNone {
		return ConnectionDirect, nil
	}

	// 如果目标节点支持 UPnP，可以使用 UPnP 连接
	// TODO: 实现 UPnP 检测

	// 根据 NAT 类型确定是否可以打洞
	if c.canHolePunch(sourcePeer.NATType, targetPeer.NATType) {
		return ConnectionHolePunch, nil
	}

	// 如果无法打洞，使用中继连接
	return ConnectionRelay, nil
}

// canHolePunch 判断两个 NAT 类型是否可以打洞
func (c *Coordinator) canHolePunch(sourceNAT, targetNAT NATType) bool {
	// 如果任一节点是对称型 NAT，无法打洞
	if sourceNAT == NATSymmetric && targetNAT == NATSymmetric {
		return false
	}

	// 其他情况可以尝试打洞
	return true
}

// RecordConnection 记录连接
func (c *Coordinator) RecordConnection(sourceDeviceID, targetDeviceID uint, connectionType ConnectionType) error {
	// 创建连接记录
	connection := &db.Connection{
		SourceDeviceID: sourceDeviceID,
		TargetDeviceID: targetDeviceID,
		Type:           connectionType.String(),
		Status:         "established",
		EstablishedAt:  time.Now(),
		LastActiveAt:   time.Now(),
	}

	if err := db.DB.Create(connection).Error; err != nil {
		return fmt.Errorf("创建连接记录失败: %w", err)
	}

	return nil
}

// UpdateConnectionStats 更新连接统计信息
func (c *Coordinator) UpdateConnectionStats(connectionID uint, bytesSent, bytesReceived uint64) error {
	var connection db.Connection
	if err := db.DB.First(&connection, connectionID).Error; err != nil {
		return fmt.Errorf("查询连接失败: %w", err)
	}

	updates := map[string]interface{}{
		"bytes_sent":     connection.BytesSent + bytesSent,
		"bytes_received": connection.BytesReceived + bytesReceived,
		"last_active_at": time.Now(),
	}

	if err := db.DB.Model(&connection).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新连接统计信息失败: %w", err)
	}

	return nil
}

// CloseConnection 关闭连接
func (c *Coordinator) CloseConnection(connectionID uint) error {
	var connection db.Connection
	if err := db.DB.First(&connection, connectionID).Error; err != nil {
		return fmt.Errorf("查询连接失败: %w", err)
	}

	if err := db.DB.Model(&connection).Update("status", "closed").Error; err != nil {
		return fmt.Errorf("更新连接状态失败: %w", err)
	}

	return nil
}
