package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config 服务端配置结构
type Config struct {
	Version string `yaml:"version"`
	Server  struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	JWT struct {
		Secret     string `yaml:"secret"`
		ExpireTime int    `yaml:"expireTime"` // 单位：小时
	} `yaml:"jwt"`
	P2P struct {
		UDPPort1 int `yaml:"udpPort1"`
		UDPPort2 int `yaml:"udpPort2"`
		TCPPort  int `yaml:"tcpPort"`
	} `yaml:"p2p"`
	Relay struct {
		MaxBandwidth int `yaml:"maxBandwidth"` // 单位：Mbps
		MaxClients   int `yaml:"maxClients"`
	} `yaml:"relay"`
	Log struct {
		Level  string `yaml:"level"`
		Output string `yaml:"output"`
	} `yaml:"log"`
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
	config.Server.Host = "0.0.0.0"
	config.Server.Port = 8080

	config.Database.Host = "localhost"
	config.Database.Port = 5432
	config.Database.User = "postgres"
	config.Database.Password = "postgres"
	config.Database.DBName = "p3"
	config.Database.SSLMode = "disable"

	config.Redis.Host = "localhost"
	config.Redis.Port = 6379
	config.Redis.Password = ""
	config.Redis.DB = 0

	config.JWT.Secret = "p3_secret_key"
	config.JWT.ExpireTime = 24

	config.P2P.UDPPort1 = 27182
	config.P2P.UDPPort2 = 27183
	config.P2P.TCPPort = 27184

	config.Relay.MaxBandwidth = 10
	config.Relay.MaxClients = 100

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
