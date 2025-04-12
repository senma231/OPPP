package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

// 测试配置
var (
	serverURL = "http://localhost:8080"
	username  = "perfuser"
	password  = "perfpassword"
	email     = "perf@example.com"
)

// 测试前准备
func TestMain(m *testing.M) {
	// 检查服务器是否运行
	if !isServerRunning() {
		fmt.Println("服务器未运行，跳过性能测试")
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

// 测试登录性能
func BenchmarkLogin(b *testing.B) {
	data := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			b.Fatalf("登录请求失败: %v", err)
		}
		resp.Body.Close()
	}
}

// 测试获取用户信息性能
func BenchmarkGetUserInfo(b *testing.B) {
	token, err := login(username, password)
	if err != nil {
		b.Fatalf("登录失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", serverURL+"/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatalf("获取用户信息请求失败: %v", err)
		}
		resp.Body.Close()
	}
}

// 测试并发登录性能
func BenchmarkConcurrentLogin(b *testing.B) {
	data := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(data)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				b.Fatalf("登录请求失败: %v", err)
			}
			resp.Body.Close()
		}
	})
}

// 测试并发获取用户信息性能
func BenchmarkConcurrentGetUserInfo(b *testing.B) {
	token, err := login(username, password)
	if err != nil {
		b.Fatalf("登录失败: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", serverURL+"/api/v1/users/me", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Fatalf("获取用户信息请求失败: %v", err)
			}
			resp.Body.Close()
		}
	})
}

// 测试高并发请求
func TestHighConcurrency(t *testing.T) {
	token, err := login(username, password)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	// 并发数
	concurrency := 100
	// 每个并发请求的次数
	requestsPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 记录响应时间
	responseTimes := make([]time.Duration, concurrency*requestsPerGoroutine)
	var mutex sync.Mutex

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()

			for j := 0; j < requestsPerGoroutine; j++ {
				start := time.Now()

				req, _ := http.NewRequest("GET", serverURL+"/api/v1/users/me", nil)
				req.Header.Set("Authorization", "Bearer "+token)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Logf("请求失败: %v", err)
					continue
				}
				resp.Body.Close()

				duration := time.Since(start)

				mutex.Lock()
				responseTimes[index*requestsPerGoroutine+j] = duration
				mutex.Unlock()
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// 计算统计信息
	var totalTime time.Duration
	var minTime = time.Hour
	var maxTime time.Duration
	var count int

	for _, duration := range responseTimes {
		if duration > 0 {
			totalTime += duration
			count++

			if duration < minTime {
				minTime = duration
			}
			if duration > maxTime {
				maxTime = duration
			}
		}
	}

	if count == 0 {
		t.Fatal("没有成功的请求")
	}

	avgTime := totalTime / time.Duration(count)
	rps := float64(count) / totalDuration.Seconds()

	t.Logf("并发数: %d", concurrency)
	t.Logf("总请求数: %d", count)
	t.Logf("总耗时: %v", totalDuration)
	t.Logf("平均响应时间: %v", avgTime)
	t.Logf("最小响应时间: %v", minTime)
	t.Logf("最大响应时间: %v", maxTime)
	t.Logf("每秒请求数 (RPS): %.2f", rps)
}

// 测试长时间运行
func TestLongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过长时间运行测试")
	}

	token, err := login(username, password)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	// 测试持续时间
	duration := 5 * time.Minute
	// 请求间隔
	interval := 1 * time.Second

	// 记录成功和失败的请求数
	var successCount, failureCount int
	var mutex sync.Mutex

	// 开始时间
	startTime := time.Now()
	endTime := startTime.Add(duration)

	t.Logf("开始长时间运行测试，持续 %v", duration)

	for time.Now().Before(endTime) {
		req, _ := http.NewRequest("GET", serverURL+"/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			mutex.Lock()
			failureCount++
			mutex.Unlock()
			t.Logf("请求失败: %v", err)
		} else {
			resp.Body.Close()
			mutex.Lock()
			successCount++
			mutex.Unlock()
		}

		time.Sleep(interval)
	}

	totalDuration := time.Since(startTime)
	totalRequests := successCount + failureCount
	successRate := float64(successCount) / float64(totalRequests) * 100
	rps := float64(totalRequests) / totalDuration.Seconds()

	t.Logf("测试完成")
	t.Logf("总请求数: %d", totalRequests)
	t.Logf("成功请求数: %d", successCount)
	t.Logf("失败请求数: %d", failureCount)
	t.Logf("成功率: %.2f%%", successRate)
	t.Logf("每秒请求数 (RPS): %.2f", rps)
}
