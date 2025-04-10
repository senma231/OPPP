package relay

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/senma231/p3/server/config"
	"github.com/senma231/p3/server/db"
)

// Session 中转会话
type Session struct {
	ID            string
	SourceConn    net.Conn
	TargetConn    net.Conn
	BytesSent     uint64
	BytesReceived uint64
	StartTime     time.Time
	LastActiveTime time.Time
	mu            sync.Mutex
}

// NewSession 创建中转会话
func NewSession(id string, sourceConn, targetConn net.Conn) *Session {
	now := time.Now()
	return &Session{
		ID:            id,
		SourceConn:    sourceConn,
		TargetConn:    targetConn,
		StartTime:     now,
		LastActiveTime: now,
	}
}

// Close 关闭会话
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.SourceConn != nil {
		s.SourceConn.Close()
		s.SourceConn = nil
	}

	if s.TargetConn != nil {
		s.TargetConn.Close()
		s.TargetConn = nil
	}
}

// UpdateStats 更新统计信息
func (s *Session) UpdateStats(bytesSent, bytesReceived uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.BytesSent += bytesSent
	s.BytesReceived += bytesReceived
	s.LastActiveTime = time.Now()
}

// Service 中转服务
type Service struct {
	config   *config.Config
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewService 创建中转服务
func NewService(cfg *config.Config) *Service {
	return &Service{
		config:   cfg,
		sessions: make(map[string]*Session),
	}
}

// CreateSession 创建中转会话
func (s *Service) CreateSession(sessionID string, sourceConn, targetConn net.Conn) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionID]; exists {
		return nil, errors.New("会话 ID 已存在")
	}

	session := NewSession(sessionID, sourceConn, targetConn)
	s.sessions[sessionID] = session

	// 启动数据转发
	go s.forwardData(session, sourceConn, targetConn, true)
	go s.forwardData(session, targetConn, sourceConn, false)

	return session, nil
}

// GetSession 获取中转会话
func (s *Service) GetSession(sessionID string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, errors.New("会话不存在")
	}

	return session, nil
}

// CloseSession 关闭中转会话
func (s *Service) CloseSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return errors.New("会话不存在")
	}

	session.Close()
	delete(s.sessions, sessionID)

	return nil
}

// GetAllSessions 获取所有中转会话
func (s *Service) GetAllSessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// CleanupInactiveSessions 清理不活跃的会话
func (s *Service) CleanupInactiveSessions(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if now.Sub(session.LastActiveTime) > timeout {
			session.Close()
			delete(s.sessions, id)
		}
	}
}

// forwardData 转发数据
func (s *Service) forwardData(session *Session, src, dst net.Conn, isSourceToTarget bool) {
	buffer := make([]byte, 4096)
	for {
		// 读取数据
		n, err := src.Read(buffer)
		if err != nil {
			if err != io.EOF {
				// TODO: 记录错误日志
			}
			break
		}

		// 写入数据
		_, err = dst.Write(buffer[:n])
		if err != nil {
			// TODO: 记录错误日志
			break
		}

		// 更新统计信息
		if isSourceToTarget {
			session.UpdateStats(uint64(n), 0)
		} else {
			session.UpdateStats(0, uint64(n))
		}
	}

	// 关闭会话
	s.CloseSession(session.ID)
}

// RecordRelayStats 记录中转统计信息
func (s *Service) RecordRelayStats(sessionID string, sourceDeviceID, targetDeviceID uint) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 记录源设备统计信息
	sourceStats := &db.Stats{
		DeviceID:       sourceDeviceID,
		BytesSent:      session.BytesSent,
		BytesReceived:  session.BytesReceived,
		Connections:    1,
		ConnectionTime: uint64(time.Since(session.StartTime).Seconds()),
	}

	if err := db.DB.Create(sourceStats).Error; err != nil {
		return fmt.Errorf("记录源设备统计信息失败: %w", err)
	}

	// 记录目标设备统计信息
	targetStats := &db.Stats{
		DeviceID:       targetDeviceID,
		BytesSent:      session.BytesReceived,
		BytesReceived:  session.BytesSent,
		Connections:    1,
		ConnectionTime: uint64(time.Since(session.StartTime).Seconds()),
	}

	if err := db.DB.Create(targetStats).Error; err != nil {
		return fmt.Errorf("记录目标设备统计信息失败: %w", err)
	}

	return nil
}
