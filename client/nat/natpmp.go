package nat

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	natPmpPort       = 5351
	natPmpVersion    = 0
	natPmpExternalIP = 0
	natPmpMapUDP     = 1
	natPmpMapTCP     = 2
)

// NATPMPClient NAT-PMP 客户端
type NATPMPClient struct {
	gateway net.IP
	timeout time.Duration
}

// NewNATPMPClient 创建 NAT-PMP 客户端
func NewNATPMPClient(gateway net.IP, timeout time.Duration) *NATPMPClient {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	
	return &NATPMPClient{
		gateway: gateway,
		timeout: timeout,
	}
}

// GetExternalIP 获取外部 IP
func (c *NATPMPClient) GetExternalIP() (net.IP, error) {
	// 创建请求
	req := []byte{natPmpVersion, natPmpExternalIP}
	
	// 发送请求
	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}
	
	// 解析响应
	if len(resp) < 12 {
		return nil, errors.New("响应太短")
	}
	
	// 检查版本和操作码
	if resp[0] != 0 {
		return nil, fmt.Errorf("不支持的版本: %d", resp[0])
	}
	if resp[1] != 128+natPmpExternalIP {
		return nil, fmt.Errorf("不匹配的操作码: %d", resp[1])
	}
	
	// 检查结果码
	resultCode := binary.BigEndian.Uint16(resp[2:4])
	if resultCode != 0 {
		return nil, fmt.Errorf("请求失败: %d", resultCode)
	}
	
	// 提取 IP
	ip := net.IPv4(resp[8], resp[9], resp[10], resp[11])
	
	return ip, nil
}

// AddPortMapping 添加端口映射
func (c *NATPMPClient) AddPortMapping(protocol string, internalPort, externalPort int, lifetime int) (int, error) {
	// 确定协议
	var opcode byte
	if protocol == "tcp" {
		opcode = natPmpMapTCP
	} else if protocol == "udp" {
		opcode = natPmpMapUDP
	} else {
		return 0, fmt.Errorf("不支持的协议: %s", protocol)
	}
	
	// 创建请求
	req := make([]byte, 12)
	req[0] = natPmpVersion
	req[1] = opcode
	// 2-3 保留为 0
	binary.BigEndian.PutUint16(req[4:6], uint16(internalPort))
	binary.BigEndian.PutUint16(req[6:8], uint16(externalPort))
	binary.BigEndian.PutUint32(req[8:12], uint32(lifetime))
	
	// 发送请求
	resp, err := c.sendRequest(req)
	if err != nil {
		return 0, err
	}
	
	// 解析响应
	if len(resp) < 16 {
		return 0, errors.New("响应太短")
	}
	
	// 检查版本和操作码
	if resp[0] != 0 {
		return 0, fmt.Errorf("不支持的版本: %d", resp[0])
	}
	if resp[1] != 128+opcode {
		return 0, fmt.Errorf("不匹配的操作码: %d", resp[1])
	}
	
	// 检查结果码
	resultCode := binary.BigEndian.Uint16(resp[2:4])
	if resultCode != 0 {
		return 0, fmt.Errorf("请求失败: %d", resultCode)
	}
	
	// 提取分配的外部端口
	assignedPort := int(binary.BigEndian.Uint16(resp[10:12]))
	
	return assignedPort, nil
}

// DeletePortMapping 删除端口映射
func (c *NATPMPClient) DeletePortMapping(protocol string, internalPort int) error {
	// 调用 AddPortMapping 并将生存期设为 0
	_, err := c.AddPortMapping(protocol, internalPort, 0, 0)
	return err
}

// sendRequest 发送请求
func (c *NATPMPClient) sendRequest(req []byte) ([]byte, error) {
	// 创建 UDP 连接
	addr := &net.UDPAddr{
		IP:   c.gateway,
		Port: natPmpPort,
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("连接网关失败: %w", err)
	}
	defer conn.Close()
	
	// 设置超时
	conn.SetDeadline(time.Now().Add(c.timeout))
	
	// 发送请求
	_, err = conn.Write(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	
	// 接收响应
	resp := make([]byte, 64)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("接收响应失败: %w", err)
	}
	
	return resp[:n], nil
}
