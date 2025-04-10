package core

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/senma231/p3/client/config"
	"github.com/senma231/p3/client/nat"
)

// ConnectionType 表示连接类型
type ConnectionType int

const (
	ConnectionUnknown   ConnectionType = iota
	ConnectionDirect                   // 直接连接
	ConnectionUPnP                     // UPnP 连接
	ConnectionHolePunch                // 打洞连接
	ConnectionRelay                    // 中继连接
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

// PeerInfo 存储对等节点信息
type PeerInfo struct {
	NodeID       string
	NATType      nat.NATType
	ExternalIP   net.IP
	ExternalPort int
	LastSeen     time.Time
}

// Connection 表示一个 P2P 连接
type Connection struct {
	PeerID      string
	Type        ConnectionType
	Established time.Time
	LastActive  time.Time
	BytesSent   uint64
	BytesRecv   uint64
	conn        net.Conn
	mu          sync.Mutex
}

// Send 发送数据
func (c *Connection) Send(data []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return 0, fmt.Errorf("连接已关闭")
	}

	n, err := c.conn.Write(data)
	if err != nil {
		return n, err
	}

	c.BytesSent += uint64(n)
	c.LastActive = time.Now()
	return n, nil
}

// Receive 接收数据
func (c *Connection) Receive(buf []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return 0, fmt.Errorf("连接已关闭")
	}

	n, err := c.conn.Read(buf)
	if err != nil {
		return n, err
	}

	c.BytesRecv += uint64(n)
	c.LastActive = time.Now()
	return n, nil
}

// Close 关闭连接
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	return err
}

// Engine P2P 引擎
type Engine struct {
	config      *config.Config
	natInfo     *nat.NATInfo
	peers       map[string]*PeerInfo
	connections map[string]*Connection
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewEngine 创建一个新的 P2P 引擎
func NewEngine(cfg *config.Config) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		config:      cfg,
		peers:       make(map[string]*PeerInfo),
		connections: make(map[string]*Connection),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动 P2P 引擎
func (e *Engine) Start() error {
	// 检测 NAT 类型
	detector := nat.NewDetector(nil, 5*time.Second)
	natInfo, err := detector.Detect()
	if err != nil {
		return fmt.Errorf("NAT 类型检测失败: %w", err)
	}
	e.natInfo = natInfo

	fmt.Printf("NAT 类型: %s\n", natInfo.Type)
	fmt.Printf("外部 IP: %s\n", natInfo.ExternalIP)
	fmt.Printf("外部端口: %d\n", natInfo.ExternalPort)
	fmt.Printf("UPnP 可用: %t\n", natInfo.UPnPAvailable)

	// TODO: 连接到服务器
	// TODO: 注册节点
	// TODO: 启动监听

	return nil
}

// Stop 停止 P2P 引擎
func (e *Engine) Stop() error {
	e.cancel()

	// 关闭所有连接
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, conn := range e.connections {
		if err := conn.Close(); err != nil {
			// 记录错误但继续关闭其他连接
			fmt.Printf("关闭连接 %s 失败: %v\n", conn.PeerID, err)
		}
	}

	return nil
}

// Connect 连接到对等节点
func (e *Engine) Connect(peerID string) (*Connection, error) {
	e.mu.RLock()
	_, exists := e.peers[peerID]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("未知的对等节点: %s", peerID)
	}

	// 检查是否已经连接
	e.mu.RLock()
	conn, connected := e.connections[peerID]
	e.mu.RUnlock()

	if connected {
		return conn, nil
	}

	// 尝试建立连接
	// TODO: 实现连接逻辑
	// 1. 尝试直接连接
	// 2. 尝试 UPnP 连接
	// 3. 尝试打洞连接
	// 4. 尝试中继连接

	// 临时返回一个模拟连接
	conn = &Connection{
		PeerID:      peerID,
		Type:        ConnectionDirect,
		Established: time.Now(),
		LastActive:  time.Now(),
	}

	e.mu.Lock()
	e.connections[peerID] = conn
	e.mu.Unlock()

	return conn, nil
}

// Disconnect 断开与对等节点的连接
func (e *Engine) Disconnect(peerID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	conn, exists := e.connections[peerID]
	if !exists {
		return fmt.Errorf("未连接到对等节点: %s", peerID)
	}

	if err := conn.Close(); err != nil {
		return err
	}

	delete(e.connections, peerID)
	return nil
}

// GetPeers 获取所有对等节点
func (e *Engine) GetPeers() []*PeerInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	peers := make([]*PeerInfo, 0, len(e.peers))
	for _, peer := range e.peers {
		peers = append(peers, peer)
	}

	return peers
}

// GetConnections 获取所有连接
func (e *Engine) GetConnections() []*Connection {
	e.mu.RLock()
	defer e.mu.RUnlock()

	conns := make([]*Connection, 0, len(e.connections))
	for _, conn := range e.connections {
		conns = append(conns, conn)
	}

	return conns
}

// UpdatePeer 更新对等节点信息
func (e *Engine) UpdatePeer(peer *PeerInfo) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.peers[peer.NodeID] = peer
}

// RemovePeer 移除对等节点
func (e *Engine) RemovePeer(peerID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.peers, peerID)
}
