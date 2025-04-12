package p2p

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/senma231/p3/common/errors"
	"github.com/senma231/p3/common/logger"
	"github.com/senma231/p3/server/auth"
	"github.com/senma231/p3/server/config"
	"github.com/senma231/p3/server/device"
)

// SignalType 信令类型
type SignalType string

const (
	SignalOffer           SignalType = "offer"
	SignalAnswer          SignalType = "answer"
	SignalICECandidate    SignalType = "ice-candidate"
	SignalConnect         SignalType = "connect"
	SignalDisconnect      SignalType = "disconnect"
	SignalPing            SignalType = "ping"
	SignalPong            SignalType = "pong"
	SignalRelayRequest    SignalType = "relay-request"
	SignalRelayResponse   SignalType = "relay-response"
	SignalError           SignalType = "error"
)

// Signal 信令消息
type Signal struct {
	Type      SignalType  `json:"type"`
	SenderID  string      `json:"senderId"`
	ReceiverID string     `json:"receiverId,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Client WebSocket 客户端
type Client struct {
	NodeID     string
	DeviceID   uint
	Conn       *websocket.Conn
	Send       chan []byte
	LastActive time.Time
}

// SignalingServer 信令服务器
type SignalingServer struct {
	config         *config.Config
	coordinator    *Coordinator
	authService    *auth.Service
	deviceService  *device.Service
	clients        map[string]*Client
	upgrader       websocket.Upgrader
	mu             sync.RWMutex
	stopCh         chan struct{}
}

// NewSignalingServer 创建信令服务器
func NewSignalingServer(cfg *config.Config, coordinator *Coordinator, authService *auth.Service, deviceService *device.Service) *SignalingServer {
	return &SignalingServer{
		config:         cfg,
		coordinator:    coordinator,
		authService:    authService,
		deviceService:  deviceService,
		clients:        make(map[string]*Client),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源
			},
		},
		stopCh: make(chan struct{}),
	}
}

// Start 启动信令服务器
func (s *SignalingServer) Start() {
	// 启动清理协程
	go s.cleanupLoop()
	logger.Info("信令服务器已启动")
}

// Stop 停止信令服务器
func (s *SignalingServer) Stop() {
	close(s.stopCh)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 关闭所有客户端连接
	for _, client := range s.clients {
		client.Conn.Close()
		close(client.Send)
	}
	
	logger.Info("信令服务器已停止")
}

// HandleWebSocket 处理 WebSocket 连接
func (s *SignalingServer) HandleWebSocket(c *gin.Context) {
	// 获取设备 ID
	deviceID, exists := c.Get("deviceID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取节点 ID
	nodeID, exists := c.Get("nodeID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 升级 HTTP 连接为 WebSocket
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("升级 WebSocket 失败: %v", err)
		return
	}

	// 创建客户端
	client := &Client{
		NodeID:     nodeID.(string),
		DeviceID:   deviceID.(uint),
		Conn:       conn,
		Send:       make(chan []byte, 256),
		LastActive: time.Now(),
	}

	// 注册客户端
	s.mu.Lock()
	s.clients[client.NodeID] = client
	s.mu.Unlock()

	logger.Info("WebSocket 客户端已连接: %s", client.NodeID)

	// 启动读写协程
	go s.readPump(client)
	go s.writePump(client)

	// 发送欢迎消息
	welcomeSignal := Signal{
		Type:      SignalPing,
		SenderID:  "server",
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(welcomeSignal)
	client.Send <- data
}

// readPump 从 WebSocket 读取数据
func (s *SignalingServer) readPump(client *Client) {
	defer func() {
		s.unregisterClient(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(4096)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastActive = time.Now()
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket 读取错误: %v", err)
			}
			break
		}

		// 解析信令消息
		var signal Signal
		if err := json.Unmarshal(message, &signal); err != nil {
			logger.Error("解析信令消息失败: %v", err)
			continue
		}

		// 设置发送者 ID
		signal.SenderID = client.NodeID
		signal.Timestamp = time.Now()

		// 处理信令消息
		s.handleSignal(client, &signal)
	}
}

// writePump 向 WebSocket 写入数据
func (s *SignalingServer) writePump(client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// 通道已关闭
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 添加队列中的消息
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSignal 处理信令消息
func (s *SignalingServer) handleSignal(client *Client, signal *Signal) {
	// 更新最后活动时间
	client.LastActive = time.Now()

	// 处理不同类型的信令
	switch signal.Type {
	case SignalPing:
		// 回复 pong
		pongSignal := Signal{
			Type:      SignalPong,
			SenderID:  "server",
			ReceiverID: client.NodeID,
			Timestamp: time.Now(),
		}
		s.sendSignal(client, &pongSignal)

	case SignalConnect:
		// 处理连接请求
		s.handleConnectSignal(client, signal)

	case SignalOffer, SignalAnswer, SignalICECandidate:
		// 转发给接收者
		s.forwardSignal(signal)

	case SignalRelayRequest:
		// 处理中继请求
		s.handleRelayRequest(client, signal)

	default:
		// 未知信令类型
		errorSignal := Signal{
			Type:      SignalError,
			SenderID:  "server",
			ReceiverID: client.NodeID,
			Payload:   "未知的信令类型",
			Timestamp: time.Now(),
		}
		s.sendSignal(client, &errorSignal)
	}
}

// handleConnectSignal 处理连接请求
func (s *SignalingServer) handleConnectSignal(client *Client, signal *Signal) {
	// 检查接收者是否存在
	if signal.ReceiverID == "" {
		errorSignal := Signal{
			Type:      SignalError,
			SenderID:  "server",
			ReceiverID: client.NodeID,
			Payload:   "接收者 ID 不能为空",
			Timestamp: time.Now(),
		}
		s.sendSignal(client, &errorSignal)
		return
	}

	// 检查接收者是否在线
	s.mu.RLock()
	_, exists := s.clients[signal.ReceiverID]
	s.mu.RUnlock()

	if !exists {
		errorSignal := Signal{
			Type:      SignalError,
			SenderID:  "server",
			ReceiverID: client.NodeID,
			Payload:   "接收者不在线",
			Timestamp: time.Now(),
		}
		s.sendSignal(client, &errorSignal)
		return
	}

	// 确定连接类型
	connectionType, err := s.coordinator.DetermineConnectionType(client.NodeID, signal.ReceiverID)
	if err != nil {
		errorSignal := Signal{
			Type:      SignalError,
			SenderID:  "server",
			ReceiverID: client.NodeID,
			Payload:   fmt.Sprintf("确定连接类型失败: %v", err),
			Timestamp: time.Now(),
		}
		s.sendSignal(client, &errorSignal)
		return
	}

	// 创建连接响应
	connectResponse := Signal{
		Type:      SignalConnect,
		SenderID:  "server",
		ReceiverID: client.NodeID,
		Payload: map[string]interface{}{
			"connectionType": connectionType.String(),
			"targetId":       signal.ReceiverID,
		},
		Timestamp: time.Now(),
	}
	s.sendSignal(client, &connectResponse)

	// 转发连接请求给接收者
	forwardSignal := *signal
	forwardSignal.Payload = map[string]interface{}{
		"connectionType": connectionType.String(),
		"sourceId":       client.NodeID,
	}
	s.forwardSignal(&forwardSignal)
}

// handleRelayRequest 处理中继请求
func (s *SignalingServer) handleRelayRequest(client *Client, signal *Signal) {
	// 检查接收者是否存在
	if signal.ReceiverID == "" {
		errorSignal := Signal{
			Type:      SignalError,
			SenderID:  "server",
			ReceiverID: client.NodeID,
			Payload:   "接收者 ID 不能为空",
			Timestamp: time.Now(),
		}
		s.sendSignal(client, &errorSignal)
		return
	}

	// 选择中继节点
	relayNode, err := s.coordinator.SelectRelayNode(client.NodeID, signal.ReceiverID)
	if err != nil {
		errorSignal := Signal{
			Type:      SignalError,
			SenderID:  "server",
			ReceiverID: client.NodeID,
			Payload:   fmt.Sprintf("选择中继节点失败: %v", err),
			Timestamp: time.Now(),
		}
		s.sendSignal(client, &errorSignal)
		return
	}

	// 创建中继响应
	relayResponse := Signal{
		Type:      SignalRelayResponse,
		SenderID:  "server",
		ReceiverID: client.NodeID,
		Payload: map[string]interface{}{
			"relayId":   relayNode.NodeID,
			"relayHost": relayNode.ExternalIP.String(),
			"relayPort": relayNode.ExternalPort,
			"targetId":  signal.ReceiverID,
		},
		Timestamp: time.Now(),
	}
	s.sendSignal(client, &relayResponse)

	// 转发中继请求给接收者
	forwardSignal := *signal
	forwardSignal.Type = SignalRelayResponse
	forwardSignal.Payload = map[string]interface{}{
		"relayId":   relayNode.NodeID,
		"relayHost": relayNode.ExternalIP.String(),
		"relayPort": relayNode.ExternalPort,
		"sourceId":  client.NodeID,
	}
	s.forwardSignal(&forwardSignal)
}

// forwardSignal 转发信令消息
func (s *SignalingServer) forwardSignal(signal *Signal) {
	if signal.ReceiverID == "" {
		logger.Error("转发信令失败: 接收者 ID 为空")
		return
	}

	s.mu.RLock()
	receiver, exists := s.clients[signal.ReceiverID]
	s.mu.RUnlock()

	if !exists {
		logger.Error("转发信令失败: 接收者 %s 不在线", signal.ReceiverID)
		return
	}

	data, err := json.Marshal(signal)
	if err != nil {
		logger.Error("序列化信令消息失败: %v", err)
		return
	}

	receiver.Send <- data
}

// sendSignal 发送信令消息
func (s *SignalingServer) sendSignal(client *Client, signal *Signal) {
	data, err := json.Marshal(signal)
	if err != nil {
		logger.Error("序列化信令消息失败: %v", err)
		return
	}

	client.Send <- data
}

// unregisterClient 注销客户端
func (s *SignalingServer) unregisterClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[client.NodeID]; exists {
		delete(s.clients, client.NodeID)
		close(client.Send)
		logger.Info("WebSocket 客户端已断开连接: %s", client.NodeID)
	}
}

// cleanupLoop 清理循环
func (s *SignalingServer) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanupInactiveClients()
		}
	}
}

// cleanupInactiveClients 清理不活跃的客户端
func (s *SignalingServer) cleanupInactiveClients() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for nodeID, client := range s.clients {
		if now.Sub(client.LastActive) > 5*time.Minute {
			logger.Info("清理不活跃的客户端: %s", nodeID)
			client.Conn.Close()
			close(client.Send)
			delete(s.clients, nodeID)
		}
	}
}

// GetClientCount 获取客户端数量
func (s *SignalingServer) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// IsClientOnline 检查客户端是否在线
func (s *SignalingServer) IsClientOnline(nodeID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.clients[nodeID]
	return exists
}

// RegisterRoutes 注册路由
func (s *SignalingServer) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/ws", s.authMiddleware(), s.HandleWebSocket)
}

// authMiddleware 认证中间件
func (s *SignalingServer) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取节点 ID 和令牌
		nodeID := c.GetHeader("X-Node-ID")
		token := c.GetHeader("X-Node-Token")

		if nodeID == "" || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供节点 ID 或令牌"})
			c.Abort()
			return
		}

		// 认证设备
		device, err := s.deviceService.AuthenticateDevice(nodeID, token)
		if err != nil {
			errObj := errors.AsError(err)
			c.JSON(errObj.StatusCode(), gin.H{"error": errObj.Error()})
			c.Abort()
			return
		}

		// 将设备信息存储在上下文中
		c.Set("device", device)
		c.Set("deviceID", device.ID)
		c.Set("nodeID", device.NodeID)
		c.Set("userID", device.UserID)

		c.Next()
	}
}
