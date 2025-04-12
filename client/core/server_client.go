package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/senma231/p3/client/config"
	"github.com/senma231/p3/client/nat"
	"github.com/senma231/p3/common/logger"
)

// ServerClient 服务器客户端
type ServerClient struct {
	config  *config.Config
	natInfo *nat.NATInfo
	client  *http.Client
}

// NewServerClient 创建服务器客户端
func NewServerClient(cfg *config.Config, natInfo *nat.NATInfo) *ServerClient {
	return &ServerClient{
		config:  cfg,
		natInfo: natInfo,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Register 注册设备
func (c *ServerClient) Register() error {
	// 如果已有节点 ID 和令牌，则不需要注册
	if c.config.Node.ID != "" && c.config.Node.Token != "" {
		return nil
	}

	// 创建注册请求
	reqBody := map[string]interface{}{
		"name": c.config.Node.Name,
	}

	// 发送请求
	resp, err := c.post("/api/v1/devices", reqBody)
	if err != nil {
		return fmt.Errorf("注册设备失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusCreated {
		errMsg := "未知错误"
		if errObj, ok := result["error"]; ok {
			errMsg = fmt.Sprintf("%v", errObj)
		}
		return fmt.Errorf("注册设备失败: %s", errMsg)
	}

	// 提取节点 ID 和令牌
	nodeID, ok := result["nodeId"].(string)
	if !ok {
		return fmt.Errorf("响应中缺少节点 ID")
	}

	token, ok := result["token"].(string)
	if !ok {
		return fmt.Errorf("响应中缺少令牌")
	}

	// 更新配置
	c.config.Node.ID = nodeID
	c.config.Node.Token = token

	// 保存配置
	if err := config.SaveConfig(c.config, "config.yaml"); err != nil {
		logger.Error("保存配置失败: %v", err)
	}

	return nil
}

// Heartbeat 发送心跳
func (c *ServerClient) Heartbeat() error {
	// 创建心跳请求
	reqBody := map[string]interface{}{
		"status":     "online",
		"natType":    c.natInfo.Type.String(),
		"externalIP": c.natInfo.ExternalIP.String(),
		"localIP":    c.natInfo.LocalIP.String(),
		"version":    "1.0.0",
		"os":         getOS(),
		"arch":       getArch(),
	}

	// 发送请求
	resp, err := c.post("/api/v1/device/status", reqBody)
	if err != nil {
		return fmt.Errorf("发送心跳失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}

		errMsg := "未知错误"
		if errObj, ok := result["error"]; ok {
			errMsg = fmt.Sprintf("%v", errObj)
		}
		return fmt.Errorf("发送心跳失败: %s", errMsg)
	}

	return nil
}

// GetPeerInfo 获取对等节点信息
func (c *ServerClient) GetPeerInfo(peerNodeID string) (*PeerInfo, error) {
	// 发送请求
	resp, err := c.get(fmt.Sprintf("/api/v1/devices/%s", peerNodeID))
	if err != nil {
		return nil, fmt.Errorf("获取对等节点信息失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		errMsg := "未知错误"
		if errObj, ok := result["error"]; ok {
			errMsg = fmt.Sprintf("%v", errObj)
		}
		return nil, fmt.Errorf("获取对等节点信息失败: %s", errMsg)
	}

	// 提取对等节点信息
	nodeID, ok := result["nodeId"].(string)
	if !ok {
		return nil, fmt.Errorf("响应中缺少节点 ID")
	}

	natTypeStr, ok := result["natType"].(string)
	if !ok {
		return nil, fmt.Errorf("响应中缺少 NAT 类型")
	}

	externalIP, ok := result["externalIP"].(string)
	if !ok {
		return nil, fmt.Errorf("响应中缺少外部 IP")
	}

	status, ok := result["status"].(string)
	if !ok {
		return nil, fmt.Errorf("响应中缺少状态")
	}

	// 检查节点是否在线
	if status != "online" {
		return nil, fmt.Errorf("对等节点不在线")
	}

	// 解析 NAT 类型
	var natType nat.NATType
	switch natTypeStr {
	case "No NAT (Public IP)":
		natType = nat.NATNone
	case "Full Cone NAT":
		natType = nat.NATFull
	case "Restricted Cone NAT":
		natType = nat.NATRestricted
	case "Port Restricted Cone NAT":
		natType = nat.NATPortRestricted
	case "Symmetric NAT":
		natType = nat.NATSymmetric
	default:
		natType = nat.NATUnknown
	}

	// 创建对等节点信息
	peerInfo := &PeerInfo{
		NodeID:       nodeID,
		NATType:      natType,
		ExternalIP:   externalIP,
		ExternalPort: 27182, // 默认端口
		LastSeen:     time.Now(),
	}

	return peerInfo, nil
}

// GetRelayServer 获取中继服务器
func (c *ServerClient) GetRelayServer() (string, error) {
	// 发送请求
	resp, err := c.get("/api/v1/relay/server")
	if err != nil {
		return "", fmt.Errorf("获取中继服务器失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		errMsg := "未知错误"
		if errObj, ok := result["error"]; ok {
			errMsg = fmt.Sprintf("%v", errObj)
		}
		return "", fmt.Errorf("获取中继服务器失败: %s", errMsg)
	}

	// 提取中继服务器地址
	server, ok := result["server"].(string)
	if !ok {
		return "", fmt.Errorf("响应中缺少服务器地址")
	}

	return server, nil
}

// GetApps 获取应用列表
func (c *ServerClient) GetApps() ([]config.AppConfig, error) {
	// 发送请求
	resp, err := c.get("/api/v1/device/apps")
	if err != nil {
		return nil, fmt.Errorf("获取应用列表失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		errMsg := "未知错误"
		if errObj, ok := result["error"]; ok {
			errMsg = fmt.Sprintf("%v", errObj)
		}
		return nil, fmt.Errorf("获取应用列表失败: %s", errMsg)
	}

	// 提取应用列表
	appsData, ok := result["apps"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应中缺少应用列表")
	}

	// 解析应用列表
	apps := make([]config.AppConfig, 0, len(appsData))
	for _, appData := range appsData {
		appMap, ok := appData.(map[string]interface{})
		if !ok {
			continue
		}

		app := config.AppConfig{
			Name:        getString(appMap, "name", ""),
			Protocol:    getString(appMap, "protocol", "tcp"),
			SrcPort:     getInt(appMap, "srcPort", 0),
			PeerNode:    getString(appMap, "peerNode", ""),
			DstPort:     getInt(appMap, "dstPort", 0),
			DstHost:     getString(appMap, "dstHost", ""),
			Description: getString(appMap, "description", ""),
			AutoStart:   getBool(appMap, "status", "running"),
		}

		apps = append(apps, app)
	}

	return apps, nil
}

// get 发送 GET 请求
func (c *ServerClient) get(path string) (*http.Response, error) {
	// 创建请求
	req, err := http.NewRequest(http.MethodGet, c.config.Server.Address+path, nil)
	if err != nil {
		return nil, err
	}

	// 添加认证头
	req.Header.Set("X-Node-ID", c.config.Node.ID)
	req.Header.Set("X-Node-Token", c.config.Node.Token)

	// 发送请求
	return c.client.Do(req)
}

// post 发送 POST 请求
func (c *ServerClient) post(path string, body interface{}) (*http.Response, error) {
	// 序列化请求体
	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// 创建请求
	req, err := http.NewRequest(http.MethodPost, c.config.Server.Address+path, bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, err
	}

	// 添加认证头
	req.Header.Set("X-Node-ID", c.config.Node.ID)
	req.Header.Set("X-Node-Token", c.config.Node.Token)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	return c.client.Do(req)
}

// put 发送 PUT 请求
func (c *ServerClient) put(path string, body interface{}) (*http.Response, error) {
	// 序列化请求体
	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// 创建请求
	req, err := http.NewRequest(http.MethodPut, c.config.Server.Address+path, bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, err
	}

	// 添加认证头
	req.Header.Set("X-Node-ID", c.config.Node.ID)
	req.Header.Set("X-Node-Token", c.config.Node.Token)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	return c.client.Do(req)
}

// delete 发送 DELETE 请求
func (c *ServerClient) delete(path string) (*http.Response, error) {
	// 创建请求
	req, err := http.NewRequest(http.MethodDelete, c.config.Server.Address+path, nil)
	if err != nil {
		return nil, err
	}

	// 添加认证头
	req.Header.Set("X-Node-ID", c.config.Node.ID)
	req.Header.Set("X-Node-Token", c.config.Node.Token)

	// 发送请求
	return c.client.Do(req)
}

// getString 从 map 中获取字符串
func getString(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}

// getInt 从 map 中获取整数
func getInt(m map[string]interface{}, key string, defaultValue int) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

// getBool 从 map 中获取布尔值
func getBool(m map[string]interface{}, key, trueValue string) bool {
	if val, ok := m[key].(string); ok {
		return val == trueValue
	}
	return false
}

// getOS 获取操作系统
func getOS() string {
	return "unknown"
}

// getArch 获取架构
func getArch() string {
	return "unknown"
}
