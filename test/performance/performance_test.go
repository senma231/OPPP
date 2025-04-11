package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"
)

const (
	apiURL = "http://localhost:8080/api/v1"
)

// TestPerformance 性能测试
func TestPerformance(t *testing.T) {
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

	// 创建应用
	appID := createApp(t, token, device1ID, device2ID)
	if appID == "" {
		t.Fatal("创建应用失败")
	}

	// 等待应用启动
	time.Sleep(5 * time.Second)

	// 测试带宽
	testBandwidth(t, "localhost", 23389)

	// 测试延迟
	testLatency(t, "localhost", 23389)

	// 测试并发连接
	testConcurrentConnections(t, "localhost", 23389)

	// 清理
	deleteApp(t, token, appID)
	deleteDevice(t, token, device1ID)
	deleteDevice(t, token, device2ID)
}

// registerAndLogin 注册并登录
func registerAndLogin(t *testing.T) string {
	// 注册
	registerData := map[string]string{
		"username": "perfuser",
		"password": "perfpassword",
		"email":    "perf@example.com",
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
		"username": "perfuser",
		"password": "perfpassword",
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

// createApp 创建应用
func createApp(t *testing.T, token, sourceDeviceID, targetDeviceID string) string {
	appData := map[string]interface{}{
		"name":        "PerfApp",
		"deviceId":    sourceDeviceID,
		"protocol":    "tcp",
		"srcPort":     23389,
		"peerNode":    "client2",
		"dstPort":     3389,
		"dstHost":     "localhost",
		"description": "Performance Test Application",
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

// testBandwidth 测试带宽
func testBandwidth(t *testing.T, host string, port int) {
	// 建立连接
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	// 测试参数
	dataSize := 10 * 1024 * 1024 // 10MB
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// 测试上传带宽
	startTime := time.Now()
	n, err := conn.Write(data)
	if err != nil {
		t.Fatalf("写入数据失败: %v", err)
	}
	duration := time.Since(startTime)
	uploadBandwidth := float64(n) / duration.Seconds() / 1024 / 1024 // MB/s

	t.Logf("上传带宽: %.2f MB/s", uploadBandwidth)

	// 测试下载带宽
	startTime = time.Now()
	received := 0
	buffer := make([]byte, 4096)
	for received < dataSize {
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("读取数据失败: %v", err)
		}
		received += n
	}
	duration = time.Since(startTime)
	downloadBandwidth := float64(received) / duration.Seconds() / 1024 / 1024 // MB/s

	t.Logf("下载带宽: %.2f MB/s", downloadBandwidth)
}

// testLatency 测试延迟
func testLatency(t *testing.T, host string, port int) {
	// 测试参数
	testCount := 100
	data := []byte("ping")
	totalLatency := time.Duration(0)

	for i := 0; i < testCount; i++ {
		// 建立连接
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		// 测量延迟
		startTime := time.Now()
		_, err = conn.Write(data)
		if err != nil {
			t.Fatalf("写入数据失败: %v", err)
		}

		buffer := make([]byte, 4)
		_, err = conn.Read(buffer)
		if err != nil {
			t.Fatalf("读取数据失败: %v", err)
		}

		latency := time.Since(startTime)
		totalLatency += latency

		conn.Close()
	}

	avgLatency := totalLatency / time.Duration(testCount)
	t.Logf("平均延迟: %v", avgLatency)
}

// testConcurrentConnections 测试并发连接
func testConcurrentConnections(t *testing.T, host string, port int) {
	// 测试参数
	concurrentCount := 100
	successCount := 0
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < concurrentCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// 建立连接
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
			if err != nil {
				t.Logf("连接失败: %v", err)
				return
			}
			defer conn.Close()

			// 发送和接收数据
			data := []byte("test")
			_, err = conn.Write(data)
			if err != nil {
				t.Logf("写入数据失败: %v", err)
				return
			}

			buffer := make([]byte, 4)
			_, err = conn.Read(buffer)
			if err != nil {
				t.Logf("读取数据失败: %v", err)
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}()
	}

	wg.Wait()
	t.Logf("并发连接成功率: %d/%d (%.2f%%)", successCount, concurrentCount, float64(successCount)/float64(concurrentCount)*100)
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
