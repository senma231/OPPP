package forward

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// ForwardRule 表示一个端口转发规则
type ForwardRule struct {
	ID          string
	Protocol    string // tcp, udp
	SrcPort     int
	DstHost     string
	DstPort     int
	Description string
	Enabled     bool
	Stats       *ForwardStats
}

// ForwardStats 存储转发统计信息
type ForwardStats struct {
	BytesSent     uint64
	BytesReceived uint64
	Connections   uint64
	StartTime     time.Time
	mu            sync.Mutex
}

// NewForwardStats 创建一个新的转发统计信息
func NewForwardStats() *ForwardStats {
	return &ForwardStats{
		StartTime: time.Now(),
	}
}

// AddBytesSent 增加发送字节数
func (s *ForwardStats) AddBytesSent(n uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BytesSent += n
}

// AddBytesReceived 增加接收字节数
func (s *ForwardStats) AddBytesReceived(n uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BytesReceived += n
}

// IncrementConnections 增加连接数
func (s *ForwardStats) IncrementConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Connections++
}

// Forwarder 端口转发器
type Forwarder struct {
	rules     map[string]*ForwardRule
	listeners map[string]net.Listener
	mu        sync.RWMutex
	done      chan struct{}
}

// NewForwarder 创建一个新的端口转发器
func NewForwarder() *Forwarder {
	return &Forwarder{
		rules:     make(map[string]*ForwardRule),
		listeners: make(map[string]net.Listener),
		done:      make(chan struct{}),
	}
}

// AddRule 添加一个转发规则
func (f *Forwarder) AddRule(rule *ForwardRule) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 检查端口是否已被使用
	if _, exists := f.rules[rule.ID]; exists {
		return fmt.Errorf("规则 ID %s 已存在", rule.ID)
	}

	// 初始化统计信息
	if rule.Stats == nil {
		rule.Stats = NewForwardStats()
	}

	f.rules[rule.ID] = rule

	// 如果规则已启用，立即启动转发
	if rule.Enabled {
		return f.startForwarding(rule)
	}

	return nil
}

// RemoveRule 移除一个转发规则
func (f *Forwarder) RemoveRule(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	rule, exists := f.rules[id]
	if !exists {
		return fmt.Errorf("规则 ID %s 不存在", id)
	}

	// 如果规则已启用，先停止转发
	if rule.Enabled {
		if err := f.stopForwarding(rule); err != nil {
			return err
		}
	}

	delete(f.rules, id)
	return nil
}

// EnableRule 启用一个转发规则
func (f *Forwarder) EnableRule(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	rule, exists := f.rules[id]
	if !exists {
		return fmt.Errorf("规则 ID %s 不存在", id)
	}

	if rule.Enabled {
		return nil // 已经启用，无需操作
	}

	rule.Enabled = true
	return f.startForwarding(rule)
}

// DisableRule 禁用一个转发规则
func (f *Forwarder) DisableRule(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	rule, exists := f.rules[id]
	if !exists {
		return fmt.Errorf("规则 ID %s 不存在", id)
	}

	if !rule.Enabled {
		return nil // 已经禁用，无需操作
	}

	rule.Enabled = false
	return f.stopForwarding(rule)
}

// GetRule 获取一个转发规则
func (f *Forwarder) GetRule(id string) (*ForwardRule, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	rule, exists := f.rules[id]
	if !exists {
		return nil, fmt.Errorf("规则 ID %s 不存在", id)
	}

	return rule, nil
}

// ListRules 列出所有转发规则
func (f *Forwarder) ListRules() []*ForwardRule {
	f.mu.RLock()
	defer f.mu.RUnlock()

	rules := make([]*ForwardRule, 0, len(f.rules))
	for _, rule := range f.rules {
		rules = append(rules, rule)
	}

	return rules
}

// Close 关闭转发器
func (f *Forwarder) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	close(f.done)

	// 停止所有转发
	for _, rule := range f.rules {
		if rule.Enabled {
			if err := f.stopForwarding(rule); err != nil {
				return err
			}
		}
	}

	return nil
}

// startForwarding 启动一个规则的转发
func (f *Forwarder) startForwarding(rule *ForwardRule) error {
	// 根据协议类型启动不同的转发
	switch rule.Protocol {
	case "tcp":
		return f.startTCPForwarding(rule)
	case "udp":
		return f.startUDPForwarding(rule)
	default:
		return fmt.Errorf("不支持的协议类型: %s", rule.Protocol)
	}
}

// stopForwarding 停止一个规则的转发
func (f *Forwarder) stopForwarding(rule *ForwardRule) error {
	// 根据协议类型停止不同的转发
	switch rule.Protocol {
	case "tcp":
		return f.stopTCPForwarding(rule)
	case "udp":
		return f.stopUDPForwarding(rule)
	default:
		return fmt.Errorf("不支持的协议类型: %s", rule.Protocol)
	}
}

// startTCPForwarding 启动 TCP 转发
func (f *Forwarder) startTCPForwarding(rule *ForwardRule) error {
	// 监听本地端口
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", rule.SrcPort))
	if err != nil {
		return fmt.Errorf("监听端口 %d 失败: %w", rule.SrcPort, err)
	}

	f.listeners[rule.ID] = listener

	// 启动 goroutine 处理连接
	go func() {
		for {
			select {
			case <-f.done:
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					// 检查是否是因为关闭监听器导致的错误
					select {
					case <-f.done:
						return
					default:
						// TODO: 记录错误日志
						continue
					}
				}

				// 增加连接计数
				rule.Stats.IncrementConnections()

				// 启动 goroutine 处理连接
				go f.handleTCPConnection(conn, rule)
			}
		}
	}()

	return nil
}

// stopTCPForwarding 停止 TCP 转发
func (f *Forwarder) stopTCPForwarding(rule *ForwardRule) error {
	listener, exists := f.listeners[rule.ID]
	if !exists {
		return nil // 没有监听器，无需操作
	}

	if err := listener.Close(); err != nil {
		return fmt.Errorf("关闭监听器失败: %w", err)
	}

	delete(f.listeners, rule.ID)
	return nil
}

// handleTCPConnection 处理 TCP 连接
func (f *Forwarder) handleTCPConnection(clientConn net.Conn, rule *ForwardRule) {
	defer clientConn.Close()

	// 连接目标服务器
	targetConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", rule.DstHost, rule.DstPort))
	if err != nil {
		// TODO: 记录错误日志
		return
	}
	defer targetConn.Close()

	// 双向转发数据
	var wg sync.WaitGroup
	wg.Add(2)

	// 客户端 -> 目标服务器
	go func() {
		defer wg.Done()
		n, err := io.Copy(targetConn, clientConn)
		if err != nil {
			// TODO: 记录错误日志
		}
		rule.Stats.AddBytesSent(uint64(n))
	}()

	// 目标服务器 -> 客户端
	go func() {
		defer wg.Done()
		n, err := io.Copy(clientConn, targetConn)
		if err != nil {
			// TODO: 记录错误日志
		}
		rule.Stats.AddBytesReceived(uint64(n))
	}()

	wg.Wait()
}

// startUDPForwarding 启动 UDP 转发
func (f *Forwarder) startUDPForwarding(rule *ForwardRule) error {
	// TODO: 实现 UDP 转发
	return fmt.Errorf("UDP 转发尚未实现")
}

// stopUDPForwarding 停止 UDP 转发
func (f *Forwarder) stopUDPForwarding(rule *ForwardRule) error {
	// TODO: 实现停止 UDP 转发
	return fmt.Errorf("停止 UDP 转发尚未实现")
}
