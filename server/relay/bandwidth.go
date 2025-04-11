package relay

import (
	"sync"
	"time"
)

// BandwidthManager 带宽管理器
type BandwidthManager struct {
	// 每个节点的带宽限制 (bytes/s)
	nodeLimits map[string]int64
	// 每个会话的带宽限制 (bytes/s)
	sessionLimits map[string]int64
	// 每个节点的当前使用量
	nodeUsage map[string]*UsageCounter
	// 每个会话的当前使用量
	sessionUsage map[string]*UsageCounter
	mu           sync.RWMutex
}

// UsageCounter 使用量计数器
type UsageCounter struct {
	bytesTotal   int64
	bytesPerSec  int64
	lastUpdated  time.Time
	windowSize   time.Duration
	bytesInWindow int64
	mu           sync.Mutex
}

// NewBandwidthManager 创建带宽管理器
func NewBandwidthManager() *BandwidthManager {
	return &BandwidthManager{
		nodeLimits:     make(map[string]int64),
		sessionLimits:  make(map[string]int64),
		nodeUsage:      make(map[string]*UsageCounter),
		sessionUsage:   make(map[string]*UsageCounter),
	}
}

// SetNodeLimit 设置节点带宽限制
func (bm *BandwidthManager) SetNodeLimit(nodeID string, bytesPerSec int64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.nodeLimits[nodeID] = bytesPerSec
}

// SetSessionLimit 设置会话带宽限制
func (bm *BandwidthManager) SetSessionLimit(sessionID string, bytesPerSec int64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.sessionLimits[sessionID] = bytesPerSec
}

// GetNodeUsage 获取节点使用量
func (bm *BandwidthManager) GetNodeUsage(nodeID string) *UsageCounter {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	counter, exists := bm.nodeUsage[nodeID]
	if !exists {
		counter = &UsageCounter{
			lastUpdated: time.Now(),
			windowSize:  time.Second,
		}
		bm.nodeUsage[nodeID] = counter
	}
	
	return counter
}

// GetSessionUsage 获取会话使用量
func (bm *BandwidthManager) GetSessionUsage(sessionID string) *UsageCounter {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	counter, exists := bm.sessionUsage[sessionID]
	if !exists {
		counter = &UsageCounter{
			lastUpdated: time.Now(),
			windowSize:  time.Second,
		}
		bm.sessionUsage[sessionID] = counter
	}
	
	return counter
}

// AddBytes 添加字节数
func (uc *UsageCounter) AddBytes(bytes int64) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(uc.lastUpdated)
	
	// 更新总字节数
	uc.bytesTotal += bytes
	
	// 更新窗口内字节数
	if elapsed >= uc.windowSize {
		// 如果已经过了一个窗口，重置窗口内字节数
		uc.bytesInWindow = bytes
		uc.bytesPerSec = bytes
	} else {
		// 否则累加窗口内字节数
		uc.bytesInWindow += bytes
		// 计算每秒字节数
		uc.bytesPerSec = int64(float64(uc.bytesInWindow) / elapsed.Seconds())
	}
	
	uc.lastUpdated = now
}

// GetBytesPerSec 获取每秒字节数
func (uc *UsageCounter) GetBytesPerSec() int64 {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.bytesPerSec
}

// GetTotalBytes 获取总字节数
func (uc *UsageCounter) GetTotalBytes() int64 {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.bytesTotal
}

// CheckLimit 检查是否超过限制
func (bm *BandwidthManager) CheckLimit(nodeID, sessionID string, bytes int64) bool {
	bm.mu.RLock()
	nodeLimit, nodeExists := bm.nodeLimits[nodeID]
	sessionLimit, sessionExists := bm.sessionLimits[sessionID]
	bm.mu.RUnlock()
	
	// 获取当前使用量
	nodeUsage := bm.GetNodeUsage(nodeID)
	sessionUsage := bm.GetSessionUsage(sessionID)
	
	// 检查节点限制
	if nodeExists && nodeLimit > 0 {
		nodeBps := nodeUsage.GetBytesPerSec()
		if nodeBps+bytes > nodeLimit {
			return false
		}
	}
	
	// 检查会话限制
	if sessionExists && sessionLimit > 0 {
		sessionBps := sessionUsage.GetBytesPerSec()
		if sessionBps+bytes > sessionLimit {
			return false
		}
	}
	
	// 更新使用量
	nodeUsage.AddBytes(bytes)
	sessionUsage.AddBytes(bytes)
	
	return true
}
