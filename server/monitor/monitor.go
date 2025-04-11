package monitor

import (
	"sync"
	"time"
)

// Monitor 监控器
type Monitor struct {
	// 设备状态
	deviceStatus map[uint]string
	// 应用状态
	appStatus map[uint]string
	// 连接状态
	connectionStatus map[uint]string
	// 订阅者
	subscribers map[string]chan Event
	mu          sync.RWMutex
}

// Event 事件
type Event struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// NewMonitor 创建监控器
func NewMonitor() *Monitor {
	return &Monitor{
		deviceStatus:     make(map[uint]string),
		appStatus:        make(map[uint]string),
		connectionStatus: make(map[uint]string),
		subscribers:      make(map[string]chan Event),
	}
}

// Subscribe 订阅事件
func (m *Monitor) Subscribe(id string) chan Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	ch := make(chan Event, 100)
	m.subscribers[id] = ch
	
	return ch
}

// Unsubscribe 取消订阅
func (m *Monitor) Unsubscribe(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if ch, exists := m.subscribers[id]; exists {
		close(ch)
		delete(m.subscribers, id)
	}
}

// UpdateDeviceStatus 更新设备状态
func (m *Monitor) UpdateDeviceStatus(deviceID uint, status string) {
	m.mu.Lock()
	oldStatus, exists := m.deviceStatus[deviceID]
	m.deviceStatus[deviceID] = status
	m.mu.Unlock()
	
	// 如果状态发生变化，发布事件
	if !exists || oldStatus != status {
		m.publishEvent("device_status", map[string]interface{}{
			"deviceID": deviceID,
			"status":   status,
		})
	}
}

// UpdateAppStatus 更新应用状态
func (m *Monitor) UpdateAppStatus(appID uint, status string) {
	m.mu.Lock()
	oldStatus, exists := m.appStatus[appID]
	m.appStatus[appID] = status
	m.mu.Unlock()
	
	// 如果状态发生变化，发布事件
	if !exists || oldStatus != status {
		m.publishEvent("app_status", map[string]interface{}{
			"appID":  appID,
			"status": status,
		})
	}
}

// UpdateConnectionStatus 更新连接状态
func (m *Monitor) UpdateConnectionStatus(connectionID uint, status string) {
	m.mu.Lock()
	oldStatus, exists := m.connectionStatus[connectionID]
	m.connectionStatus[connectionID] = status
	m.mu.Unlock()
	
	// 如果状态发生变化，发布事件
	if !exists || oldStatus != status {
		m.publishEvent("connection_status", map[string]interface{}{
			"connectionID": connectionID,
			"status":       status,
		})
	}
}

// publishEvent 发布事件
func (m *Monitor) publishEvent(eventType string, data interface{}) {
	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, ch := range m.subscribers {
		select {
		case ch <- event:
			// 事件已发送
		default:
			// 通道已满，丢弃事件
		}
	}
}

// Start 启动监控
func (m *Monitor) Start() {
	// 定期检查设备状态
	go m.checkDeviceStatus()
	
	// 定期检查应用状态
	go m.checkAppStatus()
	
	// 定期检查连接状态
	go m.checkConnectionStatus()
}

// checkDeviceStatus 检查设备状态
func (m *Monitor) checkDeviceStatus() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		// 这里应该从数据库或其他数据源获取设备状态
		// 为了简化，这里只是示例
		
		// 更新设备状态
		// m.UpdateDeviceStatus(deviceID, status)
	}
}

// checkAppStatus 检查应用状态
func (m *Monitor) checkAppStatus() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		// 这里应该从数据库或其他数据源获取应用状态
		// 为了简化，这里只是示例
		
		// 更新应用状态
		// m.UpdateAppStatus(appID, status)
	}
}

// checkConnectionStatus 检查连接状态
func (m *Monitor) checkConnectionStatus() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		// 这里应该从数据库或其他数据源获取连接状态
		// 为了简化，这里只是示例
		
		// 更新连接状态
		// m.UpdateConnectionStatus(connectionID, status)
	}
}
