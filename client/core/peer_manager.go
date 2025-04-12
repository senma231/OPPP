package core

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/senma231/p3/client/nat"
	"github.com/senma231/p3/client/p2p"
	"github.com/senma231/p3/common/logger"
)

// PeerInfo 对等节点信息
type PeerInfo struct {
	NodeID       string
	NATType      nat.NATType
	ExternalIP   string
	ExternalPort int
	LastSeen     time.Time
}

// PeerStatus 对等节点状态
type PeerStatus struct {
	NodeID         string
	NATType        nat.NATType
	ExternalIP     string
	ExternalPort   int
	Connected      bool
	ConnectionType p2p.ConnectionType
	BytesSent      uint64
	BytesReceived  uint64
	LastSeen       time.Time
	LastActive     time.Time
}

// PeerManager 对等节点管理器
type PeerManager struct {
	peers map[string]*peerConnection
	mu    sync.RWMutex
}

// peerConnection 对等节点连接
type peerConnection struct {
	info          *PeerInfo
	conn          net.Conn
	connType      p2p.ConnectionType
	connected     bool
	bytesSent     uint64
	bytesReceived uint64
	lastActive    time.Time
	mu            sync.RWMutex
}

// NewPeerManager 创建对等节点管理器
func NewPeerManager() *PeerManager {
	return &PeerManager{
		peers: make(map[string]*peerConnection),
	}
}

// AddPeer 添加对等节点
func (m *PeerManager) AddPeer(nodeID string, info *PeerInfo, conn net.Conn, connType p2p.ConnectionType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在
	if peer, exists := m.peers[nodeID]; exists {
		// 如果已存在，关闭旧连接
		if peer.conn != nil {
			peer.conn.Close()
		}
		// 更新连接
		peer.info = info
		peer.conn = conn
		peer.connType = connType
		peer.connected = true
		peer.lastActive = time.Now()
		return
	}

	// 创建新的对等节点连接
	m.peers[nodeID] = &peerConnection{
		info:       info,
		conn:       conn,
		connType:   connType,
		connected:  true,
		lastActive: time.Now(),
	}
}

// RemovePeer 移除对等节点
func (m *PeerManager) RemovePeer(nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	peer, exists := m.peers[nodeID]
	if !exists {
		return fmt.Errorf("对等节点不存在: %s", nodeID)
	}

	// 关闭连接
	if peer.conn != nil {
		peer.conn.Close()
	}

	// 移除对等节点
	delete(m.peers, nodeID)
	return nil
}

// GetPeer 获取对等节点
func (m *PeerManager) GetPeer(nodeID string) (*peerConnection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peer, exists := m.peers[nodeID]
	if !exists {
		return nil, fmt.Errorf("对等节点不存在: %s", nodeID)
	}

	return peer, nil
}

// GetPeerStatus 获取对等节点状态
func (m *PeerManager) GetPeerStatus(nodeID string) (*PeerStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peer, exists := m.peers[nodeID]
	if !exists {
		return nil, fmt.Errorf("对等节点不存在: %s", nodeID)
	}

	peer.mu.RLock()
	defer peer.mu.RUnlock()

	return &PeerStatus{
		NodeID:         peer.info.NodeID,
		NATType:        peer.info.NATType,
		ExternalIP:     peer.info.ExternalIP,
		ExternalPort:   peer.info.ExternalPort,
		Connected:      peer.connected,
		ConnectionType: peer.connType,
		BytesSent:      peer.bytesSent,
		BytesReceived:  peer.bytesReceived,
		LastSeen:       peer.info.LastSeen,
		LastActive:     peer.lastActive,
	}, nil
}

// GetAllPeers 获取所有对等节点
func (m *PeerManager) GetAllPeers() map[string]*PeerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*PeerStatus, len(m.peers))
	for nodeID, peer := range m.peers {
		peer.mu.RLock()
		result[nodeID] = &PeerStatus{
			NodeID:         peer.info.NodeID,
			NATType:        peer.info.NATType,
			ExternalIP:     peer.info.ExternalIP,
			ExternalPort:   peer.info.ExternalPort,
			Connected:      peer.connected,
			ConnectionType: peer.connType,
			BytesSent:      peer.bytesSent,
			BytesReceived:  peer.bytesReceived,
			LastSeen:       peer.info.LastSeen,
			LastActive:     peer.lastActive,
		}
		peer.mu.RUnlock()
	}

	return result
}

// Send 发送数据
func (m *PeerManager) Send(nodeID string, data []byte) (int, error) {
	peer, err := m.GetPeer(nodeID)
	if err != nil {
		return 0, err
	}

	peer.mu.Lock()
	defer peer.mu.Unlock()

	if !peer.connected || peer.conn == nil {
		return 0, fmt.Errorf("对等节点未连接: %s", nodeID)
	}

	n, err := peer.conn.Write(data)
	if err != nil {
		peer.connected = false
		return 0, fmt.Errorf("发送数据失败: %w", err)
	}

	peer.bytesSent += uint64(n)
	peer.lastActive = time.Now()
	return n, nil
}

// Receive 接收数据
func (m *PeerManager) Receive(nodeID string, buffer []byte) (int, error) {
	peer, err := m.GetPeer(nodeID)
	if err != nil {
		return 0, err
	}

	peer.mu.Lock()
	defer peer.mu.Unlock()

	if !peer.connected || peer.conn == nil {
		return 0, fmt.Errorf("对等节点未连接: %s", nodeID)
	}

	n, err := peer.conn.Read(buffer)
	if err != nil {
		peer.connected = false
		return 0, fmt.Errorf("接收数据失败: %w", err)
	}

	peer.bytesReceived += uint64(n)
	peer.lastActive = time.Now()
	return n, nil
}

// CleanupInactivePeers 清理不活跃的对等节点
func (m *PeerManager) CleanupInactivePeers(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for nodeID, peer := range m.peers {
		peer.mu.RLock()
		inactive := now.Sub(peer.lastActive) > timeout
		peer.mu.RUnlock()

		if inactive {
			logger.Info("清理不活跃的对等节点: %s", nodeID)
			if peer.conn != nil {
				peer.conn.Close()
			}
			delete(m.peers, nodeID)
		}
	}
}
