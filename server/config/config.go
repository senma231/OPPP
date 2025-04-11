package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `yaml:"driver"` // postgres, mysql, sqlite3
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireTime int    `yaml:"expireTime"` // 单位：小时
}

// P2PConfig P2P 配置
type P2PConfig struct {
	UDPPort1 int `yaml:"udpPort1"`
	UDPPort2 int `yaml:"udpPort2"`
	TCPPort  int `yaml:"tcpPort"`
}

// RelayConfig 中继配置
type RelayConfig struct {
	MaxBandwidth int `yaml:"maxBandwidth"` // 单位：Mbps
	MaxClients   int `yaml:"maxClients"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Output string `yaml:"output"` // stdout, file
	File   string `yaml:"file"`   // 日志文件路径
}

// TURNConfig TURN 服务器配置
type TURNConfig struct {
	Address    string `yaml:"address"`
	Realm      string `yaml:"realm"`
	AuthSecret string `yaml:"authSecret"`
}

// Config 服务端配置结构
type Config struct {
	Version  string         `yaml:"version"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	P2P      P2PConfig      `yaml:"p2p"`
	Relay    RelayConfig    `yaml:"relay"`
	Log      LogConfig      `yaml:"log"`
	TURN     TURNConfig     `yaml:"turn"`
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
		Version: "0.1.0",
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "p3",
			SSLMode:  "disable",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		JWT: JWTConfig{
			Secret:     "p3_secret_key",
			ExpireTime: 24,
		},
		P2P: P2PConfig{
			UDPPort1: 27182,
			UDPPort2: 27183,
			TCPPort:  27184,
		},
		Relay: RelayConfig{
			MaxBandwidth: 10,
			MaxClients:   100,
		},
		Log: LogConfig{
			Level:  "info",
			Output: "stdout",
			File:   "p3-server.log",
		},
		TURN: TURNConfig{
			Address:    "0.0.0.0:3478",
			Realm:      "p3.example.com",
			AuthSecret: "p3_turn_secret",
		},
	}
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(config *Config) {
	// 服务器配置
	if host := os.Getenv("P3_SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("P3_SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	// 数据库配置
	if driver := os.Getenv("P3_DB_DRIVER"); driver != "" {
		config.Database.Driver = driver
	}
	if host := os.Getenv("P3_DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("P3_DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}
	if user := os.Getenv("P3_DB_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("P3_DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if dbname := os.Getenv("P3_DB_NAME"); dbname != "" {
		config.Database.DBName = dbname
	}
	if sslmode := os.Getenv("P3_DB_SSLMODE"); sslmode != "" {
		config.Database.SSLMode = sslmode
	}

	// Redis 配置
	if host := os.Getenv("P3_REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}
	if port := os.Getenv("P3_REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Redis.Port = p
		}
	}
	if password := os.Getenv("P3_REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}
	if db := os.Getenv("P3_REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			config.Redis.DB = d
		}
	}

	// JWT 配置
	if secret := os.Getenv("P3_JWT_SECRET"); secret != "" {
		config.JWT.Secret = secret
	}
	if expireTime := os.Getenv("P3_JWT_EXPIRE_TIME"); expireTime != "" {
		if t, err := strconv.Atoi(expireTime); err == nil {
			config.JWT.ExpireTime = t
		}
	}

	// P2P 配置
	if udpPort1 := os.Getenv("P3_P2P_UDP_PORT1"); udpPort1 != "" {
		if p, err := strconv.Atoi(udpPort1); err == nil {
			config.P2P.UDPPort1 = p
		}
	}
	if udpPort2 := os.Getenv("P3_P2P_UDP_PORT2"); udpPort2 != "" {
		if p, err := strconv.Atoi(udpPort2); err == nil {
			config.P2P.UDPPort2 = p
		}
	}
	if tcpPort := os.Getenv("P3_P2P_TCP_PORT"); tcpPort != "" {
		if p, err := strconv.Atoi(tcpPort); err == nil {
			config.P2P.TCPPort = p
		}
	}

	// 中继配置
	if maxBandwidth := os.Getenv("P3_RELAY_MAX_BANDWIDTH"); maxBandwidth != "" {
		if b, err := strconv.Atoi(maxBandwidth); err == nil {
			config.Relay.MaxBandwidth = b
		}
	}
	if maxClients := os.Getenv("P3_RELAY_MAX_CLIENTS"); maxClients != "" {
		if c, err := strconv.Atoi(maxClients); err == nil {
			config.Relay.MaxClients = c
		}
	}

	// 日志配置
	if level := os.Getenv("P3_LOG_LEVEL"); level != "" {
		config.Log.Level = level
	}
	if output := os.Getenv("P3_LOG_OUTPUT"); output != "" {
		config.Log.Output = output
	}
	if file := os.Getenv("P3_LOG_FILE"); file != "" {
		config.Log.File = file
	}

	// TURN 配置
	if address := os.Getenv("P3_TURN_ADDRESS"); address != "" {
		config.TURN.Address = address
	}
	if realm := os.Getenv("P3_TURN_REALM"); realm != "" {
		config.TURN.Realm = realm
	}
	if authSecret := os.Getenv("P3_TURN_AUTH_SECRET"); authSecret != "" {
		config.TURN.AuthSecret = authSecret
	}
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return errors.New("服务器端口无效")
	}

	// 验证数据库配置
	if config.Database.Driver == "" {
		return errors.New("数据库驱动不能为空")
	}
	if config.Database.Driver != "sqlite3" {
		if config.Database.Host == "" {
			return errors.New("数据库主机不能为空")
		}
		if config.Database.Port <= 0 || config.Database.Port > 65535 {
			return errors.New("数据库端口无效")
		}
		if config.Database.User == "" {
			return errors.New("数据库用户不能为空")
		}
		if config.Database.DBName == "" {
			return errors.New("数据库名不能为空")
		}
	}

	// 验证 JWT 配置
	if config.JWT.Secret == "" {
		return errors.New("JWT 密钥不能为空")
	}
	if config.JWT.ExpireTime <= 0 {
		return errors.New("JWT 过期时间无效")
	}

	// 验证 P2P 配置
	if config.P2P.UDPPort1 <= 0 || config.P2P.UDPPort1 > 65535 {
		return errors.New("P2P UDP 端口 1 无效")
	}
	if config.P2P.UDPPort2 <= 0 || config.P2P.UDPPort2 > 65535 {
		return errors.New("P2P UDP 端口 2 无效")
	}
	if config.P2P.TCPPort <= 0 || config.P2P.TCPPort > 65535 {
		return errors.New("P2P TCP 端口无效")
	}

	// 验证中继配置
	if config.Relay.MaxBandwidth <= 0 {
		return errors.New("中继最大带宽无效")
	}
	if config.Relay.MaxClients <= 0 {
		return errors.New("中继最大客户端数无效")
	}

	// 验证日志配置
	logLevel := strings.ToLower(config.Log.Level)
	if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" {
		return errors.New("日志级别无效")
	}
	logOutput := strings.ToLower(config.Log.Output)
	if logOutput != "stdout" && logOutput != "file" {
		return errors.New("日志输出类型无效")
	}
	if logOutput == "file" && config.Log.File == "" {
		return errors.New("日志文件路径不能为空")
	}

	// 验证 TURN 配置
	if config.TURN.Address == "" {
		return errors.New("TURN 服务器地址不能为空")
	}
	if config.TURN.Realm == "" {
		return errors.New("TURN 服务器域不能为空")
	}
	if config.TURN.AuthSecret == "" {
		return errors.New("TURN 服务器认证密钥不能为空")
	}

	return nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	switch c.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.User, c.Password, c.Host, c.Port, c.DBName)
	case "sqlite3":
		return c.DBName
	default:
		return ""
	}
}
