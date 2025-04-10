package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config 客户端配置结构
type Config struct {
	Version string `yaml:"version"`
	Network struct {
		Node           string `yaml:"node"`
		Token          string `yaml:"token"`
		ShareBandwidth int    `yaml:"shareBandwidth"` // 单位：Mbps，0表示不共享
		ServerHost     string `yaml:"serverHost"`
		ServerPort     int    `yaml:"serverPort"`
		UDPPort1       int    `yaml:"udpPort1"`
		UDPPort2       int    `yaml:"udpPort2"`
	} `yaml:"network"`
	Apps []AppConfig `yaml:"apps"`
	Log  struct {
		Level  string `yaml:"level"`
		Output string `yaml:"output"`
	} `yaml:"log"`
}

// AppConfig P2P应用配置
type AppConfig struct {
	AppName  string `yaml:"appName"`
	Protocol string `yaml:"protocol"` // tcp, udp
	SrcPort  int    `yaml:"srcPort"`
	PeerNode string `yaml:"peerNode"`
	DstPort  int    `yaml:"dstPort"`
	DstHost  string `yaml:"dstHost"`
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	config := &Config{
		Version: "0.1.0",
	}
	config.Network.ServerHost = "api.p3.example.com"
	config.Network.ServerPort = 8080
	config.Network.UDPPort1 = 27182
	config.Network.UDPPort2 = 27183
	config.Network.ShareBandwidth = 10

	config.Log.Level = "info"
	config.Log.Output = "stdout"

	return config
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}
