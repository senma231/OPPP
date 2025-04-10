package nat

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/huin/goupnp/dcps/internetgateway2"
)

// UPnPClient UPnP 客户端
type UPnPClient struct {
	Timeout time.Duration
}

// NewUPnPClient 创建 UPnP 客户端
func NewUPnPClient(timeout time.Duration) *UPnPClient {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &UPnPClient{
		Timeout: timeout,
	}
}

// AddPortMapping 添加端口映射
func (c *UPnPClient) AddPortMapping(
	externalPort int,
	internalPort int,
	protocol string,
	description string,
) (bool, string, error) {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	// 获取本地 IP
	localIP, err := getLocalIP()
	if err != nil {
		return false, "", fmt.Errorf("获取本地 IP 失败: %w", err)
	}

	// 尝试 IGDv2
	success, externalIP, err := c.addPortMappingIGDv2(
		ctx, externalPort, internalPort, localIP.String(), protocol, description,
	)
	if err == nil && success {
		return true, externalIP, nil
	}

	// 尝试 IGDv1
	success, externalIP, err = c.addPortMappingIGDv1(
		ctx, externalPort, internalPort, localIP.String(), protocol, description,
	)
	if err == nil && success {
		return true, externalIP, nil
	}

	return false, "", fmt.Errorf("添加端口映射失败: %w", err)
}

// addPortMappingIGDv2 使用 IGDv2 添加端口映射
func (c *UPnPClient) addPortMappingIGDv2(
	ctx context.Context,
	externalPort int,
	internalPort int,
	internalClient string,
	protocol string,
	description string,
) (bool, string, error) {
	// 发现 IGDv2 设备
	clients, _, err := internetgateway2.NewWANIPConnection2ClientsCtx(ctx)
	if err != nil {
		return false, "", err
	}

	for _, client := range clients {
		// 获取外部 IP
		externalIP, err := client.GetExternalIPAddressCtx(ctx)
		if err != nil {
			continue
		}

		// 添加端口映射
		err = client.AddPortMappingCtx(
			ctx,
			"",                // 远程主机（空表示任意）
			uint16(externalPort),
			protocol,
			uint16(internalPort),
			internalClient,
			description,
			true,             // 启用
			uint32(86400),    // 租期（秒）
		)
		if err != nil {
			continue
		}

		return true, externalIP, nil
	}

	// 尝试 WANIPConnection1
	clients1, _, err := internetgateway2.NewWANIPConnection1ClientsCtx(ctx)
	if err != nil {
		return false, "", err
	}

	for _, client := range clients1 {
		// 获取外部 IP
		externalIP, err := client.GetExternalIPAddressCtx(ctx)
		if err != nil {
			continue
		}

		// 添加端口映射
		err = client.AddPortMappingCtx(
			ctx,
			"",                // 远程主机（空表示任意）
			uint16(externalPort),
			protocol,
			uint16(internalPort),
			internalClient,
			description,
			true,             // 启用
			uint32(86400),    // 租期（秒）
		)
		if err != nil {
			continue
		}

		return true, externalIP, nil
	}

	// 尝试 WANPPPConnection1
	pppClients, _, err := internetgateway2.NewWANPPPConnection1ClientsCtx(ctx)
	if err != nil {
		return false, "", err
	}

	for _, client := range pppClients {
		// 获取外部 IP
		externalIP, err := client.GetExternalIPAddressCtx(ctx)
		if err != nil {
			continue
		}

		// 添加端口映射
		err = client.AddPortMappingCtx(
			ctx,
			"",                // 远程主机（空表示任意）
			uint16(externalPort),
			protocol,
			uint16(internalPort),
			internalClient,
			description,
			true,             // 启用
			uint32(86400),    // 租期（秒）
		)
		if err != nil {
			continue
		}

		return true, externalIP, nil
	}

	return false, "", fmt.Errorf("没有找到可用的 IGDv2 设备")
}

// addPortMappingIGDv1 使用 IGDv1 添加端口映射
func (c *UPnPClient) addPortMappingIGDv1(
	ctx context.Context,
	externalPort int,
	internalPort int,
	internalClient string,
	protocol string,
	description string,
) (bool, string, error) {
	// 发现 IGDv1 设备
	clients, _, err := internetgateway1.NewWANIPConnection1ClientsCtx(ctx)
	if err != nil {
		return false, "", err
	}

	for _, client := range clients {
		// 获取外部 IP
		externalIP, err := client.GetExternalIPAddressCtx(ctx)
		if err != nil {
			continue
		}

		// 添加端口映射
		err = client.AddPortMappingCtx(
			ctx,
			"",                // 远程主机（空表示任意）
			uint16(externalPort),
			protocol,
			uint16(internalPort),
			internalClient,
			description,
			true,             // 启用
			uint32(86400),    // 租期（秒）
		)
		if err != nil {
			continue
		}

		return true, externalIP, nil
	}

	// 尝试 WANPPPConnection1
	pppClients, _, err := internetgateway1.NewWANPPPConnection1ClientsCtx(ctx)
	if err != nil {
		return false, "", err
	}

	for _, client := range pppClients {
		// 获取外部 IP
		externalIP, err := client.GetExternalIPAddressCtx(ctx)
		if err != nil {
			continue
		}

		// 添加端口映射
		err = client.AddPortMappingCtx(
			ctx,
			"",                // 远程主机（空表示任意）
			uint16(externalPort),
			protocol,
			uint16(internalPort),
			internalClient,
			description,
			true,             // 启用
			uint32(86400),    // 租期（秒）
		)
		if err != nil {
			continue
		}

		return true, externalIP, nil
	}

	return false, "", fmt.Errorf("没有找到可用的 IGDv1 设备")
}

// DeletePortMapping 删除端口映射
func (c *UPnPClient) DeletePortMapping(externalPort int, protocol string) error {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	// 尝试 IGDv2
	err := c.deletePortMappingIGDv2(ctx, externalPort, protocol)
	if err == nil {
		return nil
	}

	// 尝试 IGDv1
	err = c.deletePortMappingIGDv1(ctx, externalPort, protocol)
	if err == nil {
		return nil
	}

	return fmt.Errorf("删除端口映射失败: %w", err)
}

// deletePortMappingIGDv2 使用 IGDv2 删除端口映射
func (c *UPnPClient) deletePortMappingIGDv2(ctx context.Context, externalPort int, protocol string) error {
	// 发现 IGDv2 设备
	clients, _, err := internetgateway2.NewWANIPConnection2ClientsCtx(ctx)
	if err != nil {
		return err
	}

	for _, client := range clients {
		// 删除端口映射
		err = client.DeletePortMappingCtx(
			ctx,
			"",                // 远程主机（空表示任意）
			uint16(externalPort),
			protocol,
		)
		if err == nil {
			return nil
		}
	}

	// 尝试 WANIPConnection1
	clients1, _, err := internetgateway2.NewWANIPConnection1ClientsCtx(ctx)
	if err != nil {
		return err
	}

	for _, client := range clients1 {
		// 删除端口映射
		err = client.DeletePortMappingCtx(
			ctx,
			"",                // 远程主机（空表示任意）
			uint16(externalPort),
			protocol,
		)
		if err == nil {
			return nil
		}
	}

	// 尝试 WANPPPConnection1
	pppClients, _, err := internetgateway2.NewWANPPPConnection1ClientsCtx(ctx)
	if err != nil {
		return err
	}

	for _, client := range pppClients {
		// 删除端口映射
		err = client.DeletePortMappingCtx(
			ctx,
			"",                // 远程主机（空表示任意）
			uint16(externalPort),
			protocol,
		)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("没有找到可用的 IGDv2 设备")
}

// deletePortMappingIGDv1 使用 IGDv1 删除端口映射
func (c *UPnPClient) deletePortMappingIGDv1(ctx context.Context, externalPort int, protocol string) error {
	// 发现 IGDv1 设备
	clients, _, err := internetgateway1.NewWANIPConnection1ClientsCtx(ctx)
	if err != nil {
		return err
	}

	for _, client := range clients {
		// 删除端口映射
		err = client.DeletePortMappingCtx(
			ctx,
			"",                // 远程主机（空表示任意）
			uint16(externalPort),
			protocol,
		)
		if err == nil {
			return nil
		}
	}

	// 尝试 WANPPPConnection1
	pppClients, _, err := internetgateway1.NewWANPPPConnection1ClientsCtx(ctx)
	if err != nil {
		return err
	}

	for _, client := range pppClients {
		// 删除端口映射
		err = client.DeletePortMappingCtx(
			ctx,
			"",                // 远程主机（空表示任意）
			uint16(externalPort),
			protocol,
		)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("没有找到可用的 IGDv1 设备")
}

// GetExternalIP 获取外部 IP
func (c *UPnPClient) GetExternalIP() (string, error) {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	// 尝试 IGDv2
	clients, _, err := internetgateway2.NewWANIPConnection2ClientsCtx(ctx)
	if err == nil && len(clients) > 0 {
		for _, client := range clients {
			externalIP, err := client.GetExternalIPAddressCtx(ctx)
			if err == nil {
				return externalIP, nil
			}
		}
	}

	// 尝试 IGDv1
	clients1, _, err := internetgateway1.NewWANIPConnection1ClientsCtx(ctx)
	if err == nil && len(clients1) > 0 {
		for _, client := range clients1 {
			externalIP, err := client.GetExternalIPAddressCtx(ctx)
			if err == nil {
				return externalIP, nil
			}
		}
	}

	return "", fmt.Errorf("获取外部 IP 失败")
}

// DiscoverGateways 发现网关设备
func (c *UPnPClient) DiscoverGateways() ([]string, error) {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	// 发现设备
	devices, err := goupnp.DiscoverDevicesCtx(ctx, internetgateway2.URN_WANIPConnection_2)
	if err != nil {
		return nil, err
	}

	// 提取设备信息
	var gateways []string
	for _, device := range devices {
		if device.Root != nil && device.Root.Device != nil {
			gateways = append(gateways, device.Root.Device.FriendlyName)
		}
	}

	return gateways, nil
}

// IsUPnPAvailable 检查 UPnP 是否可用
func (c *UPnPClient) IsUPnPAvailable() bool {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	// 尝试 IGDv2
	clients, _, err := internetgateway2.NewWANIPConnection2ClientsCtx(ctx)
	if err == nil && len(clients) > 0 {
		return true
	}

	// 尝试 IGDv1
	clients1, _, err := internetgateway1.NewWANIPConnection1ClientsCtx(ctx)
	if err == nil && len(clients1) > 0 {
		return true
	}

	return false
}
