package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// NodeConfig 节点配置
type NodeConfig struct {
	ID    string `yaml:"id"`
	Token string `yaml:"token"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Address           string `yaml:"address"`
	HeartbeatInterval int    `yaml:"heartbeatInterval"` // 单位：秒
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	EnableUPnP   bool     `yaml:"enableUPnP"`
	EnableNATPMP bool     `yaml:"enableNATPMP"`
	STUNServers  []string `yaml:"stunServers"`
	TURNServers  []struct {
		Address  string `yaml:"address"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"turnServers"`
	UDPPort1 int `yaml:"udpPort1"`
	UDPPort2 int `yaml:"udpPort2"`
	TCPPort  int `yaml:"tcpPort"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableTLS bool   `yaml:"enableTLS"`
	CertFile  string `yaml:"certFile"`
	KeyFile   string `yaml:"keyFile"`
	CAFile    string `yaml:"caFile"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	MaxConnections    int `yaml:"maxConnections"`
	ConnectionTimeout int `yaml:"connectionTimeout"`
	KeepAliveInterval int `yaml:"keepAliveInterval"`
	BufferSize        int `yaml:"bufferSize"`
	BandwidthLimit    struct {
		Upload   int `yaml:"upload"`
		Download int `yaml:"download"`
	} `yaml:"bandwidthLimit"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `yaml:"name"`
	Protocol    string `yaml:"protocol"` // tcp, udp
	SrcPort     int    `yaml:"srcPort"`
	PeerNode    string `yaml:"peerNode"`
	DstPort     int    `yaml:"dstPort"`
	DstHost     string `yaml:"dstHost"`
	Description string `yaml:"description"`
	AutoStart   bool   `yaml:"autoStart"`
}

// Config 客户端配置
type Config struct {
	Node        NodeConfig        `yaml:"node"`
	Server      ServerConfig      `yaml:"server"`
	Network     NetworkConfig     `yaml:"network"`
	Security    SecurityConfig    `yaml:"security"`
	Logging     LoggingConfig     `yaml:"logging"`
	Performance PerformanceConfig `yaml:"performance"`
	Apps        []AppConfig       `yaml:"apps"`
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	// 加载默认配置
	config := DefaultConfig()

	// 读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		// 如果文件不存在，使用默认配置
		if os.IsNotExist(err) {
			fmt.Printf("配置文件 %s 不存在，使用默认配置\n", path)
			return config, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置文件
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 从环境变量加载配置
	loadFromEnv(config)

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return config, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Node: NodeConfig{
			ID:    "my-node",
			Token: "your-node-token",
		},
		Server: ServerConfig{
			Address:           "http://localhost:8080",
			HeartbeatInterval: 30,
		},
		Network: NetworkConfig{
			EnableUPnP:   true,
			EnableNATPMP: true,
			STUNServers: []string{
				"stun.l.google.com:19302",
				"stun.stunprotocol.org:3478",
			},
			TURNServers: []struct {
				Address  string `yaml:"address"`
				Username string `yaml:"username"`
				Password string `yaml:"password"`
			}{
				{
					Address:  "turn.example.com:3478",
					Username: "username",
					Password: "password",
				},
			},
			UDPPort1: 27182,
			UDPPort2: 27183,
			TCPPort:  27184,
		},
		Security: SecurityConfig{
			EnableTLS: true,
			CertFile:  "cert.pem",
			KeyFile:   "key.pem",
			CAFile:    "ca.pem",
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  "p3-client.log",
		},
		Performance: PerformanceConfig{
			MaxConnections:    100,
			ConnectionTimeout: 30,
			KeepAliveInterval: 15,
			BufferSize:        4096,
			BandwidthLimit: struct {
				Upload   int `yaml:"upload"`
				Download int `yaml:"download"`
			}{
				Upload:   10,
				Download: 10,
			},
		},
		Apps: []AppConfig{},
	}
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, path string) error {
	// 序列化配置
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 添加注释
	header := []byte("# P3 客户端配置文件\n")
	data = append(header, data...)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(config *Config) {
	// 节点配置
	if id := os.Getenv("P3_NODE_ID"); id != "" {
		config.Node.ID = id
	}
	if token := os.Getenv("P3_NODE_TOKEN"); token != "" {
		config.Node.Token = token
	}

	// 服务器配置
	if address := os.Getenv("P3_SERVER_ADDRESS"); address != "" {
		config.Server.Address = address
	}
	if interval := os.Getenv("P3_SERVER_HEARTBEAT_INTERVAL"); interval != "" {
		if i, err := strconv.Atoi(interval); err == nil {
			config.Server.HeartbeatInterval = i
		}
	}

	// 网络配置
	if upnp := os.Getenv("P3_NETWORK_ENABLE_UPNP"); upnp != "" {
		config.Network.EnableUPnP = strings.ToLower(upnp) == "true"
	}
	if natpmp := os.Getenv("P3_NETWORK_ENABLE_NATPMP"); natpmp != "" {
		config.Network.EnableNATPMP = strings.ToLower(natpmp) == "true"
	}
	if stunServers := os.Getenv("P3_NETWORK_STUN_SERVERS"); stunServers != "" {
		config.Network.STUNServers = strings.Split(stunServers, ",")
	}

	// 安全配置
	if enableTLS := os.Getenv("P3_SECURITY_ENABLE_TLS"); enableTLS != "" {
		config.Security.EnableTLS = strings.ToLower(enableTLS) == "true"
	}
	if certFile := os.Getenv("P3_SECURITY_CERT_FILE"); certFile != "" {
		config.Security.CertFile = certFile
	}
	if keyFile := os.Getenv("P3_SECURITY_KEY_FILE"); keyFile != "" {
		config.Security.KeyFile = keyFile
	}
	if caFile := os.Getenv("P3_SECURITY_CA_FILE"); caFile != "" {
		config.Security.CAFile = caFile
	}

	// 日志配置
	if level := os.Getenv("P3_LOGGING_LEVEL"); level != "" {
		config.Logging.Level = level
	}
	if file := os.Getenv("P3_LOGGING_FILE"); file != "" {
		config.Logging.File = file
	}
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证节点配置
	if config.Node.ID == "" {
		return errors.New("节点 ID 不能为空")
	}
	if config.Node.Token == "" {
		return errors.New("节点令牌不能为空")
	}

	// 验证服务器配置
	if config.Server.Address == "" {
		return errors.New("服务器地址不能为空")
	}
	if config.Server.HeartbeatInterval <= 0 {
		return errors.New("心跳间隔必须大于 0")
	}

	// 验证网络配置
	if len(config.Network.STUNServers) == 0 {
		return errors.New("STUN 服务器列表不能为空")
	}

	// 验证安全配置
	if config.Security.EnableTLS {
		if config.Security.CertFile == "" {
			return errors.New("启用 TLS 时证书文件不能为空")
		}
		if config.Security.KeyFile == "" {
			return errors.New("启用 TLS 时密钥文件不能为空")
		}
	}

	// 验证日志配置
	if config.Logging.Level == "" {
		return errors.New("日志级别不能为空")
	}

	// 验证应用配置
	for i, app := range config.Apps {
		if app.Name == "" {
			return fmt.Errorf("应用 %d 的名称不能为空", i+1)
		}
		if app.Protocol != "tcp" && app.Protocol != "udp" {
			return fmt.Errorf("应用 %s 的协议必须为 tcp 或 udp", app.Name)
		}
		if app.SrcPort <= 0 || app.SrcPort > 65535 {
			return fmt.Errorf("应用 %s 的源端口无效", app.Name)
		}
		if app.PeerNode == "" {
			return fmt.Errorf("应用 %s 的对等节点不能为空", app.Name)
		}
		if app.DstPort <= 0 || app.DstPort > 65535 {
			return fmt.Errorf("应用 %s 的目标端口无效", app.Name)
		}
		if app.DstHost == "" {
			return fmt.Errorf("应用 %s 的目标主机不能为空", app.Name)
		}
	}

	return nil
}
