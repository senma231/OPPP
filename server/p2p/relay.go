package p2p

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/senma231/p3/common/logger"
	"github.com/senma231/p3/server/config"
)

// RelaySession 中继会话
type RelaySession struct {
	ID            string
	SourceID      string
	TargetID      string
	SourceConn    net.Conn
	TargetConn    net.Conn
	BytesSent     uint64
	BytesReceived uint64
	CreatedAt     time.Time
	LastActiveAt  time.Time
	mu            sync.Mutex
}

// RelayServer 中继服务器
type RelayServer struct {
	config     *config.Config
	coordinator *Coordinator
	sessions   map[string]*RelaySession
	listener   net.Listener
	running    bool
	mu         sync.RWMutex
	stopCh     chan struct{}
}

// NewRelayServer 创建中继服务器
func NewRelayServer(cfg *config.Config, coordinator *Coordinator) *RelayServer {
	return &RelayServer{
		config:     cfg,
		coordinator: coordinator,
		sessions:   make(map[string]*RelaySession),
		stopCh:     make(chan struct{}),
	}
}

// Start 启动中继服务器
func (s *RelayServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("中继服务器已在运行")
	}

	// 创建监听器
	addr := fmt.Sprintf("%s:%d", s.config.Relay.Host, s.config.Relay.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("创建监听器失败: %w", err)
	}
	s.listener = listener

	s.running = true
	logger.Info("中继服务器已启动，监听地址: %s", addr)

	// 启动接收协程
	go s.acceptLoop()

	// 启动清理协程
	go s.cleanupLoop()

	return nil
}

// Stop 停止中继服务器
func (s *RelayServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// 关闭监听器
	if s.listener != nil {
		s.listener.Close()
	}

	// 发送停止信号
	close(s.stopCh)

	// 关闭所有会话
	for _, session := range s.sessions {
		s.closeSession(session)
	}

	s.running = false
	logger.Info("中继服务器已停止")
	return nil
}

// acceptLoop 接受连接循环
func (s *RelayServer) acceptLoop() {
	for {
		// 接受连接
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
				return
			default:
				logger.Error("接受连接失败: %v", err)
				time.Sleep(time.Second)
				continue
			}
		}

		// 处理连接
		go s.handleConnection(conn)
	}
}

// handleConnection 处理连接
func (s *RelayServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 设置超时
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// 读取请求
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		logger.Error("读取请求失败: %v", err)
		return
	}

	// 解析请求
	request := string(buffer[:n])
	if len(request) < 7 || request[:6] != "RELAY " {
		logger.Error("无效的请求: %s", request)
		conn.Write([]byte("ERROR: Invalid request"))
		return
	}

	// 提取目标节点 ID
	targetID := request[6:]
	if targetID == "" {
		logger.Error("目标节点 ID 为空")
		conn.Write([]byte("ERROR: Empty target ID"))
		return
	}

	// 获取源节点 ID（通过认证信息）
	sourceID := "unknown" // 实际应该从认证信息中获取

	// 检查目标节点是否在线
	targetPeer, err := s.coordinator.GetPeerInfo(targetID)
	if err != nil {
		logger.Error("目标节点不存在或不在线: %v", err)
		conn.Write([]byte("ERROR: Target node not found or offline"))
		return
	}

	// 连接到目标节点
	targetAddr := fmt.Sprintf("%s:%d", targetPeer.ExternalIP.String(), targetPeer.ExternalPort)
	targetConn, err := net.DialTimeout("tcp", targetAddr, 5*time.Second)
	if err != nil {
		logger.Error("连接目标节点失败: %v", err)
		conn.Write([]byte("ERROR: Failed to connect to target node"))
		return
	}

	// 创建会话
	sessionID := fmt.Sprintf("%s-%s-%d", sourceID, targetID, time.Now().UnixNano())
	session := &RelaySession{
		ID:            sessionID,
		SourceID:      sourceID,
		TargetID:      targetID,
		SourceConn:    conn,
		TargetConn:    targetConn,
		CreatedAt:     time.Now(),
		LastActiveAt:  time.Now(),
	}

	// 添加会话
	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	// 发送成功响应
	conn.Write([]byte("OK"))

	// 清除超时
	conn.SetDeadline(time.Time{})
	targetConn.SetDeadline(time.Time{})

	// 启动中继
	go s.relay(session)

	logger.Info("中继会话已创建: %s -> %s", sourceID, targetID)
}

// relay 中继数据
func (s *RelayServer) relay(session *RelaySession) {
	// 创建同步组
	var wg sync.WaitGroup
	wg.Add(2)

	// 源 -> 目标
	go func() {
		defer wg.Done()
		s.copyData(session, session.TargetConn, session.SourceConn)
	}()

	// 目标 -> 源
	go func() {
		defer wg.Done()
		s.copyData(session, session.SourceConn, session.TargetConn)
	}()

	// 等待两个方向的数据传输完成
	wg.Wait()

	// 关闭会话
	s.mu.Lock()
	delete(s.sessions, session.ID)
	s.mu.Unlock()

	s.closeSession(session)
	logger.Info("中继会话已关闭: %s -> %s", session.SourceID, session.TargetID)
}

// copyData 复制数据
func (s *RelayServer) copyData(session *RelaySession, dst, src net.Conn) {
	buffer := make([]byte, 4096)
	for {
		// 读取数据
		n, err := src.Read(buffer)
		if err != nil {
			if err != io.EOF {
				logger.Error("读取数据失败: %v", err)
			}
			break
		}

		// 写入数据
		_, err = dst.Write(buffer[:n])
		if err != nil {
			logger.Error("写入数据失败: %v", err)
			break
		}

		// 更新统计信息
		session.mu.Lock()
		if src == session.SourceConn {
			session.BytesSent += uint64(n)
		} else {
			session.BytesReceived += uint64(n)
		}
		session.LastActiveAt = time.Now()
		session.mu.Unlock()
	}
}

// closeSession 关闭会话
func (s *RelayServer) closeSession(session *RelaySession) {
	if session.SourceConn != nil {
		session.SourceConn.Close()
	}
	if session.TargetConn != nil {
		session.TargetConn.Close()
	}
}

// cleanupLoop 清理循环
func (s *RelayServer) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanupInactiveSessions()
		}
	}
}

// cleanupInactiveSessions 清理不活跃的会话
func (s *RelayServer) cleanupInactiveSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		session.mu.Lock()
		inactive := now.Sub(session.LastActiveAt) > 5*time.Minute
		session.mu.Unlock()

		if inactive {
			logger.Info("清理不活跃的会话: %s", id)
			s.closeSession(session)
			delete(s.sessions, id)
		}
	}
}

// GetSessionCount 获取会话数量
func (s *RelayServer) GetSessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// GetTotalBytesTransferred 获取总传输字节数
func (s *RelayServer) GetTotalBytesTransferred() (uint64, uint64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalSent, totalReceived uint64
	for _, session := range s.sessions {
		session.mu.Lock()
		totalSent += session.BytesSent
		totalReceived += session.BytesReceived
		session.mu.Unlock()
	}

	return totalSent, totalReceived
}
