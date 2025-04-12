package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// 测试配置
var (
	serverURL = "http://localhost:8080"
	username  = "secuser"
	password  = "secpassword"
	email     = "sec@example.com"
)

// 测试前准备
func TestMain(m *testing.M) {
	// 检查服务器是否运行
	if !isServerRunning() {
		fmt.Println("服务器未运行，跳过安全测试")
		os.Exit(0)
	}

	// 注册测试用户
	registerTestUser()

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

// 注册测试用户
func registerTestUser() {
	data := map[string]string{
		"username": username,
		"password": password,
		"email":    email,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(serverURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("注册测试用户失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		fmt.Printf("注册测试用户失败: %d\n", resp.StatusCode)
	}
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

// 测试 SQL 注入
func TestSQLInjection(t *testing.T) {
	// 测试用户名中的 SQL 注入
	injectionUsernames := []string{
		"' OR '1'='1",
		"admin' --",
		"admin'; DROP TABLE users; --",
		"' UNION SELECT username, password FROM users --",
	}

	for _, injectionUsername := range injectionUsernames {
		data := map[string]string{
			"username": injectionUsername,
			"password": "password",
		}
		jsonData, _ := json.Marshal(data)

		resp, err := http.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		// 应该返回 401 Unauthorized，而不是 500 Internal Server Error
		if resp.StatusCode == http.StatusInternalServerError {
			t.Fatalf("SQL 注入可能成功: %s", injectionUsername)
		}
	}
}

// 测试 XSS 攻击
func TestXSS(t *testing.T) {
	token, err := login(username, password)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	// 测试 XSS 攻击
	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=\"x\" onerror=\"alert('XSS')\">",
		"<a href=\"javascript:alert('XSS')\">Click me</a>",
		"<svg onload=\"alert('XSS')\">",
	}

	for _, xssPayload := range xssPayloads {
		// 尝试更新用户信息
		data := map[string]string{
			"username": username,
			"email":    xssPayload,
		}
		jsonData, _ := json.Marshal(data)

		req, _ := http.NewRequest("PUT", serverURL+"/api/v1/users/me", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		// 获取用户信息
		req, _ = http.NewRequest("GET", serverURL+"/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}

		// 检查返回的数据是否包含未转义的 XSS 载荷
		email, ok := result["email"].(string)
		if ok && email == xssPayload && strings.Contains(xssPayload, "<") {
			t.Fatalf("XSS 攻击可能成功: %s", xssPayload)
		}
	}
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
}

// 测试 JWT 令牌篡改
func TestJWTTampering(t *testing.T) {
	token, err := login(username, password)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	// 篡改令牌
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("无效的 JWT 令牌格式: %s", token)
	}

	// 篡改载荷部分
	tamperedToken := parts[0] + ".eyJzdWIiOiJhZG1pbiIsInJvbGUiOiJhZG1pbiJ9." + parts[2]

	// 使用篡改的令牌
	req, _ := http.NewRequest("GET", serverURL+"/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+tamperedToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 应该返回 401 Unauthorized
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("JWT 令牌篡改保护未生效: %d", resp.StatusCode)
	}
}

// 测试密码强度
func TestPasswordStrength(t *testing.T) {
	weakPasswords := []string{
		"123456",
		"password",
		"qwerty",
		"abc123",
		"letmein",
	}

	for _, weakPassword := range weakPasswords {
		data := map[string]string{
			"username": "weakuser",
			"password": weakPassword,
			"email":    "weak@example.com",
		}
		jsonData, _ := json.Marshal(data)

		resp, err := http.Post(serverURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		// 应该返回 400 Bad Request
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			t.Fatalf("弱密码被接受: %s", weakPassword)
		}
	}
}

// 测试会话固定攻击
func TestSessionFixation(t *testing.T) {
	// 获取初始会话 Cookie
	resp, err := http.Get(serverURL + "/api/v1/health")
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	initialCookies := resp.Cookies()

	// 登录
	data := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", serverURL+"/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 添加初始 Cookie
	for _, cookie := range initialCookies {
		req.AddCookie(cookie)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 获取登录后的 Cookie
	loginCookies := resp.Cookies()

	// 检查会话 ID 是否改变
	var initialSessionID, loginSessionID string
	for _, cookie := range initialCookies {
		if cookie.Name == "session" {
			initialSessionID = cookie.Value
			break
		}
	}
	for _, cookie := range loginCookies {
		if cookie.Name == "session" {
			loginSessionID = cookie.Value
			break
		}
	}

	if initialSessionID != "" && loginSessionID != "" && initialSessionID == loginSessionID {
		t.Fatal("会话固定攻击保护未生效")
	}
}

// 测试暴力破解保护
func TestBruteForceProtection(t *testing.T) {
	// 尝试多次登录失败
	for i := 0; i < 10; i++ {
		data := map[string]string{
			"username": username,
			"password": "wrong_password",
		}
		jsonData, _ := json.Marshal(data)

		resp, err := http.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		resp.Body.Close()

		// 如果收到 429 状态码，说明暴力破解保护生效
		if resp.StatusCode == http.StatusTooManyRequests {
			return
		}

		// 短暂暂停，避免请求过快
		time.Sleep(100 * time.Millisecond)
	}

	// 再次尝试登录，应该被限制
	data := map[string]string{
		"username": username,
		"password": "wrong_password",
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatal("暴力破解保护未生效")
	}
}
