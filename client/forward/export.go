package forward

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// ExportConfig 导出配置
type ExportConfig struct {
	Version string                  `json:"version"`
	Rules   map[string]*ForwardRule `json:"rules"`
}

// ExportRules 导出规则
func ExportRules(rules map[string]*ForwardRule, filePath string) error {
	// 创建导出配置
	config := ExportConfig{
		Version: "1.0",
		Rules:   rules,
	}
	
	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	
	// 写入文件
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	
	return nil
}

// ImportRules 导入规则
func ImportRules(filePath string) (map[string]*ForwardRule, error) {
	// 读取文件
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	
	// 反序列化配置
	var config ExportConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("反序列化配置失败: %w", err)
	}
	
	// 验证版本
	if config.Version != "1.0" {
		return nil, fmt.Errorf("不支持的配置版本: %s", config.Version)
	}
	
	return config.Rules, nil
}
