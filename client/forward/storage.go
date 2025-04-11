package forward

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// RuleStorage 规则存储
type RuleStorage struct {
	filePath string
	mu       sync.RWMutex
}

// NewRuleStorage 创建规则存储
func NewRuleStorage(filePath string) *RuleStorage {
	return &RuleStorage{
		filePath: filePath,
	}
}

// SaveRules 保存规则
func (s *RuleStorage) SaveRules(rules map[string]*ForwardRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 创建目录
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	
	// 序列化规则
	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化规则失败: %w", err)
	}
	
	// 写入文件
	if err := ioutil.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	
	return nil
}

// LoadRules 加载规则
func (s *RuleStorage) LoadRules() (map[string]*ForwardRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// 检查文件是否存在
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return make(map[string]*ForwardRule), nil
	}
	
	// 读取文件
	data, err := ioutil.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	
	// 反序列化规则
	var rules map[string]*ForwardRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("反序列化规则失败: %w", err)
	}
	
	return rules, nil
}
