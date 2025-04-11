package stats

import (
	"sync"
	"time"
)

// TrafficStats 流量统计
type TrafficStats struct {
	// 总发送字节数
	TotalSent int64
	// 总接收字节数
	TotalReceived int64
	// 每秒发送字节数
	BytesSentPerSecond int64
	// 每秒接收字节数
	BytesReceivedPerSecond int64
	// 连接数
	Connections int
	// 连接时长（秒）
	ConnectionTime int64
	// 最后更新时间
	LastUpdated time.Time
	
	// 窗口大小
	windowSize time.Duration
	// 窗口内发送字节数
	sentInWindow int64
	// 窗口内接收字节数
	receivedInWindow int64
	// 窗口开始时间
	windowStart time.Time
	
	mu sync.Mutex
}

// NewTrafficStats 创建流量统计
func NewTrafficStats() *TrafficStats {
	now := time.Now()
	return &TrafficStats{
		windowSize:  time.Second,
		windowStart: now,
		LastUpdated: now,
	}
}

// AddSent 添加发送字节数
func (s *TrafficStats) AddSent(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	s.TotalSent += bytes
	
	// 更新窗口
	if now.Sub(s.windowStart) >= s.windowSize {
		// 计算每秒字节数
		duration := now.Sub(s.windowStart).Seconds()
		if duration > 0 {
			s.BytesSentPerSecond = int64(float64(s.sentInWindow) / duration)
		}
		
		// 重置窗口
		s.sentInWindow = bytes
		s.windowStart = now
	} else {
		// 累加窗口内字节数
		s.sentInWindow += bytes
	}
	
	s.LastUpdated = now
}

// AddReceived 添加接收字节数
func (s *TrafficStats) AddReceived(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	s.TotalReceived += bytes
	
	// 更新窗口
	if now.Sub(s.windowStart) >= s.windowSize {
		// 计算每秒字节数
		duration := now.Sub(s.windowStart).Seconds()
		if duration > 0 {
			s.BytesReceivedPerSecond = int64(float64(s.receivedInWindow) / duration)
		}
		
		// 重置窗口
		s.receivedInWindow = bytes
		s.windowStart = now
	} else {
		// 累加窗口内字节数
		s.receivedInWindow += bytes
	}
	
	s.LastUpdated = now
}

// AddConnection 添加连接
func (s *TrafficStats) AddConnection() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.Connections++
	s.LastUpdated = time.Now()
}

// RemoveConnection 移除连接
func (s *TrafficStats) RemoveConnection() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.Connections > 0 {
		s.Connections--
	}
	s.LastUpdated = time.Now()
}

// UpdateConnectionTime 更新连接时长
func (s *TrafficStats) UpdateConnectionTime(seconds int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.ConnectionTime = seconds
	s.LastUpdated = time.Now()
}

// GetStats 获取统计信息
func (s *TrafficStats) GetStats() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	return map[string]interface{}{
		"totalSent":             s.TotalSent,
		"totalReceived":         s.TotalReceived,
		"bytesSentPerSecond":    s.BytesSentPerSecond,
		"bytesReceivedPerSecond": s.BytesReceivedPerSecond,
		"connections":           s.Connections,
		"connectionTime":        s.ConnectionTime,
		"lastUpdated":           s.LastUpdated,
	}
}

// Reset 重置统计信息
func (s *TrafficStats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	s.TotalSent = 0
	s.TotalReceived = 0
	s.BytesSentPerSecond = 0
	s.BytesReceivedPerSecond = 0
	s.Connections = 0
	s.ConnectionTime = 0
	s.sentInWindow = 0
	s.receivedInWindow = 0
	s.windowStart = now
	s.LastUpdated = now
}
