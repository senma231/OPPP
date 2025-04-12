package core

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/senma231/p3/client/config"
	"github.com/senma231/p3/client/nat"
	"github.com/senma231/p3/client/p2p"
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
	connector   *p2p.Connector
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

// SetConnector 设置 P2P 连接器
func (e *Engine) SetConnector(connector *p2p.Connector) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.connector = connector
}

// Start 启动 P2P 引擎
func (e *Engine) Start() error {
	// 检查是否设置了连接器
	if e.connector == nil {
		// 如果没有设置连接器，则使用默认的 NAT 检测
		detector := nat.NewDetector(e.config.Network.STUNServers, 5*time.Second)
		natInfo, err := detector.Detect()
		if err != nil {
			return fmt.Errorf("NAT 类型检测失败: %w", err)
		}
		e.natInfo = natInfo

		fmt.Printf("NAT 类型: %s\n", natInfo.Type)
		fmt.Printf("外部 IP: %s\n", natInfo.ExternalIP)
		fmt.Printf("外部端口: %d\n", natInfo.ExternalPort)
		fmt.Printf("UPnP 可用: %t\n", natInfo.UPnPAvailable)
	}

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
	peer, exists := e.peers[peerID]
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
	var netConn net.Conn
	var connType ConnectionType
	var err error

	// 1. 尝试直接连接
	if peer.NATType == nat.NATNone || e.natInfo.Type == nat.NATNone {
		// 如果对方或自己有公网 IP，可以直接连接
		netConn, err = e.directConnect(peer)
		if err == nil {
			connType = ConnectionDirect
		}
	}

	// 2. 尝试 UPnP 连接
	if netConn == nil && e.natInfo.UPnPAvailable {
		netConn, err = e.upnpConnect(peer)
		if err == nil {
			connType = ConnectionUPnP
		}
	}

	// 3. 尝试打洞连接
	if netConn == nil {
		netConn, connType, err = e.holePunchConnect(peer)
	}

	// 4. 尝试中继连接
	if netConn == nil {
		netConn, err = e.relayConnect(peer)
		if err == nil {
			connType = ConnectionRelay
		}
	}

	// 如果所有尝试都失败
	if netConn == nil {
		return nil, fmt.Errorf("无法连接到对等节点: %s, 所有尝试都失败", peerID)
	}

	// 创建连接对象
	conn = &Connection{
		PeerID:      peerID,
		Type:        connType,
		Established: time.Now(),
		LastActive:  time.Now(),
		conn:        netConn,
	}

	e.mu.Lock()
	e.connections[peerID] = conn
	e.mu.Unlock()

	return conn, nil
}

// directConnect 直接连接
func (e *Engine) directConnect(peer *PeerInfo) (net.Conn, error) {
	// 创建目标地址
	peerAddr := net.JoinHostPort(peer.ExternalIP.String(), fmt.Sprintf("%d", peer.ExternalPort))

	// 尝试连接
	conn, err := net.DialTimeout("tcp", peerAddr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("直接连接失败: %w", err)
	}

	return conn, nil
}

// upnpConnect 使用 UPnP 连接
func (e *Engine) upnpConnect(peer *PeerInfo) (net.Conn, error) {
	// 使用 UPnP 映射端口
	port := 10000 + rand.Intn(10000) // 随机端口
	success, err := nat.UPnPMapping(port, "TCP", "P3 Connection")
	if err != nil || !success {
		return nil, fmt.Errorf("UPnP 映射失败: %w", err)
	}

	// 创建监听器
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// 删除端口映射
		_ = nat.UPnPRemoveMapping(port, "TCP")
		return nil, fmt.Errorf("创建监听器失败: %w", err)
	}
	defer listener.Close()

	// 通知对方连接
	// TODO: 实现信令通道，通知对方连接

	// 等待连接
	listener.(*net.TCPListener).SetDeadline(time.Now().Add(10 * time.Second))
	conn, err := listener.Accept()
	if err != nil {
		// 删除端口映射
		_ = nat.UPnPRemoveMapping(port, "TCP")
		return nil, fmt.Errorf("等待连接超时: %w", err)
	}

	// 检查连接来源
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	if !remoteAddr.IP.Equal(peer.ExternalIP) {
		conn.Close()
		// 删除端口映射
		_ = nat.UPnPRemoveMapping(port, "TCP")
		return nil, fmt.Errorf("收到非目标地址的连接: %s", remoteAddr.String())
	}

	// 返回连接
	return conn, nil
}

// holePunchConnect 使用打洞连接
func (e *Engine) holePunchConnect(peer *PeerInfo) (net.Conn, ConnectionType, error) {
	// 创建打洞器
	puncher := NewPuncher(e.config.Network.UDPPort1, e.natInfo, 10*time.Second, 5)

	// 尝试打洞
	result := puncher.Punch(peer.ExternalIP, peer.ExternalPort, peer.NATType)
	if !result.Success {
		return nil, ConnectionUnknown, fmt.Errorf("打洞失败: %v", result.Error)
	}

	// 根据打洞类型返回连接类型
	var connType ConnectionType
	if result.Type == PunchUDP {
		connType = ConnectionHolePunch
	} else if result.Type == PunchTCP {
		connType = ConnectionHolePunch
	} else {
		result.Conn.Close()
		return nil, ConnectionUnknown, fmt.Errorf("不支持的打洞类型: %s", result.Type)
	}

	return result.Conn, connType, nil
}

// relayConnect 使用中继连接
func (e *Engine) relayConnect(peer *PeerInfo) (net.Conn, error) {
	// TODO: 实现中继连接
	return nil, fmt.Errorf("中继连接尚未实现")
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
