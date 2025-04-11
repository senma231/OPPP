package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const (
	apiURL = "http://localhost:8080/api/v1"
)

// TestConnection 测试连接
func TestConnection(t *testing.T) {
	// 注册用户
	token := registerAndLogin(t)
	if token == "" {
		t.Fatal("注册和登录失败")
	}

	// 添加设备
	device1ID := addDevice(t, token, "client1", "test-token1")
	if device1ID == "" {
		t.Fatal("添加设备 1 失败")
	}

	device2ID := addDevice(t, token, "client2", "test-token2")
	if device2ID == "" {
		t.Fatal("添加设备 2 失败")
	}

	// 等待设备上线
	time.Sleep(5 * time.Second)

	// 检查设备状态
	if !checkDeviceStatus(t, token, device1ID, "online") {
		t.Fatal("设备 1 未上线")
	}

	if !checkDeviceStatus(t, token, device2ID, "online") {
		t.Fatal("设备 2 未上线")
	}

	// 创建应用
	appID := createApp(t, token, device1ID, device2ID)
	if appID == "" {
		t.Fatal("创建应用失败")
	}

	// 等待应用启动
	time.Sleep(5 * time.Second)

	// 检查应用状态
	if !checkAppStatus(t, token, appID, "running") {
		t.Fatal("应用未运行")
	}

	// 测试连接
	if !testAppConnection(t, "localhost", 23389) {
		t.Fatal("应用连接测试失败")
	}

	// 清理
	deleteApp(t, token, appID)
	deleteDevice(t, token, device1ID)
	deleteDevice(t, token, device2ID)
}

// registerAndLogin 注册并登录
func registerAndLogin(t *testing.T) string {
	// 注册
	registerData := map[string]string{
		"username": "testuser",
		"password": "testpassword",
		"email":    "test@example.com",
	}
	registerJSON, _ := json.Marshal(registerData)

	resp, err := http.Post(apiURL+"/auth/register", "application/json", bytes.NewBuffer(registerJSON))
	if err != nil {
		t.Fatalf("注册请求失败: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("注册失败: %s", body)
		return ""
	}

	// 登录
	loginData := map[string]string{
		"username": "testuser",
		"password": "testpassword",
	}
	loginJSON, _ := json.Marshal(loginData)

	resp, err = http.Post(apiURL+"/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		t.Fatalf("登录请求失败: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("登录失败: %s", body)
		return ""
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析登录响应失败: %v", err)
		return ""
	}

	token, ok := result["token"].(string)
	if !ok {
		t.Fatal("登录响应中没有 token")
		return ""
	}

	return token
}

// addDevice 添加设备
func addDevice(t *testing.T, token, name, nodeToken string) string {
	deviceData := map[string]string{
		"name":  name,
		"token": nodeToken,
	}
	deviceJSON, _ := json.Marshal(deviceData)

	req, _ := http.NewRequest("POST", apiURL+"/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("添加设备请求失败: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("添加设备失败: %s", body)
		return ""
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析添加设备响应失败: %v", err)
		return ""
	}

	deviceID, ok := result["id"].(string)
	if !ok {
		t.Fatal("添加设备响应中没有 id")
		return ""
	}

	return deviceID
}

// checkDeviceStatus 检查设备状态
func checkDeviceStatus(t *testing.T, token, deviceID, expectedStatus string) bool {
	req, _ := http.NewRequest("GET", apiURL+"/devices/"+deviceID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("获取设备请求失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("获取设备失败: %s", body)
		return false
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析获取设备响应失败: %v", err)
		return false
	}

	status, ok := result["status"].(string)
	if !ok {
		t.Fatal("获取设备响应中没有 status")
		return false
	}

	return status == expectedStatus
}

// createApp 创建应用
func createApp(t *testing.T, token, sourceDeviceID, targetDeviceID string) string {
	appData := map[string]interface{}{
		"name":        "TestApp",
		"deviceId":    sourceDeviceID,
		"protocol":    "tcp",
		"srcPort":     23389,
		"peerNode":    "client2",
		"dstPort":     3389,
		"dstHost":     "localhost",
		"description": "Test Application",
	}
	appJSON, _ := json.Marshal(appData)

	req, _ := http.NewRequest("POST", apiURL+"/apps", bytes.NewBuffer(appJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("创建应用请求失败: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("创建应用失败: %s", body)
		return ""
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析创建应用响应失败: %v", err)
		return ""
	}

	appID, ok := result["id"].(string)
	if !ok {
		t.Fatal("创建应用响应中没有 id")
		return ""
	}

	return appID
}

// checkAppStatus 检查应用状态
func checkAppStatus(t *testing.T, token, appID, expectedStatus string) bool {
	req, _ := http.NewRequest("GET", apiURL+"/apps/"+appID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("获取应用请求失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("获取应用失败: %s", body)
		return false
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析获取应用响应失败: %v", err)
		return false
	}

	status, ok := result["status"].(string)
	if !ok {
		t.Fatal("获取应用响应中没有 status")
		return false
	}

	return status == expectedStatus
}

// testAppConnection 测试应用连接
func testAppConnection(t *testing.T, host string, port int) bool {
	// 这里简化为尝试建立 TCP 连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 5*time.Second)
	if err != nil {
		t.Logf("连接失败: %v", err)
		return false
	}
	defer conn.Close()
	return true
}

// deleteApp 删除应用
func deleteApp(t *testing.T, token, appID string) {
	req, _ := http.NewRequest("DELETE", apiURL+"/apps/"+appID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("删除应用请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("删除应用失败: %s", body)
	}
}

// deleteDevice 删除设备
func deleteDevice(t *testing.T, token, deviceID string) {
	req, _ := http.NewRequest("DELETE", apiURL+"/devices/"+deviceID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("删除设备请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("删除设备失败: %s", body)
	}
}
