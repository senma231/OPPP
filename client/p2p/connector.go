package p2p

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/senma231/p3/client/config"
	"github.com/senma231/p3/client/nat"
)

// ConnectionType 连接类型
type ConnectionType int

const (
	ConnectionTypeUnknown ConnectionType = iota
	ConnectionTypeDirect               // 直接连接
	ConnectionTypeHolePunch            // 打洞连接
	ConnectionTypeRelay                // 中继连接
)

// String 返回连接类型的字符串表示
func (t ConnectionType) String() string {
	switch t {
	case ConnectionTypeDirect:
		return "Direct"
	case ConnectionTypeHolePunch:
		return "HolePunch"
	case ConnectionTypeRelay:
		return "Relay"
	default:
		return "Unknown"
	}
}

// ConnectionResult 连接结果
type ConnectionResult struct {
	Success        bool
	Conn           net.Conn
	ConnectionType ConnectionType
	Error          error
}

// PeerInfo 对等节点信息
type PeerInfo struct {
	NodeID       string
	NATType      nat.NATType
	ExternalIP   string
	ExternalPort int
}

// Connector P2P 连接器
type Connector struct {
	config         *config.Config
	natInfo        *nat.NATInfo
	signalingClient *SignalingClient
	puncher        *Puncher
	connectResults map[string]chan *ConnectionResult
	mu             sync.RWMutex
}

// NewConnector 创建 P2P 连接器
func NewConnector(cfg *config.Config, natInfo *nat.NATInfo, signalingClient *SignalingClient) *Connector {
	connector := &Connector{
		config:         cfg,
		natInfo:        natInfo,
		signalingClient: signalingClient,
		puncher:        NewPuncher(cfg.Network.UDPPort1, natInfo, 10*time.Second, 5),
		connectResults: make(map[string]chan *ConnectionResult),
	}

	// 注册信令处理函数
	signalingClient.RegisterHandler(SignalConnect, connector.handleConnectSignal)
	signalingClient.RegisterHandler(SignalOffer, connector.handleOfferSignal)
	signalingClient.RegisterHandler(SignalAnswer, connector.handleAnswerSignal)
	signalingClient.RegisterHandler(SignalICECandidate, connector.handleICECandidateSignal)
	signalingClient.RegisterHandler(SignalRelayResponse, connector.handleRelayResponseSignal)

	return connector
}

// Connect 连接到对等节点
func (c *Connector) Connect(peerID string) (*ConnectionResult, error) {
	// 创建结果通道
	resultCh := make(chan *ConnectionResult, 1)

	// 注册结果通道
	c.mu.Lock()
	c.connectResults[peerID] = resultCh
	c.mu.Unlock()

	// 发送连接请求
	if err := c.signalingClient.RequestConnect(peerID); err != nil {
		c.mu.Lock()
		delete(c.connectResults, peerID)
		c.mu.Unlock()
		return nil, fmt.Errorf("发送连接请求失败: %w", err)
	}

	// 等待连接结果
	select {
	case result := <-resultCh:
		return result, nil
	case <-time.After(30 * time.Second):
		c.mu.Lock()
		delete(c.connectResults, peerID)
		c.mu.Unlock()
		return nil, fmt.Errorf("连接超时")
	}
}

// handleConnectSignal 处理连接信令
func (c *Connector) handleConnectSignal(signal *Signal) {
	// 提取对等节点信息
	payload, ok := signal.Payload.(map[string]interface{})
	if !ok {
		fmt.Printf("无效的连接信令负载: %v\n", signal.Payload)
		return
	}

	// 检查是否是服务器响应
	if signal.SenderID == "server" {
		// 处理服务器响应
		c.handleServerConnectResponse(signal)
		return
	}

	// 处理对等节点连接请求
	natTypeStr, _ := payload["natType"].(string)
	externalIP, _ := payload["externalIP"].(string)
	externalPort, _ := payload["externalPort"].(float64)

	// 解析 NAT 类型
	var natType nat.NATType
	switch natTypeStr {
	case "No NAT (Public IP)":
		natType = nat.NATNone
	case "Full Cone NAT":
		natType = nat.NATFull
	case "Restricted Cone NAT":
		natType = nat.NATRestricted
	case "Port Restricted Cone NAT":
		natType = nat.NATPortRestricted
	case "Symmetric NAT":
		natType = nat.NATSymmetric
	default:
		natType = nat.NATUnknown
	}

	// 创建对等节点信息
	peerInfo := &PeerInfo{
		NodeID:       signal.SenderID,
		NATType:      natType,
		ExternalIP:   externalIP,
		ExternalPort: int(externalPort),
	}

	// 尝试连接
	go c.tryConnect(peerInfo)
}

// handleServerConnectResponse 处理服务器连接响应
func (c *Connector) handleServerConnectResponse(signal *Signal) {
	payload, ok := signal.Payload.(map[string]interface{})
	if !ok {
		fmt.Printf("无效的服务器连接响应负载: %v\n", signal.Payload)
		return
	}

	// 获取目标节点 ID
	targetID, ok := payload["targetId"].(string)
	if !ok {
		fmt.Printf("服务器连接响应中缺少目标节点 ID\n")
		return
	}

	// 获取连接类型
	connectionTypeStr, ok := payload["connectionType"].(string)
	if !ok {
		fmt.Printf("服务器连接响应中缺少连接类型\n")
		return
	}

	// 解析连接类型
	var connectionType ConnectionType
	switch connectionTypeStr {
	case "Direct":
		connectionType = ConnectionTypeDirect
	case "HolePunch":
		connectionType = ConnectionTypeHolePunch
	case "Relay":
		connectionType = ConnectionTypeRelay
	default:
		connectionType = ConnectionTypeUnknown
	}

	// 如果是中继连接，则发送中继请求
	if connectionType == ConnectionTypeRelay {
		if err := c.signalingClient.RequestRelay(targetID); err != nil {
			fmt.Printf("发送中继请求失败: %v\n", err)
			c.sendConnectResult(targetID, &ConnectionResult{
				Success:        false,
				ConnectionType: ConnectionTypeUnknown,
				Error:          fmt.Errorf("发送中继请求失败: %w", err),
			})
		}
	}
}

// tryConnect 尝试连接到对等节点
func (c *Connector) tryConnect(peer *PeerInfo) {
	// 尝试直接连接
	if c.canDirectConnect(peer.NATType) {
		conn, err := c.directConnect(peer.ExternalIP, peer.ExternalPort)
		if err == nil {
			c.sendConnectResult(peer.NodeID, &ConnectionResult{
				Success:        true,
				Conn:           conn,
				ConnectionType: ConnectionTypeDirect,
			})
			return
		}
		fmt.Printf("直接连接失败: %v\n", err)
	}

	// 尝试打洞连接
	result := c.puncher.Punch(peer.ExternalIP, peer.ExternalPort, peer.NATType)
	if result.Success {
		c.sendConnectResult(peer.NodeID, &ConnectionResult{
			Success:        true,
			Conn:           result.Conn,
			ConnectionType: ConnectionTypeHolePunch,
		})
		return
	}
	fmt.Printf("打洞连接失败: %v\n", result.Error)

	// 如果直接连接和打洞连接都失败，则等待中继连接
	fmt.Printf("等待中继连接...\n")
}

// canDirectConnect 检查是否可以直接连接
func (c *Connector) canDirectConnect(peerNATType nat.NATType) bool {
	// 如果对方没有 NAT，可以直接连接
	if peerNATType == nat.NATNone {
		return true
	}

	// 如果对方是完全锥形 NAT，可以直接连接
	if peerNATType == nat.NATFull {
		return true
	}

	// 如果本地没有 NAT，可以直接连接
	if c.natInfo.Type == nat.NATNone {
		return true
	}

	return false
}

// directConnect 直接连接
func (c *Connector) directConnect(peerIP string, peerPort int) (net.Conn, error) {
	// 创建 TCP 连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", peerIP, peerPort), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("直接连接失败: %w", err)
	}
	return conn, nil
}

// handleOfferSignal 处理 Offer 信令
func (c *Connector) handleOfferSignal(signal *Signal) {
	// 暂时不处理 WebRTC 信令
	fmt.Printf("收到 Offer 信令: %v\n", signal)
}

// handleAnswerSignal 处理 Answer 信令
func (c *Connector) handleAnswerSignal(signal *Signal) {
	// 暂时不处理 WebRTC 信令
	fmt.Printf("收到 Answer 信令: %v\n", signal)
}

// handleICECandidateSignal 处理 ICE 候选信令
func (c *Connector) handleICECandidateSignal(signal *Signal) {
	// 暂时不处理 WebRTC 信令
	fmt.Printf("收到 ICE 候选信令: %v\n", signal)
}

// handleRelayResponseSignal 处理中继响应信令
func (c *Connector) handleRelayResponseSignal(signal *Signal) {
	payload, ok := signal.Payload.(map[string]interface{})
	if !ok {
		fmt.Printf("无效的中继响应负载: %v\n", signal.Payload)
		return
	}

	// 获取中继信息
	relayID, _ := payload["relayId"].(string)
	relayHost, _ := payload["relayHost"].(string)
	relayPort, _ := payload["relayPort"].(float64)

	// 获取目标节点 ID
	var targetID string
	if signal.SenderID == "server" {
		targetID, _ = payload["targetId"].(string)
	} else {
		targetID = signal.SenderID
	}

	if relayHost == "" || relayPort == 0 {
		fmt.Printf("中继响应中缺少中继地址或端口\n")
		c.sendConnectResult(targetID, &ConnectionResult{
			Success:        false,
			ConnectionType: ConnectionTypeUnknown,
			Error:          fmt.Errorf("中继响应中缺少中继地址或端口"),
		})
		return
	}

	// 连接到中继服务器
	relayAddr := fmt.Sprintf("%s:%d", relayHost, int(relayPort))
	conn, err := net.DialTimeout("tcp", relayAddr, 10*time.Second)
	if err != nil {
		fmt.Printf("连接中继服务器失败: %v\n", err)
		c.sendConnectResult(targetID, &ConnectionResult{
			Success:        false,
			ConnectionType: ConnectionTypeUnknown,
			Error:          fmt.Errorf("连接中继服务器失败: %w", err),
		})
		return
	}

	// 发送中继请求
	relayRequest := fmt.Sprintf("RELAY %s", targetID)
	_, err = conn.Write([]byte(relayRequest))
	if err != nil {
		conn.Close()
		fmt.Printf("发送中继请求失败: %v\n", err)
		c.sendConnectResult(targetID, &ConnectionResult{
			Success:        false,
			ConnectionType: ConnectionTypeUnknown,
			Error:          fmt.Errorf("发送中继请求失败: %w", err),
		})
		return
	}

	// 读取中继响应
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buffer)
	if err != nil {
		conn.Close()
		fmt.Printf("读取中继响应失败: %v\n", err)
		c.sendConnectResult(targetID, &ConnectionResult{
			Success:        false,
			ConnectionType: ConnectionTypeUnknown,
			Error:          fmt.Errorf("读取中继响应失败: %w", err),
		})
		return
	}

	// 检查响应
	response := string(buffer[:n])
	if response != "OK" {
		conn.Close()
		fmt.Printf("中继服务器拒绝请求: %s\n", response)
		c.sendConnectResult(targetID, &ConnectionResult{
			Success:        false,
			ConnectionType: ConnectionTypeUnknown,
			Error:          fmt.Errorf("中继服务器拒绝请求: %s", response),
		})
		return
	}

	// 中继连接成功
	conn.SetReadDeadline(time.Time{})
	c.sendConnectResult(targetID, &ConnectionResult{
		Success:        true,
		Conn:           conn,
		ConnectionType: ConnectionTypeRelay,
	})
}

// sendConnectResult 发送连接结果
func (c *Connector) sendConnectResult(peerID string, result *ConnectionResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	resultCh, exists := c.connectResults[peerID]
	if !exists {
		// 如果没有注册结果通道，则关闭连接
		if result.Success && result.Conn != nil {
			result.Conn.Close()
		}
		return
	}

	// 发送结果
	resultCh <- result

	// 删除结果通道
	delete(c.connectResults, peerID)
}
