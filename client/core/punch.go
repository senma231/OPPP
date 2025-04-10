package core

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/senma231/p3/client/nat"
)

// PunchType 打洞类型
type PunchType int

const (
	PunchUnknown PunchType = iota
	PunchUDP               // UDP 打洞
	PunchTCP               // TCP 打洞
)

// String 返回打洞类型的字符串表示
func (t PunchType) String() string {
	switch t {
	case PunchUDP:
		return "UDP"
	case PunchTCP:
		return "TCP"
	default:
		return "Unknown"
	}
}

// PunchResult 打洞结果
type PunchResult struct {
	Success bool
	Type    PunchType
	Conn    net.Conn
	Error   error
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
		timeout = 5 * time.Second
	}
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &Puncher{
		localPort:  localPort,
		natInfo:    natInfo,
		timeout:    timeout,
		maxRetries: maxRetries,
	}
}

// Punch 尝试打洞连接
func (p *Puncher) Punch(peerIP net.IP, peerPort int, peerNATType nat.NATType) *PunchResult {
	// 根据 NAT 类型选择打洞策略
	canUDP := p.canUDPPunch(p.natInfo.Type, peerNATType)
	canTCP := p.canTCPPunch(p.natInfo.Type, peerNATType)

	if !canUDP && !canTCP {
		return &PunchResult{
			Success: false,
			Error:   fmt.Errorf("无法打洞：NAT 类型不兼容"),
		}
	}

	// 并行尝试 UDP 和 TCP 打洞
	var wg sync.WaitGroup
	resultCh := make(chan *PunchResult, 2)

	if canUDP {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := p.punchUDP(peerIP, peerPort)
			resultCh <- result
		}()
	}

	if canTCP {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := p.punchTCP(peerIP, peerPort)
			resultCh <- result
		}()
	}

	// 等待所有打洞尝试完成
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 选择第一个成功的结果
	for result := range resultCh {
		if result.Success {
			return result
		}
	}

	// 所有尝试都失败
	return &PunchResult{
		Success: false,
		Error:   fmt.Errorf("所有打洞尝试都失败"),
	}
}

// canUDPPunch 判断是否可以进行 UDP 打洞
func (p *Puncher) canUDPPunch(localNATType, peerNATType nat.NATType) bool {
	// 如果任一方是公网 IP，可以直接连接，不需要打洞
	if localNATType == nat.NATNone || peerNATType == nat.NATNone {
		return true
	}

	// 如果双方都是对称型 NAT，无法进行 UDP 打洞
	if localNATType == nat.NATSymmetric && peerNATType == nat.NATSymmetric {
		return false
	}

	// 其他情况可以尝试 UDP 打洞
	return true
}

// canTCPPunch 判断是否可以进行 TCP 打洞
func (p *Puncher) canTCPPunch(localNATType, peerNATType nat.NATType) bool {
	// TCP 打洞成功率较低，只在特定情况下尝试
	// 如果任一方是公网 IP，可以直接连接，不需要打洞
	if localNATType == nat.NATNone || peerNATType == nat.NATNone {
		return true
	}

	// 如果双方都是完全锥形或受限锥形 NAT，可以尝试 TCP 打洞
	if (localNATType == nat.NATFull || localNATType == nat.NATRestricted) &&
		(peerNATType == nat.NATFull || peerNATType == nat.NATRestricted) {
		return true
	}

	// 其他情况不尝试 TCP 打洞
	return false
}

// punchUDP 尝试 UDP 打洞
func (p *Puncher) punchUDP(peerIP net.IP, peerPort int) *PunchResult {
	// 创建 UDP 连接
	localAddr := &net.UDPAddr{Port: p.localPort}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return &PunchResult{
			Success: false,
			Type:    PunchUDP,
			Error:   fmt.Errorf("创建 UDP 连接失败: %w", err),
		}
	}
	defer conn.Close()

	// 设置超时
	conn.SetDeadline(time.Now().Add(p.timeout))

	// 创建目标地址
	peerAddr := &net.UDPAddr{
		IP:   peerIP,
		Port: peerPort,
	}

	// 发送打洞包
	punchData := []byte("P3_UDP_PUNCH")
	for i := 0; i < p.maxRetries; i++ {
		_, err = conn.WriteToUDP(punchData, peerAddr)
		if err != nil {
			continue
		}

		// 等待响应
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		// 检查是否是目标地址
		if addr.IP.Equal(peerIP) && addr.Port == peerPort {
			// 检查响应数据
			if n >= len(punchData) && string(buf[:len(punchData)]) == "P3_UDP_PUNCH_ACK" {
				// 创建新连接
				newConn, err := net.DialUDP("udp", nil, peerAddr)
				if err != nil {
					return &PunchResult{
						Success: false,
						Type:    PunchUDP,
						Error:   fmt.Errorf("创建 UDP 连接失败: %w", err),
					}
				}

				return &PunchResult{
					Success: true,
					Type:    PunchUDP,
					Conn:    newConn,
				}
			}
		}

		// 等待一段时间再重试
		time.Sleep(500 * time.Millisecond)
	}

	return &PunchResult{
		Success: false,
		Type:    PunchUDP,
		Error:   fmt.Errorf("UDP 打洞失败: 超时"),
	}
}

// punchTCP 尝试 TCP 打洞
func (p *Puncher) punchTCP(peerIP net.IP, peerPort int) *PunchResult {
	// TCP 打洞需要同时监听和连接
	// 创建监听器
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", p.localPort))
	if err != nil {
		return &PunchResult{
			Success: false,
			Type:    PunchTCP,
			Error:   fmt.Errorf("创建 TCP 监听器失败: %w", err),
		}
	}
	defer listener.Close()

	// 创建连接通道
	connCh := make(chan net.Conn, 1)
	errCh := make(chan error, 1)

	// 启动监听协程
	go func() {
		// 设置监听超时
		listener.(*net.TCPListener).SetDeadline(time.Now().Add(p.timeout))

		// 等待连接
		conn, err := listener.Accept()
		if err != nil {
			errCh <- fmt.Errorf("接受 TCP 连接失败: %w", err)
			return
		}

		// 检查是否是目标地址
		remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
		if !remoteAddr.IP.Equal(peerIP) {
			conn.Close()
			errCh <- fmt.Errorf("收到非目标地址的连接: %s", remoteAddr.String())
			return
		}

		connCh <- conn
	}()

	// 尝试连接对方
	go func() {
		// 等待一段时间再连接，让对方有时间启动监听
		time.Sleep(100 * time.Millisecond)

		// 创建目标地址
		peerAddr := net.JoinHostPort(peerIP.String(), fmt.Sprintf("%d", peerPort))

		// 尝试连接
		for i := 0; i < p.maxRetries; i++ {
			conn, err := net.DialTimeout("tcp", peerAddr, p.timeout/2)
			if err == nil {
				connCh <- conn
				return
			}

			// 等待一段时间再重试
			time.Sleep(500 * time.Millisecond)
		}

		errCh <- fmt.Errorf("TCP 连接失败: 超时")
	}()

	// 等待连接或错误
	select {
	case conn := <-connCh:
		return &PunchResult{
			Success: true,
			Type:    PunchTCP,
			Conn:    conn,
		}
	case err := <-errCh:
		return &PunchResult{
			Success: false,
			Type:    PunchTCP,
			Error:   err,
		}
	case <-time.After(p.timeout):
		return &PunchResult{
			Success: false,
			Type:    PunchTCP,
			Error:   fmt.Errorf("TCP 打洞超时"),
		}
	}
}

// HandlePunchRequest 处理打洞请求
func (p *Puncher) HandlePunchRequest(conn net.Conn, punchType PunchType) error {
	switch punchType {
	case PunchUDP:
		return p.handleUDPPunchRequest(conn.(*net.UDPConn))
	case PunchTCP:
		return p.handleTCPPunchRequest(conn.(*net.TCPConn))
	default:
		return fmt.Errorf("不支持的打洞类型: %s", punchType)
	}
}

// handleUDPPunchRequest 处理 UDP 打洞请求
func (p *Puncher) handleUDPPunchRequest(conn *net.UDPConn) error {
	// 设置超时
	conn.SetDeadline(time.Now().Add(p.timeout))

	// 读取打洞请求
	buf := make([]byte, 1024)
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return fmt.Errorf("读取 UDP 打洞请求失败: %w", err)
	}

	// 检查请求数据
	if n < len("P3_UDP_PUNCH") || string(buf[:len("P3_UDP_PUNCH")]) != "P3_UDP_PUNCH" {
		return fmt.Errorf("无效的 UDP 打洞请求")
	}

	// 发送响应
	_, err = conn.WriteToUDP([]byte("P3_UDP_PUNCH_ACK"), addr)
	if err != nil {
		return fmt.Errorf("发送 UDP 打洞响应失败: %w", err)
	}

	return nil
}

// handleTCPPunchRequest 处理 TCP 打洞请求
func (p *Puncher) handleTCPPunchRequest(conn *net.TCPConn) error {
	// TCP 打洞不需要特殊处理，连接已经建立
	return nil
}
