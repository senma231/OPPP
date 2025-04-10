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
	rules        map[string]*ForwardRule
	listeners    map[string]net.Listener
	udpListeners map[string]*net.UDPConn
	mu           sync.RWMutex
	done         chan struct{}
}

// NewForwarder 创建一个新的端口转发器
func NewForwarder() *Forwarder {
	return &Forwarder{
		rules:        make(map[string]*ForwardRule),
		listeners:    make(map[string]net.Listener),
		udpListeners: make(map[string]*net.UDPConn),
		done:         make(chan struct{}),
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
	targetAddr := net.JoinHostPort(rule.DstHost, fmt.Sprintf("%d", rule.DstPort))
	targetConn, err := net.Dial("tcp", targetAddr)
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
	// 监听本地 UDP 端口
	listener, err := net.ListenUDP("udp", &net.UDPAddr{Port: rule.SrcPort})
	if err != nil {
		return fmt.Errorf("监听 UDP 端口 %d 失败: %w", rule.SrcPort, err)
	}

	// 存储监听器
	f.mu.Lock()
	f.udpListeners[rule.ID] = listener
	f.mu.Unlock()

	// 创建会话映射
	sessions := make(map[string]*udpSession)
	sessionsMutex := &sync.RWMutex{}

	// 启动 goroutine 处理数据
	go func() {
		buf := make([]byte, 65507) // UDP 最大包大小
		for {
			select {
			case <-f.done:
				return
			default:
				// 设置读取超时
				listener.SetReadDeadline(time.Now().Add(1 * time.Second))

				// 读取数据
				n, clientAddr, err := listener.ReadFromUDP(buf)
				if err != nil {
					// 检查是否是超时错误
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}

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

				// 获取或创建会话
				clientKey := clientAddr.String()
				sessionsMutex.RLock()
				session, exists := sessions[clientKey]
				sessionsMutex.RUnlock()

				if !exists {
					// 创建到目标的连接
					targetAddrStr := net.JoinHostPort(rule.DstHost, fmt.Sprintf("%d", rule.DstPort))
					targetAddr, err := net.ResolveUDPAddr("udp", targetAddrStr)
					if err != nil {
						// TODO: 记录错误日志
						continue
					}

					targetConn, err := net.DialUDP("udp", nil, targetAddr)
					if err != nil {
						// TODO: 记录错误日志
						continue
					}

					// 创建新会话
					session = &udpSession{
						clientAddr: clientAddr,
						targetConn: targetConn,
						lastActive: time.Now(),
					}

					sessionsMutex.Lock()
					sessions[clientKey] = session
					sessionsMutex.Unlock()

					// 启动 goroutine 处理目标到客户端的数据
					go func() {
						targetBuf := make([]byte, 65507)
						for {
							// 设置读取超时
							targetConn.SetReadDeadline(time.Now().Add(30 * time.Second))

							// 读取数据
							n, err := targetConn.Read(targetBuf)
							if err != nil {
								// 检查是否是超时错误
								if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
									// 检查会话是否过期
									sessionsMutex.RLock()
									lastActive := session.lastActive
									sessionsMutex.RUnlock()

									if time.Since(lastActive) > 60*time.Second {
										// 关闭连接
										targetConn.Close()

										// 移除会话
										sessionsMutex.Lock()
										delete(sessions, clientKey)
										sessionsMutex.Unlock()

										return
									}

									continue
								}

								// 其他错误
								// TODO: 记录错误日志

								// 关闭连接
								targetConn.Close()

								// 移除会话
								sessionsMutex.Lock()
								delete(sessions, clientKey)
								sessionsMutex.Unlock()

								return
							}

							// 发送数据到客户端
							_, err = listener.WriteToUDP(targetBuf[:n], clientAddr)
							if err != nil {
								// TODO: 记录错误日志
								continue
							}

							// 更新统计信息
							rule.Stats.AddBytesReceived(uint64(n))

							// 更新最后活动时间
							sessionsMutex.Lock()
							session.lastActive = time.Now()
							sessionsMutex.Unlock()
						}
					}()
				} else {
					// 更新最后活动时间
					sessionsMutex.Lock()
					session.lastActive = time.Now()
					sessionsMutex.Unlock()
				}

				// 发送数据到目标
				_, err = session.targetConn.Write(buf[:n])
				if err != nil {
					// TODO: 记录错误日志
					continue
				}

				// 更新统计信息
				rule.Stats.AddBytesSent(uint64(n))
			}
		}
	}()

	return nil
}

// stopUDPForwarding 停止 UDP 转发
func (f *Forwarder) stopUDPForwarding(rule *ForwardRule) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	listener, exists := f.udpListeners[rule.ID]
	if !exists {
		return nil // 没有监听器，无需操作
	}

	if err := listener.Close(); err != nil {
		return fmt.Errorf("关闭 UDP 监听器失败: %w", err)
	}

	delete(f.udpListeners, rule.ID)
	return nil
}

// udpSession UDP 会话
type udpSession struct {
	clientAddr *net.UDPAddr
	targetConn *net.UDPConn
	lastActive time.Time
}
