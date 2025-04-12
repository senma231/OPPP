package p2p

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/senma231/p3/client/config"
	"github.com/senma231/p3/client/nat"
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

// SignalHandler 信令处理函数
type SignalHandler func(signal *Signal)

// SignalingClient 信令客户端
type SignalingClient struct {
	config      *config.Config
	natInfo     *nat.NATInfo
	conn        *websocket.Conn
	handlers    map[SignalType][]SignalHandler
	sendCh      chan *Signal
	stopCh      chan struct{}
	connected   bool
	reconnect   bool
	mu          sync.RWMutex
	pingTicker  *time.Ticker
	pongWait    time.Duration
	pingPeriod  time.Duration
}

// NewSignalingClient 创建信令客户端
func NewSignalingClient(cfg *config.Config, natInfo *nat.NATInfo) *SignalingClient {
	return &SignalingClient{
		config:     cfg,
		natInfo:    natInfo,
		handlers:   make(map[SignalType][]SignalHandler),
		sendCh:     make(chan *Signal, 100),
		stopCh:     make(chan struct{}),
		reconnect:  true,
		pongWait:   60 * time.Second,
		pingPeriod: 30 * time.Second,
	}
}

// Connect 连接到信令服务器
func (c *SignalingClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// 解析服务器地址
	serverURL := c.config.Server.Address
	if serverURL == "" {
		return fmt.Errorf("服务器地址为空")
	}

	// 将 HTTP 地址转换为 WebSocket 地址
	u, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("解析服务器地址失败: %w", err)
	}

	var wsURL string
	if u.Scheme == "https" {
		wsURL = "wss://" + u.Host + "/api/v1/ws"
	} else {
		wsURL = "ws://" + u.Host + "/api/v1/ws"
	}

	// 设置请求头
	header := make(map[string][]string)
	header["X-Node-ID"] = []string{c.config.Node.ID}
	header["X-Node-Token"] = []string{c.config.Node.Token}

	// 连接到 WebSocket 服务器
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return fmt.Errorf("连接到信令服务器失败: %w", err)
	}

	c.conn = conn
	c.connected = true

	// 设置 Pong 处理函数
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
		return nil
	})

	// 启动读写协程
	go c.readPump()
	go c.writePump()

	// 启动 Ping 定时器
	c.pingTicker = time.NewTicker(c.pingPeriod)
	go c.pingLoop()

	fmt.Printf("已连接到信令服务器: %s\n", wsURL)
	return nil
}

// Disconnect 断开与信令服务器的连接
func (c *SignalingClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// 停止重连
	c.reconnect = false

	// 关闭连接
	if c.conn != nil {
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
		c.conn = nil
	}

	// 停止 Ping 定时器
	if c.pingTicker != nil {
		c.pingTicker.Stop()
		c.pingTicker = nil
	}

	// 发送停止信号
	close(c.stopCh)

	c.connected = false
	fmt.Println("已断开与信令服务器的连接")
	return nil
}

// readPump 从 WebSocket 读取数据
func (c *SignalingClient) readPump() {
	defer func() {
		c.handleDisconnect()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Printf("读取信令消息失败: %v\n", err)
			break
		}

		// 解析信令消息
		var signal Signal
		if err := json.Unmarshal(message, &signal); err != nil {
			fmt.Printf("解析信令消息失败: %v\n", err)
			continue
		}

		// 处理信令消息
		c.handleSignal(&signal)
	}
}

// writePump 向 WebSocket 写入数据
func (c *SignalingClient) writePump() {
	for {
		select {
		case <-c.stopCh:
			return
		case signal := <-c.sendCh:
			c.mu.RLock()
			if !c.connected || c.conn == nil {
				c.mu.RUnlock()
				continue
			}

			// 序列化信令消息
			data, err := json.Marshal(signal)
			if err != nil {
				fmt.Printf("序列化信令消息失败: %v\n", err)
				c.mu.RUnlock()
				continue
			}

			// 发送信令消息
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Printf("发送信令消息失败: %v\n", err)
				c.mu.RUnlock()
				c.handleDisconnect()
				return
			}
			c.mu.RUnlock()
		}
	}
}

// pingLoop 发送 Ping 消息
func (c *SignalingClient) pingLoop() {
	for {
		select {
		case <-c.stopCh:
			return
		case <-c.pingTicker.C:
			c.mu.RLock()
			if !c.connected || c.conn == nil {
				c.mu.RUnlock()
				continue
			}

			// 发送 Ping 消息
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("发送 Ping 消息失败: %v\n", err)
				c.mu.RUnlock()
				c.handleDisconnect()
				return
			}
			c.mu.RUnlock()
		}
	}
}

// handleDisconnect 处理断开连接
func (c *SignalingClient) handleDisconnect() {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return
	}

	// 关闭连接
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.connected = false
	reconnect := c.reconnect
	c.mu.Unlock()

	fmt.Println("与信令服务器的连接已断开")

	// 如果需要重连，则尝试重连
	if reconnect {
		go c.reconnectLoop()
	}
}

// reconnectLoop 重连循环
func (c *SignalingClient) reconnectLoop() {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		// 检查是否已停止重连
		c.mu.RLock()
		if !c.reconnect {
			c.mu.RUnlock()
			return
		}
		c.mu.RUnlock()

		// 等待一段时间后重连
		time.Sleep(backoff)

		// 尝试重连
		fmt.Printf("尝试重新连接到信令服务器...\n")
		err := c.Connect()
		if err == nil {
			fmt.Println("重新连接成功")
			return
		}

		fmt.Printf("重新连接失败: %v\n", err)

		// 增加重连间隔，但不超过最大值
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// handleSignal 处理信令消息
func (c *SignalingClient) handleSignal(signal *Signal) {
	// 处理特殊信令类型
	switch signal.Type {
	case SignalPing:
		// 回复 Pong
		c.Send(&Signal{
			Type:      SignalPong,
			SenderID:  c.config.Node.ID,
			ReceiverID: signal.SenderID,
			Timestamp: time.Now(),
		})
		return
	case SignalPong:
		// 收到 Pong，不需要特殊处理
		return
	}

	// 调用注册的处理函数
	c.mu.RLock()
	handlers, exists := c.handlers[signal.Type]
	c.mu.RUnlock()

	if exists {
		for _, handler := range handlers {
			handler(signal)
		}
	}
}

// Send 发送信令消息
func (c *SignalingClient) Send(signal *Signal) {
	// 设置发送者 ID
	if signal.SenderID == "" {
		signal.SenderID = c.config.Node.ID
	}

	// 设置时间戳
	if signal.Timestamp.IsZero() {
		signal.Timestamp = time.Now()
	}

	// 发送信令消息
	c.sendCh <- signal
}

// RegisterHandler 注册信令处理函数
func (c *SignalingClient) RegisterHandler(signalType SignalType, handler SignalHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.handlers[signalType] = append(c.handlers[signalType], handler)
}

// IsConnected 检查是否已连接
func (c *SignalingClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// RequestConnect 请求连接到对等节点
func (c *SignalingClient) RequestConnect(peerID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("未连接到信令服务器")
	}

	// 发送连接请求
	c.Send(&Signal{
		Type:      SignalConnect,
		ReceiverID: peerID,
		Payload:   map[string]interface{}{
			"natType":     c.natInfo.Type.String(),
			"externalIP":  c.natInfo.ExternalIP.String(),
			"externalPort": c.natInfo.ExternalPort,
		},
	})

	return nil
}

// RequestRelay 请求中继连接
func (c *SignalingClient) RequestRelay(peerID string) error {
	if !c.IsConnected() {
		return fmt.Errorf("未连接到信令服务器")
	}

	// 发送中继请求
	c.Send(&Signal{
		Type:      SignalRelayRequest,
		ReceiverID: peerID,
	})

	return nil
}

// SendOffer 发送 Offer
func (c *SignalingClient) SendOffer(peerID string, offer interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("未连接到信令服务器")
	}

	// 发送 Offer
	c.Send(&Signal{
		Type:      SignalOffer,
		ReceiverID: peerID,
		Payload:   offer,
	})

	return nil
}

// SendAnswer 发送 Answer
func (c *SignalingClient) SendAnswer(peerID string, answer interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("未连接到信令服务器")
	}

	// 发送 Answer
	c.Send(&Signal{
		Type:      SignalAnswer,
		ReceiverID: peerID,
		Payload:   answer,
	})

	return nil
}

// SendICECandidate 发送 ICE 候选
func (c *SignalingClient) SendICECandidate(peerID string, candidate interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("未连接到信令服务器")
	}

	// 发送 ICE 候选
	c.Send(&Signal{
		Type:      SignalICECandidate,
		ReceiverID: peerID,
		Payload:   candidate,
	})

	return nil
}
