package relay

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	// TURN 消息类型
	turnBindingRequest       = 0x0001
	turnAllocateRequest      = 0x0003
	turnRefreshRequest       = 0x0004
	turnSendIndication       = 0x0006
	turnDataIndication       = 0x0007
	turnCreatePermission     = 0x0008
	turnChannelBind          = 0x0009
	turnBindingResponse      = 0x0101
	turnAllocateResponse     = 0x0103
	turnRefreshResponse      = 0x0104
	turnCreatePermissionResp = 0x0108
	turnChannelBindResponse  = 0x0109
)

// TURNServer TURN 服务器
type TURNServer struct {
	addr        string
	realm       string
	authSecret  string
	allocations map[string]*Allocation
}

// Allocation 分配
type Allocation struct {
	fiveTuple    string
	relayAddr    *net.UDPAddr
	permissions  map[string]time.Time
	channelBinds map[uint16]string
	lifetime     time.Duration
	createdAt    time.Time
}

// NewTURNServer 创建 TURN 服务器
func NewTURNServer(addr, realm, authSecret string) *TURNServer {
	return &TURNServer{
		addr:        addr,
		realm:       realm,
		authSecret:  authSecret,
		allocations: make(map[string]*Allocation),
	}
}

// Start 启动 TURN 服务器
func (s *TURNServer) Start() error {
	// 解析地址
	udpAddr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return fmt.Errorf("解析地址失败: %w", err)
	}

	// 创建 UDP 连接
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("监听 UDP 失败: %w", err)
	}
	defer conn.Close()

	fmt.Printf("TURN 服务器已启动，监听地址: %s\n", s.addr)

	// 处理请求
	buffer := make([]byte, 1500)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("读取 UDP 失败: %v\n", err)
			continue
		}

		// 处理 TURN 消息
		go s.handleTURNMessage(conn, addr, buffer[:n])
	}
}

// handleTURNMessage 处理 TURN 消息
func (s *TURNServer) handleTURNMessage(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	// 解析消息头
	if len(data) < 20 {
		fmt.Println("消息太短")
		return
	}

	// 提取消息类型
	messageType := binary.BigEndian.Uint16(data[0:2])

	// 根据消息类型处理
	switch messageType {
	case turnBindingRequest:
		s.handleBindingRequest(conn, addr, data)
	case turnAllocateRequest:
		s.handleAllocateRequest(conn, addr, data)
	case turnRefreshRequest:
		s.handleRefreshRequest(conn, addr, data)
	case turnCreatePermission:
		s.handleCreatePermission(conn, addr, data)
	case turnChannelBind:
		s.handleChannelBind(conn, addr, data)
	case turnSendIndication:
		s.handleSendIndication(conn, addr, data)
	default:
		fmt.Printf("未知消息类型: %04x\n", messageType)
	}
}

// handleBindingRequest 处理 Binding 请求
func (s *TURNServer) handleBindingRequest(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	// 提取事务 ID
	transactionID := data[8:20]

	// 创建响应
	response := new(bytes.Buffer)
	
	// 写入消息类型
	binary.Write(response, binary.BigEndian, uint16(turnBindingResponse))
	
	// 写入消息长度（暂时为0）
	binary.Write(response, binary.BigEndian, uint16(8))
	
	// 写入魔术字
	binary.Write(response, binary.BigEndian, uint32(0x2112A442))
	
	// 写入事务 ID
	response.Write(transactionID)
	
	// 写入 XOR-MAPPED-ADDRESS 属性
	binary.Write(response, binary.BigEndian, uint16(0x0020)) // 属性类型
	binary.Write(response, binary.BigEndian, uint16(8))      // 属性长度
	response.WriteByte(0)                                    // 保留
	response.WriteByte(0x01)                                 // IPv4
	
	// 异或端口
	port := uint16(addr.Port)
	port ^= 0x2112 // 魔术字的前 16 位
	binary.Write(response, binary.BigEndian, port)
	
	// 异或 IP
	ip := addr.IP.To4()
	xorIP := make([]byte, 4)
	for i := 0; i < 4; i++ {
		xorIP[i] = ip[i] ^ 0x21
	}
	response.Write(xorIP)
	
	// 发送响应
	conn.WriteToUDP(response.Bytes(), addr)
}

// handleAllocateRequest 处理 Allocate 请求
func (s *TURNServer) handleAllocateRequest(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	// 这里简化实现，实际应该解析请求属性并验证认证
	
	// 创建分配
	fiveTuple := addr.String()
	
	// 分配中继地址
	relayAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		fmt.Printf("解析中继地址失败: %v\n", err)
		return
	}
	
	// 创建中继套接字
	relayConn, err := net.ListenUDP("udp", relayAddr)
	if err != nil {
		fmt.Printf("创建中继套接字失败: %v\n", err)
		return
	}
	
	// 获取实际分配的地址
	relayAddr = relayConn.LocalAddr().(*net.UDPAddr)
	
	// 创建分配
	allocation := &Allocation{
		fiveTuple:    fiveTuple,
		relayAddr:    relayAddr,
		permissions:  make(map[string]time.Time),
		channelBinds: make(map[uint16]string),
		lifetime:     10 * time.Minute,
		createdAt:    time.Now(),
	}
	
	// 保存分配
	s.allocations[fiveTuple] = allocation
	
	// 提取事务 ID
	transactionID := data[8:20]
	
	// 创建响应
	response := new(bytes.Buffer)
	
	// 写入消息类型
	binary.Write(response, binary.BigEndian, uint16(turnAllocateResponse))
	
	// 写入消息长度（暂时为0）
	binary.Write(response, binary.BigEndian, uint16(16))
	
	// 写入魔术字
	binary.Write(response, binary.BigEndian, uint32(0x2112A442))
	
	// 写入事务 ID
	response.Write(transactionID)
	
	// 写入 XOR-RELAYED-ADDRESS 属性
	binary.Write(response, binary.BigEndian, uint16(0x0016)) // 属性类型
	binary.Write(response, binary.BigEndian, uint16(8))      // 属性长度
	response.WriteByte(0)                                    // 保留
	response.WriteByte(0x01)                                 // IPv4
	
	// 异或端口
	port := uint16(relayAddr.Port)
	port ^= 0x2112 // 魔术字的前 16 位
	binary.Write(response, binary.BigEndian, port)
	
	// 异或 IP
	ip := relayAddr.IP.To4()
	xorIP := make([]byte, 4)
	for i := 0; i < 4; i++ {
		xorIP[i] = ip[i] ^ 0x21
	}
	response.Write(xorIP)
	
	// 写入 LIFETIME 属性
	binary.Write(response, binary.BigEndian, uint16(0x000D)) // 属性类型
	binary.Write(response, binary.BigEndian, uint16(4))      // 属性长度
	binary.Write(response, binary.BigEndian, uint32(600))    // 10分钟
	
	// 发送响应
	conn.WriteToUDP(response.Bytes(), addr)
	
	// 启动中继
	go s.relay(conn, relayConn, allocation)
}

// handleRefreshRequest 处理 Refresh 请求
func (s *TURNServer) handleRefreshRequest(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	// 实现 Refresh 请求处理
}

// handleCreatePermission 处理 CreatePermission 请求
func (s *TURNServer) handleCreatePermission(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	// 实现 CreatePermission 请求处理
}

// handleChannelBind 处理 ChannelBind 请求
func (s *TURNServer) handleChannelBind(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	// 实现 ChannelBind 请求处理
}

// handleSendIndication 处理 SendIndication 请求
func (s *TURNServer) handleSendIndication(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	// 实现 SendIndication 请求处理
}

// relay 中继数据
func (s *TURNServer) relay(clientConn *net.UDPConn, relayConn *net.UDPConn, allocation *Allocation) {
	defer relayConn.Close()
	
	// 从客户端到对等方
	go func() {
		buffer := make([]byte, 1500)
		for {
			n, _, err := relayConn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Printf("从中继读取失败: %v\n", err)
				return
			}
			
			// 创建 Data 指示
			indication := new(bytes.Buffer)
			
			// 写入消息类型
			binary.Write(indication, binary.BigEndian, uint16(turnDataIndication))
			
			// 写入消息长度（暂时为0）
			binary.Write(indication, binary.BigEndian, uint16(n + 16))
			
			// 写入魔术字
			binary.Write(indication, binary.BigEndian, uint32(0x2112A442))
			
			// 写入事务 ID（随机生成）
			transactionID := make([]byte, 12)
			rand.Read(transactionID)
			indication.Write(transactionID)
			
			// 写入 DATA 属性
			binary.Write(indication, binary.BigEndian, uint16(0x0013)) // 属性类型
			binary.Write(indication, binary.BigEndian, uint16(n))      // 属性长度
			indication.Write(buffer[:n])
			
			// 发送指示
			clientAddr, _ := net.ResolveUDPAddr("udp", allocation.fiveTuple)
			clientConn.WriteToUDP(indication.Bytes(), clientAddr)
		}
	}()
	
	// 从对等方到客户端
	buffer := make([]byte, 1500)
	for {
		n, addr, err := clientConn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("从客户端读取失败: %v\n", err)
			return
		}
		
		// 检查是否是 SendIndication
		if n >= 20 && binary.BigEndian.Uint16(buffer[:2]) == turnSendIndication {
			// 解析 SendIndication
			// 提取 XOR-PEER-ADDRESS 和 DATA 属性
			// 发送数据到对等方
			// 这里简化实现
			relayConn.WriteToUDP(buffer[20:n], addr)
		}
	}
}
