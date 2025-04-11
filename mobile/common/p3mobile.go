// Package p3mobile 提供 P3 移动平台的共享代码
package p3mobile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// P3Client P3 移动客户端
type P3Client struct {
	serverAddress string
	nodeID        string
	token         string
	connected     bool
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	eventCallback EventCallback
}

// Config P3 客户端配置
type Config struct {
	ServerAddress string
	NodeID        string
	Token         string
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string
	Protocol    string
	SrcPort     int
	PeerNode    string
	DstPort     int
	DstHost     string
	Description string
	AutoStart   bool
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID        string
	Name      string
	Status    string
	IP        string
	NATType   string
	LastSeen  string
	CreatedAt string
}

// AppInfo 应用信息
type AppInfo struct {
	ID          string
	Name        string
	Protocol    string
	SrcPort     int
	PeerNode    string
	DstPort     int
	DstHost     string
	Description string
	Status      string
	AutoStart   bool
	CreatedAt   string
	UpdatedAt   string
}

// NetworkStatus 网络状态
type NetworkStatus struct {
	ExternalIP   string
	NATType      string
	UPnPAvailable bool
	ConnectedPeers int
	Bandwidth     struct {
		Upload   int64
		Download int64
	}
}

// EventType 事件类型
type EventType int

const (
	// EventConnected 连接成功事件
	EventConnected EventType = iota
	// EventDisconnected 连接断开事件
	EventDisconnected
	// EventError 错误事件
	EventError
	// EventAppStarted 应用启动事件
	EventAppStarted
	// EventAppStopped 应用停止事件
	EventAppStopped
	// EventPeerConnected 对等节点连接事件
	EventPeerConnected
	// EventPeerDisconnected 对等节点断开事件
	EventPeerDisconnected
)

// Event 事件
type Event struct {
	Type    EventType
	Message string
	Data    string
}

// EventCallback 事件回调接口
type EventCallback interface {
	OnEvent(event Event)
}

// NewP3Client 创建 P3 客户端
func NewP3Client(config Config) *P3Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &P3Client{
		serverAddress: config.ServerAddress,
		nodeID:        config.NodeID,
		token:         config.Token,
		connected:     false,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// SetEventCallback 设置事件回调
func (c *P3Client) SetEventCallback(callback EventCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventCallback = callback
}

// Connect 连接到服务器
func (c *P3Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return errors.New("已连接到服务器")
	}

	// 模拟连接过程
	time.Sleep(1 * time.Second)

	c.connected = true
	c.emitEvent(EventConnected, "已连接到服务器", "")
	return nil
}

// Disconnect 断开连接
func (c *P3Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("未连接到服务器")
	}

	c.cancel()
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel
	c.connected = false
	c.emitEvent(EventDisconnected, "已断开连接", "")
	return nil
}

// IsConnected 检查是否已连接
func (c *P3Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// GetDevices 获取设备列表
func (c *P3Client) GetDevices() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("未连接到服务器")
	}

	// 模拟获取设备列表
	devices := []DeviceInfo{
		{
			ID:        "device-1",
			Name:      "My PC",
			Status:    "online",
			IP:        "192.168.1.100",
			NATType:   "NAT2",
			LastSeen:  time.Now().Format(time.RFC3339),
			CreatedAt: time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
		},
		{
			ID:        "device-2",
			Name:      "My Server",
			Status:    "offline",
			IP:        "10.0.0.1",
			NATType:   "NAT1",
			LastSeen:  time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
			CreatedAt: time.Now().AddDate(0, -2, 0).Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(devices)
	if err != nil {
		return "", fmt.Errorf("序列化设备列表失败: %w", err)
	}

	return string(data), nil
}

// GetApps 获取应用列表
func (c *P3Client) GetApps() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("未连接到服务器")
	}

	// 模拟获取应用列表
	apps := []AppInfo{
		{
			ID:          "app-1",
			Name:        "SSH",
			Protocol:    "tcp",
			SrcPort:     12222,
			PeerNode:    "device-2",
			DstPort:     22,
			DstHost:     "localhost",
			Description: "SSH access",
			Status:      "running",
			AutoStart:   true,
			CreatedAt:   time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
		{
			ID:          "app-2",
			Name:        "RDP",
			Protocol:    "tcp",
			SrcPort:     13389,
			PeerNode:    "device-2",
			DstPort:     3389,
			DstHost:     "localhost",
			Description: "Remote Desktop",
			Status:      "stopped",
			AutoStart:   false,
			CreatedAt:   time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(apps)
	if err != nil {
		return "", fmt.Errorf("序列化应用列表失败: %w", err)
	}

	return string(data), nil
}

// AddApp 添加应用
func (c *P3Client) AddApp(appConfigJSON string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("未连接到服务器")
	}

	var appConfig AppConfig
	if err := json.Unmarshal([]byte(appConfigJSON), &appConfig); err != nil {
		return "", fmt.Errorf("解析应用配置失败: %w", err)
	}

	// 模拟添加应用
	app := AppInfo{
		ID:          fmt.Sprintf("app-%d", time.Now().Unix()),
		Name:        appConfig.Name,
		Protocol:    appConfig.Protocol,
		SrcPort:     appConfig.SrcPort,
		PeerNode:    appConfig.PeerNode,
		DstPort:     appConfig.DstPort,
		DstHost:     appConfig.DstHost,
		Description: appConfig.Description,
		Status:      "stopped",
		AutoStart:   appConfig.AutoStart,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(app)
	if err != nil {
		return "", fmt.Errorf("序列化应用信息失败: %w", err)
	}

	return string(data), nil
}

// RemoveApp 删除应用
func (c *P3Client) RemoveApp(appID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("未连接到服务器")
	}

	// 模拟删除应用
	return nil
}

// StartApp 启动应用
func (c *P3Client) StartApp(appID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("未连接到服务器")
	}

	// 模拟启动应用
	c.emitEvent(EventAppStarted, fmt.Sprintf("应用 %s 已启动", appID), appID)
	return nil
}

// StopApp 停止应用
func (c *P3Client) StopApp(appID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return errors.New("未连接到服务器")
	}

	// 模拟停止应用
	c.emitEvent(EventAppStopped, fmt.Sprintf("应用 %s 已停止", appID), appID)
	return nil
}

// GetNetworkStatus 获取网络状态
func (c *P3Client) GetNetworkStatus() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("未连接到服务器")
	}

	// 模拟获取网络状态
	status := NetworkStatus{
		ExternalIP:     "203.0.113.1",
		NATType:        "NAT2",
		UPnPAvailable:  true,
		ConnectedPeers: 2,
	}
	status.Bandwidth.Upload = 1024 * 1024
	status.Bandwidth.Download = 2048 * 1024

	data, err := json.Marshal(status)
	if err != nil {
		return "", fmt.Errorf("序列化网络状态失败: %w", err)
	}

	return string(data), nil
}

// DetectNAT 检测 NAT 类型
func (c *P3Client) DetectNAT() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("未连接到服务器")
	}

	// 模拟检测 NAT 类型
	time.Sleep(2 * time.Second)
	return "NAT2", nil
}

// TestConnection 测试与节点的连接
func (c *P3Client) TestConnection(peerNode string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return "", errors.New("未连接到服务器")
	}

	// 模拟测试连接
	time.Sleep(2 * time.Second)
	result := map[string]interface{}{
		"success":      true,
		"latency":      50,
		"connection_type": "p2p",
		"nat_traversal": "direct",
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("序列化连接测试结果失败: %w", err)
	}

	return string(data), nil
}

// emitEvent 发送事件
func (c *P3Client) emitEvent(eventType EventType, message string, data string) {
	if c.eventCallback != nil {
		event := Event{
			Type:    eventType,
			Message: message,
			Data:    data,
		}
		c.eventCallback.OnEvent(event)
	}
}
