package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

// 测试配置
var (
	serverURL = "http://localhost:8080"
	username  = "testuser"
	password  = "testpassword"
	email     = "test@example.com"
)

// 测试前准备
func TestMain(m *testing.M) {
	// 检查服务器是否运行
	if !isServerRunning() {
		fmt.Println("服务器未运行，跳过集成测试")
		os.Exit(0)
	}

	// 运行测试
	code := m.Run()

	// 清理测试数据
	cleanupTestData()

	os.Exit(code)
}

// 检查服务器是否运行
func isServerRunning() bool {
	resp, err := http.Get(serverURL + "/api/v1/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// 清理测试数据
func cleanupTestData() {
	// 登录获取令牌
	token, err := login(username, password)
	if err != nil {
		return
	}

	// 删除测试用户
	req, _ := http.NewRequest("DELETE", serverURL+"/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)
}

// 登录获取令牌
func login(username, password string) (string, error) {
	data := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("登录失败: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("无法获取访问令牌")
	}

	return token, nil
}

// 测试注册
func TestRegister(t *testing.T) {
	data := map[string]string{
		"username": username,
		"password": password,
		"email":    email,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(serverURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("注册请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("注册失败: %d", resp.StatusCode)
	}
}

// 测试登录
func TestLogin(t *testing.T) {
	token, err := login(username, password)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	if token == "" {
		t.Fatal("获取的令牌为空")
	}
}

// 测试获取用户信息
func TestGetUserInfo(t *testing.T) {
	token, err := login(username, password)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	req, _ := http.NewRequest("GET", serverURL+"/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("获取用户信息请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("获取用户信息失败: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if result["username"] != username {
		t.Fatalf("用户名不匹配: 期望 %s, 实际 %s", username, result["username"])
	}
}

// 测试刷新令牌
func TestRefreshToken(t *testing.T) {
	// 登录获取令牌
	data := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("登录请求失败: %v", err)
	}
	defer resp.Body.Close()

	var loginResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&loginResult); err != nil {
		t.Fatalf("解析登录响应失败: %v", err)
	}

	refreshToken, ok := loginResult["refresh_token"].(string)
	if !ok {
		t.Fatal("无法获取刷新令牌")
	}

	// 刷新令牌
	refreshData := map[string]string{
		"refresh_token": refreshToken,
	}
	refreshJsonData, _ := json.Marshal(refreshData)

	refreshResp, err := http.Post(serverURL+"/api/v1/auth/refresh", "application/json", bytes.NewBuffer(refreshJsonData))
	if err != nil {
		t.Fatalf("刷新令牌请求失败: %v", err)
	}
	defer refreshResp.Body.Close()

	if refreshResp.StatusCode != http.StatusOK {
		t.Fatalf("刷新令牌失败: %d", refreshResp.StatusCode)
	}

	var refreshResult map[string]interface{}
	if err := json.NewDecoder(refreshResp.Body).Decode(&refreshResult); err != nil {
		t.Fatalf("解析刷新令牌响应失败: %v", err)
	}

	newToken, ok := refreshResult["access_token"].(string)
	if !ok || newToken == "" {
		t.Fatal("无法获取新的访问令牌")
	}
}

// 测试添加设备
func TestAddDevice(t *testing.T) {
	token, err := login(username, password)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	data := map[string]string{
		"name":        "Test Device",
		"description": "Integration test device",
	}
	jsonData, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", serverURL+"/api/v1/devices", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("添加设备请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("添加设备失败: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	deviceID, ok := result["id"].(string)
	if !ok || deviceID == "" {
		t.Fatal("无法获取设备 ID")
	}
}

// 测试速率限制
func TestRateLimit(t *testing.T) {
	// 发送多个请求测试速率限制
	for i := 0; i < 20; i++ {
		resp, err := http.Get(serverURL + "/api/v1/health")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		resp.Body.Close()

		// 如果收到 429 状态码，说明速率限制生效
		if resp.StatusCode == http.StatusTooManyRequests {
			return
		}

		// 短暂暂停，避免请求过快
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal("速率限制未生效")
}

// 测试 CSRF 保护
func TestCSRFProtection(t *testing.T) {
	// 获取 CSRF 令牌
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(serverURL + "/api/v1/csrf")
	if err != nil {
		t.Fatalf("获取 CSRF 令牌请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 获取 CSRF Cookie
	cookies := resp.Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "csrf_token" {
			csrfCookie = cookie
			break
		}
	}

	if csrfCookie == nil {
		t.Fatal("未找到 CSRF Cookie")
	}

	// 尝试不带 CSRF 令牌的 POST 请求
	data := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", serverURL+"/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 添加 Cookie
	req.AddCookie(csrfCookie)

	// 不添加 CSRF 令牌头
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 应该返回 403 Forbidden
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("CSRF 保护未生效: %d", resp.StatusCode)
	}

	// 尝试带 CSRF 令牌的 POST 请求
	req, _ = http.NewRequest("POST", serverURL+"/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrfCookie.Value)
	req.AddCookie(csrfCookie)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 应该返回 200 OK
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("带 CSRF 令牌的请求失败: %d", resp.StatusCode)
	}
}
