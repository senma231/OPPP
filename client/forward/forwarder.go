package forward

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/senma231/p3/client/config"
	"github.com/senma231/p3/common/logger"
)

// Forwarder 转发器
type Forwarder struct {
	config     *config.AppConfig
	listener   net.Listener
	conn       net.Conn
	stopCh     chan struct{}
	wg         sync.WaitGroup
	stats      *Stats
	bufferSize int
	running    bool
	mu         sync.Mutex
}

// Stats 统计信息
type Stats struct {
	BytesSent       uint64
	BytesReceived   uint64
	Connections     uint64
	ConnectionTime  uint64
	LastActiveTime  time.Time
	mu              sync.Mutex
}

// NewForwarder 创建转发器
func NewForwarder(cfg *config.AppConfig, bufferSize int) *Forwarder {
	if bufferSize <= 0 {
		bufferSize = 4096
	}

	return &Forwarder{
		config:     cfg,
		stopCh:     make(chan struct{}),
		stats:      &Stats{LastActiveTime: time.Now()},
		bufferSize: bufferSize,
	}
}

// Start 启动转发器
func (f *Forwarder) Start() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.running {
		return fmt.Errorf("转发器已在运行")
	}

	// 创建监听器
	var err error
	listenAddr := fmt.Sprintf(":%d", f.config.SrcPort)
	f.listener, err = net.Listen(f.config.Protocol, listenAddr)
	if err != nil {
		return fmt.Errorf("创建监听器失败: %w", err)
	}

	f.running = true
	f.wg.Add(1)

	// 启动接收协程
	go f.acceptLoop()

	logger.Info("转发器已启动: %s -> %s:%d", listenAddr, f.config.DstHost, f.config.DstPort)
	return nil
}

// Stop 停止转发器
func (f *Forwarder) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return nil
	}

	// 关闭监听器
	if f.listener != nil {
		f.listener.Close()
	}

	// 关闭连接
	if f.conn != nil {
		f.conn.Close()
	}

	// 发送停止信号
	close(f.stopCh)

	// 等待所有协程退出
	f.wg.Wait()

	f.running = false
	logger.Info("转发器已停止: %s", f.config.Name)
	return nil
}

// IsRunning 检查转发器是否正在运行
func (f *Forwarder) IsRunning() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.running
}

// GetStats 获取统计信息
func (f *Forwarder) GetStats() *Stats {
	return f.stats
}

// acceptLoop 接受连接循环
func (f *Forwarder) acceptLoop() {
	defer f.wg.Done()

	for {
		select {
		case <-f.stopCh:
			return
		default:
			// 接受连接
			conn, err := f.listener.Accept()
			if err != nil {
				select {
				case <-f.stopCh:
					return
				default:
					logger.Error("接受连接失败: %v", err)
					time.Sleep(time.Second)
					continue
				}
			}

			// 处理连接
			f.wg.Add(1)
			go f.handleConnection(conn)
		}
	}
}

// handleConnection 处理连接
func (f *Forwarder) handleConnection(clientConn net.Conn) {
	defer f.wg.Done()
	defer clientConn.Close()

	// 更新统计信息
	f.stats.mu.Lock()
	f.stats.Connections++
	f.stats.LastActiveTime = time.Now()
	f.stats.mu.Unlock()

	// 连接目标
	targetAddr := fmt.Sprintf("%s:%d", f.config.DstHost, f.config.DstPort)
	targetConn, err := net.Dial(f.config.Protocol, targetAddr)
	if err != nil {
		logger.Error("连接目标失败: %v", err)
		return
	}
	defer targetConn.Close()

	// 创建同步组
	var wg sync.WaitGroup
	wg.Add(2)

	// 客户端 -> 目标
	go func() {
		defer wg.Done()
		n, err := f.copyData(targetConn, clientConn)
		if err != nil && err != io.EOF {
			logger.Error("转发数据失败 (客户端 -> 目标): %v", err)
		}

		// 更新统计信息
		f.stats.mu.Lock()
		f.stats.BytesSent += uint64(n)
		f.stats.LastActiveTime = time.Now()
		f.stats.mu.Unlock()
	}()

	// 目标 -> 客户端
	go func() {
		defer wg.Done()
		n, err := f.copyData(clientConn, targetConn)
		if err != nil && err != io.EOF {
			logger.Error("转发数据失败 (目标 -> 客户端): %v", err)
		}

		// 更新统计信息
		f.stats.mu.Lock()
		f.stats.BytesReceived += uint64(n)
		f.stats.LastActiveTime = time.Now()
		f.stats.mu.Unlock()
	}()

	// 等待两个方向的数据传输完成
	wg.Wait()

	// 更新连接时间
	f.stats.mu.Lock()
	f.stats.ConnectionTime += uint64(time.Since(f.stats.LastActiveTime).Seconds())
	f.stats.mu.Unlock()
}

// copyData 复制数据
func (f *Forwarder) copyData(dst io.Writer, src io.Reader) (int64, error) {
	buffer := make([]byte, f.bufferSize)
	var total int64

	for {
		select {
		case <-f.stopCh:
			return total, nil
		default:
			// 读取数据
			n, err := src.Read(buffer)
			if err != nil {
				return total, err
			}

			// 写入数据
			_, err = dst.Write(buffer[:n])
			if err != nil {
				return total, err
			}

			total += int64(n)
		}
	}
}

// ForwarderManager 转发器管理器
type ForwarderManager struct {
	forwarders map[string]*Forwarder
	mu         sync.Mutex
}

// NewForwarderManager 创建转发器管理器
func NewForwarderManager() *ForwarderManager {
	return &ForwarderManager{
		forwarders: make(map[string]*Forwarder),
	}
}

// AddForwarder 添加转发器
func (m *ForwarderManager) AddForwarder(cfg *config.AppConfig, bufferSize int) (*Forwarder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在
	if _, exists := m.forwarders[cfg.Name]; exists {
		return nil, fmt.Errorf("转发器已存在: %s", cfg.Name)
	}

	// 创建转发器
	forwarder := NewForwarder(cfg, bufferSize)
	m.forwarders[cfg.Name] = forwarder

	// 如果配置为自动启动，则启动转发器
	if cfg.AutoStart {
		if err := forwarder.Start(); err != nil {
			delete(m.forwarders, cfg.Name)
			return nil, fmt.Errorf("启动转发器失败: %w", err)
		}
	}

	return forwarder, nil
}

// GetForwarder 获取转发器
func (m *ForwarderManager) GetForwarder(name string) (*Forwarder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	forwarder, exists := m.forwarders[name]
	if !exists {
		return nil, fmt.Errorf("转发器不存在: %s", name)
	}

	return forwarder, nil
}

// RemoveForwarder 移除转发器
func (m *ForwarderManager) RemoveForwarder(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	forwarder, exists := m.forwarders[name]
	if !exists {
		return fmt.Errorf("转发器不存在: %s", name)
	}

	// 停止转发器
	if err := forwarder.Stop(); err != nil {
		return fmt.Errorf("停止转发器失败: %w", err)
	}

	// 移除转发器
	delete(m.forwarders, name)
	return nil
}

// GetAllForwarders 获取所有转发器
func (m *ForwarderManager) GetAllForwarders() map[string]*Forwarder {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 创建副本
	result := make(map[string]*Forwarder, len(m.forwarders))
	for name, forwarder := range m.forwarders {
		result[name] = forwarder
	}

	return result
}

// StartAll 启动所有转发器
func (m *ForwarderManager) StartAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, forwarder := range m.forwarders {
		if !forwarder.IsRunning() {
			if err := forwarder.Start(); err != nil {
				return fmt.Errorf("启动转发器 %s 失败: %w", name, err)
			}
		}
	}

	return nil
}

// StopAll 停止所有转发器
func (m *ForwarderManager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, forwarder := range m.forwarders {
		if forwarder.IsRunning() {
			if err := forwarder.Stop(); err != nil {
				return fmt.Errorf("停止转发器 %s 失败: %w", name, err)
			}
		}
	}

	return nil
}
