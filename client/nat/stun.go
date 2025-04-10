package nat

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	stunMagicCookie = 0x2112A442
)

// STUN 消息类型
const (
	stunBindingRequest       = 0x0001
	stunBindingResponse      = 0x0101
	stunBindingErrorResponse = 0x0111
)

// STUN 属性类型
const (
	stunAttrMappedAddress    = 0x0001
	stunAttrXorMappedAddress = 0x0020
	stunAttrSoftware         = 0x8022
	stunAttrFingerprint      = 0x8028
)

// STUNMessage STUN 消息结构
type STUNMessage struct {
	Type        uint16
	Length      uint16
	MagicCookie uint32
	TransID     [12]byte
	Attributes  []STUNAttribute
}

// STUNAttribute STUN 属性结构
type STUNAttribute struct {
	Type   uint16
	Length uint16
	Value  []byte
}

// NewSTUNRequest 创建 STUN 绑定请求
func NewSTUNRequest() (*STUNMessage, error) {
	msg := &STUNMessage{
		Type:        stunBindingRequest,
		Length:      0,
		MagicCookie: stunMagicCookie,
	}

	// 生成随机事务 ID
	if _, err := rand.Read(msg.TransID[:]); err != nil {
		return nil, err
	}

	return msg, nil
}

// Marshal 将 STUN 消息序列化为字节数组
func (m *STUNMessage) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// 写入消息头
	if err := binary.Write(buf, binary.BigEndian, m.Type); err != nil {
		return nil, err
	}

	// 计算属性长度
	attrLen := 0
	for _, attr := range m.Attributes {
		attrLen += 4 + len(attr.Value) // 4 字节头 + 值长度
		// 填充到 4 字节边界
		padding := (4 - (len(attr.Value) % 4)) % 4
		attrLen += padding
	}

	// 写入消息长度
	if err := binary.Write(buf, binary.BigEndian, uint16(attrLen)); err != nil {
		return nil, err
	}

	// 写入魔术字
	if err := binary.Write(buf, binary.BigEndian, m.MagicCookie); err != nil {
		return nil, err
	}

	// 写入事务 ID
	if _, err := buf.Write(m.TransID[:]); err != nil {
		return nil, err
	}

	// 写入属性
	for _, attr := range m.Attributes {
		if err := binary.Write(buf, binary.BigEndian, attr.Type); err != nil {
			return nil, err
		}
		if err := binary.Write(buf, binary.BigEndian, uint16(len(attr.Value))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(attr.Value); err != nil {
			return nil, err
		}

		// 填充到 4 字节边界
		padding := (4 - (len(attr.Value) % 4)) % 4
		if padding > 0 {
			if _, err := buf.Write(make([]byte, padding)); err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

// Unmarshal 从字节数组解析 STUN 消息
func (m *STUNMessage) Unmarshal(data []byte) error {
	if len(data) < 20 {
		return errors.New("STUN 消息太短")
	}

	// 解析消息头
	m.Type = binary.BigEndian.Uint16(data[0:2])
	m.Length = binary.BigEndian.Uint16(data[2:4])
	m.MagicCookie = binary.BigEndian.Uint32(data[4:8])
	copy(m.TransID[:], data[8:20])

	// 检查魔术字
	if m.MagicCookie != stunMagicCookie {
		return errors.New("无效的 STUN 魔术字")
	}

	// 解析属性
	m.Attributes = nil
	offset := 20

	for offset < len(data) {
		if offset+4 > len(data) {
			return errors.New("无效的 STUN 属性头")
		}

		attrType := binary.BigEndian.Uint16(data[offset : offset+2])
		attrLen := binary.BigEndian.Uint16(data[offset+2 : offset+4])
		offset += 4

		if offset+int(attrLen) > len(data) {
			return errors.New("无效的 STUN 属性长度")
		}

		attrValue := make([]byte, attrLen)
		copy(attrValue, data[offset:offset+int(attrLen)])
		offset += int(attrLen)

		// 跳过填充字节
		padding := (4 - (int(attrLen) % 4)) % 4
		offset += padding

		m.Attributes = append(m.Attributes, STUNAttribute{
			Type:   attrType,
			Length: attrLen,
			Value:  attrValue,
		})
	}

	return nil
}

// GetXorMappedAddress 获取 XOR-MAPPED-ADDRESS 属性
func (m *STUNMessage) GetXorMappedAddress() (net.IP, int, error) {
	for _, attr := range m.Attributes {
		if attr.Type == stunAttrXorMappedAddress {
			if len(attr.Value) < 8 {
				return nil, 0, errors.New("无效的 XOR-MAPPED-ADDRESS 属性")
			}

			// 忽略第一个字节（保留）和第二个字节（地址族）
			family := attr.Value[1]
			port := binary.BigEndian.Uint16(attr.Value[2:4])
			// 异或端口与魔术字的前 16 位
			port ^= uint16(stunMagicCookie >> 16)

			var ip net.IP
			if family == 0x01 { // IPv4
				if len(attr.Value) < 8 {
					return nil, 0, errors.New("无效的 IPv4 地址")
				}
				ip = make(net.IP, 4)
				copy(ip, attr.Value[4:8])
				// 异或 IP 与魔术字
				binary.BigEndian.PutUint32(ip, binary.BigEndian.Uint32(ip)^stunMagicCookie)
			} else if family == 0x02 { // IPv6
				if len(attr.Value) < 20 {
					return nil, 0, errors.New("无效的 IPv6 地址")
				}
				ip = make(net.IP, 16)
				copy(ip, attr.Value[4:20])
				// 异或 IP 与魔术字和事务 ID
				for i := 0; i < 4; i++ {
					binary.BigEndian.PutUint32(ip[i*4:], binary.BigEndian.Uint32(ip[i*4:])^stunMagicCookie)
				}
				for i := 0; i < 12; i++ {
					ip[i+4] ^= m.TransID[i]
				}
			} else {
				return nil, 0, fmt.Errorf("不支持的地址族: %d", family)
			}

			return ip, int(port), nil
		}
	}

	// 尝试获取 MAPPED-ADDRESS 属性
	for _, attr := range m.Attributes {
		if attr.Type == stunAttrMappedAddress {
			if len(attr.Value) < 8 {
				return nil, 0, errors.New("无效的 MAPPED-ADDRESS 属性")
			}

			// 忽略第一个字节（保留）和第二个字节（地址族）
			family := attr.Value[1]
			port := binary.BigEndian.Uint16(attr.Value[2:4])

			var ip net.IP
			if family == 0x01 { // IPv4
				if len(attr.Value) < 8 {
					return nil, 0, errors.New("无效的 IPv4 地址")
				}
				ip = net.IPv4(attr.Value[4], attr.Value[5], attr.Value[6], attr.Value[7])
			} else if family == 0x02 { // IPv6
				if len(attr.Value) < 20 {
					return nil, 0, errors.New("无效的 IPv6 地址")
				}
				ip = make(net.IP, 16)
				copy(ip, attr.Value[4:20])
			} else {
				return nil, 0, fmt.Errorf("不支持的地址族: %d", family)
			}

			return ip, int(port), nil
		}
	}

	return nil, 0, errors.New("未找到地址属性")
}

// STUNClient STUN 客户端
type STUNClient struct {
	Servers []string
	Timeout time.Duration
}

// NewSTUNClient 创建 STUN 客户端
func NewSTUNClient(servers []string, timeout time.Duration) *STUNClient {
	if len(servers) == 0 {
		// 默认 STUN 服务器
		servers = []string{
			"stun.l.google.com:19302",
			"stun1.l.google.com:19302",
			"stun2.l.google.com:19302",
			"stun3.l.google.com:19302",
			"stun4.l.google.com:19302",
		}
	}

	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &STUNClient{
		Servers: servers,
		Timeout: timeout,
	}
}

// Discover 发现外部 IP 和端口
func (c *STUNClient) Discover() (net.IP, int, error) {
	// 尝试所有 STUN 服务器
	var lastErr error
	for _, server := range c.Servers {
		ip, port, err := c.discoverWithServer(server)
		if err == nil {
			return ip, port, nil
		}
		lastErr = err
	}

	return nil, 0, fmt.Errorf("所有 STUN 服务器都失败: %v", lastErr)
}

// discoverWithServer 使用指定的 STUN 服务器发现外部 IP 和端口
func (c *STUNClient) discoverWithServer(server string) (net.IP, int, error) {
	// 解析服务器地址
	serverAddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		return nil, 0, fmt.Errorf("解析 STUN 服务器地址失败: %w", err)
	}

	// 创建 UDP 连接
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return nil, 0, fmt.Errorf("连接 STUN 服务器失败: %w", err)
	}
	defer conn.Close()

	// 设置超时
	if err := conn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, 0, fmt.Errorf("设置超时失败: %w", err)
	}

	// 创建 STUN 请求
	req, err := NewSTUNRequest()
	if err != nil {
		return nil, 0, fmt.Errorf("创建 STUN 请求失败: %w", err)
	}

	// 序列化请求
	reqData, err := req.Marshal()
	if err != nil {
		return nil, 0, fmt.Errorf("序列化 STUN 请求失败: %w", err)
	}

	// 发送请求
	if _, err := conn.Write(reqData); err != nil {
		return nil, 0, fmt.Errorf("发送 STUN 请求失败: %w", err)
	}

	// 接收响应
	respData := make([]byte, 1024)
	n, err := conn.Read(respData)
	if err != nil {
		return nil, 0, fmt.Errorf("接收 STUN 响应失败: %w", err)
	}
	respData = respData[:n]

	// 解析响应
	resp := &STUNMessage{}
	if err := resp.Unmarshal(respData); err != nil {
		return nil, 0, fmt.Errorf("解析 STUN 响应失败: %w", err)
	}

	// 检查响应类型
	if resp.Type != stunBindingResponse {
		return nil, 0, fmt.Errorf("无效的 STUN 响应类型: %d", resp.Type)
	}

	// 检查事务 ID
	if !bytes.Equal(resp.TransID[:], req.TransID[:]) {
		return nil, 0, errors.New("STUN 事务 ID 不匹配")
	}

	// 获取外部 IP 和端口
	ip, port, err := resp.GetXorMappedAddress()
	if err != nil {
		return nil, 0, fmt.Errorf("获取外部地址失败: %w", err)
	}

	return ip, port, nil
}

// DetectNATType 检测 NAT 类型
func (c *STUNClient) DetectNATType() (NATType, error) {
	// 实现 NAT 类型检测算法
	// 这里使用简化版的算法，完整算法参考 RFC 5780

	// 第一次测试：检查是否有公网 IP
	ip, _, err := c.Discover()
	if err != nil {
		return NATUnknown, fmt.Errorf("第一次 STUN 测试失败: %w", err)
	}

	// 获取本地 IP
	localIP, err := getLocalIP()
	if err != nil {
		return NATUnknown, fmt.Errorf("获取本地 IP 失败: %w", err)
	}

	// 如果外部 IP 与本地 IP 相同，则没有 NAT
	if ip.Equal(localIP) {
		return NATNone, nil
	}

	// TODO: 实现完整的 NAT 类型检测算法
	// 这需要多次 STUN 测试，包括：
	// 1. 使用不同的 STUN 服务器
	// 2. 使用相同的 STUN 服务器但不同的端口
	// 3. 检查端口映射行为

	// 默认返回端口受限锥形 NAT
	return NATPortRestricted, nil
}

// getLocalIP 获取本地 IP
func getLocalIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
