package p2p

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/senma231/p3/client/nat"
)

// PunchResult 打洞结果
type PunchResult struct {
	Success      bool
	Conn         net.Conn
	ConnectionType ConnectionType
	Error        error
}

// ConnectionType 连接类型
type ConnectionType int

const (
	ConnectionTypeUnknown ConnectionType = iota
	ConnectionTypeDirect               // 直接连接
	ConnectionTypeHolePunch            // 打洞连接
	ConnectionTypeRelay                // 中继连接
)

// String 返回连接类型的字符串表示
func (t ConnectionType) String() string {
	switch t {
	case ConnectionTypeDirect:
		return "Direct"
	case ConnectionTypeHolePunch:
		return "Hole Punch"
	case ConnectionTypeRelay:
		return "Relay"
	default:
		return "Unknown"
	}
}

// Puncher 打洞器
type Puncher struct {
	localPort  int
	natInfo    *nat.NATInfo
	timeout    time.Duration
	maxRetries int
}

// NewPuncher 创建打洞器
func NewPuncher(localPort int, natInfo *nat.NATInfo, timeout time.Duration, maxRetries int) *Puncher {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	if maxRetries == 0 {
		maxRetries = 5
	}
	return &Puncher{
		localPort:  localPort,
		natInfo:    natInfo,
		timeout:    timeout,
		maxRetries: maxRetries,
	}
}

// Punch 尝试打洞连接
func (p *Puncher) Punch(peerIP string, peerPort int, peerNATType nat.NATType) *PunchResult {
	// 检查是否可以直接连接
	if p.canDirectConnect(peerNATType) {
		conn, err := p.directConnect(peerIP, peerPort)
		if err == nil {
			return &PunchResult{
				Success:      true,
				Conn:         conn,
				ConnectionType: ConnectionTypeDirect,
			}
		}
	}

	// 尝试打洞
	conn, err := p.holePunch(peerIP, peerPort, peerNATType)
	if err == nil {
		return &PunchResult{
			Success:      true,
			Conn:         conn,
			ConnectionType: ConnectionTypeHolePunch,
		}
	}

	// 打洞失败
	return &PunchResult{
		Success:      false,
		ConnectionType: ConnectionTypeUnknown,
		Error:        err,
	}
}

// canDirectConnect 检查是否可以直接连接
func (p *Puncher) canDirectConnect(peerNATType nat.NATType) bool {
	// 如果对方没有 NAT，可以直接连接
	if peerNATType == nat.NATNone {
		return true
	}

	// 如果对方是完全锥形 NAT，可以直接连接
	if peerNATType == nat.NATFull {
		return true
	}

	// 如果本地没有 NAT，可以直接连接
	if p.natInfo.Type == nat.NATNone {
		return true
	}

	return false
}

// directConnect 直接连接
func (p *Puncher) directConnect(peerIP string, peerPort int) (net.Conn, error) {
	// 创建 TCP 连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", peerIP, peerPort), p.timeout)
	if err != nil {
		return nil, fmt.Errorf("直接连接失败: %w", err)
	}
	return conn, nil
}

// holePunch 打洞连接
func (p *Puncher) holePunch(peerIP string, peerPort int, peerNATType nat.NATType) (net.Conn, error) {
	// 创建 UDP 监听器
	localAddr := &net.UDPAddr{
		IP:   p.natInfo.LocalIP,
		Port: p.localPort,
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("创建 UDP 监听器失败: %w", err)
	}
	defer conn.Close()

	// 设置超时
	conn.SetDeadline(time.Now().Add(p.timeout))

	// 创建对等端地址
	peerAddr := &net.UDPAddr{
		IP:   net.ParseIP(peerIP),
		Port: peerPort,
	}

	// 创建打洞消息
	punchMsg := []byte("PUNCH")

	// 创建接收通道
	receiveCh := make(chan *net.UDPAddr, 1)
	errorCh := make(chan error, 1)
	stopCh := make(chan struct{})
	var wg sync.WaitGroup

	// 启动接收协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer := make([]byte, 1024)
		for {
			select {
			case <-stopCh:
				return
			default:
				// 接收数据
				conn.SetReadDeadline(time.Now().Add(time.Second))
				n, addr, err := conn.ReadFromUDP(buffer)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
					errorCh <- fmt.Errorf("接收数据失败: %w", err)
					return
				}

				// 检查是否是打洞消息
				if n == len(punchMsg) && string(buffer[:n]) == string(punchMsg) {
					receiveCh <- addr
					return
				}
			}
		}
	}()

	// 发送打洞消息
	for i := 0; i < p.maxRetries; i++ {
		_, err := conn.WriteToUDP(punchMsg, peerAddr)
		if err != nil {
			close(stopCh)
			wg.Wait()
			return nil, fmt.Errorf("发送打洞消息失败: %w", err)
		}

		// 等待一段时间
		time.Sleep(time.Second)
	}

	// 等待接收结果
	select {
	case addr := <-receiveCh:
		// 创建新的 UDP 连接
		close(stopCh)
		wg.Wait()
		newConn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			return nil, fmt.Errorf("创建 UDP 连接失败: %w", err)
		}
		return newConn, nil
	case err := <-errorCh:
		close(stopCh)
		wg.Wait()
		return nil, err
	case <-time.After(p.timeout):
		close(stopCh)
		wg.Wait()
		return nil, fmt.Errorf("打洞超时")
	}
}

// PunchWithRelay 使用中继服务器打洞
func (p *Puncher) PunchWithRelay(relayServer string, peerID string) *PunchResult {
	// 连接中继服务器
	conn, err := net.DialTimeout("tcp", relayServer, p.timeout)
	if err != nil {
		return &PunchResult{
			Success:      false,
			ConnectionType: ConnectionTypeUnknown,
			Error:        fmt.Errorf("连接中继服务器失败: %w", err),
		}
	}

	// 发送中继请求
	relayRequest := fmt.Sprintf("RELAY %s", peerID)
	_, err = conn.Write([]byte(relayRequest))
	if err != nil {
		conn.Close()
		return &PunchResult{
			Success:      false,
			ConnectionType: ConnectionTypeUnknown,
			Error:        fmt.Errorf("发送中继请求失败: %w", err),
		}
	}

	// 设置超时
	conn.SetDeadline(time.Now().Add(p.timeout))

	// 接收中继响应
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		conn.Close()
		return &PunchResult{
			Success:      false,
			ConnectionType: ConnectionTypeUnknown,
			Error:        fmt.Errorf("接收中继响应失败: %w", err),
		}
	}

	// 检查响应
	response := string(buffer[:n])
	if response != "OK" {
		conn.Close()
		return &PunchResult{
			Success:      false,
			ConnectionType: ConnectionTypeUnknown,
			Error:        fmt.Errorf("中继服务器拒绝请求: %s", response),
		}
	}

	// 中继连接成功
	return &PunchResult{
		Success:      true,
		Conn:         conn,
		ConnectionType: ConnectionTypeRelay,
	}
}
