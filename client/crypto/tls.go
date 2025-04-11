package crypto

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"
)

// TLSConfig TLS 配置
type TLSConfig struct {
	CertFile   string
	KeyFile    string
	CAFile     string
	ServerName string
	SkipVerify bool
}

// CreateTLSConfig 创建 TLS 配置
func CreateTLSConfig(config *TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: config.ServerName,
	}

	// 如果跳过验证
	if config.SkipVerify {
		tlsConfig.InsecureSkipVerify = true
		return tlsConfig, nil
	}

	// 如果提供了 CA 文件，加载 CA 证书
	if config.CAFile != "" {
		caCert, err := ioutil.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("读取 CA 证书失败: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("解析 CA 证书失败")
		}

		tlsConfig.RootCAs = caCertPool
	}

	// 如果提供了客户端证书和密钥，加载客户端证书
	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("加载客户端证书失败: %w", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

// WrapConn 将普通连接包装为 TLS 连接
func WrapConn(conn net.Conn, config *tls.Config, isClient bool) (net.Conn, error) {
	if isClient {
		return tls.Client(conn, config), nil
	}
	return tls.Server(conn, config), nil
}

// CreateTLSListener 创建 TLS 监听器
func CreateTLSListener(network, address string, config *tls.Config) (net.Listener, error) {
	listener, err := tls.Listen(network, address, config)
	if err != nil {
		return nil, fmt.Errorf("创建 TLS 监听器失败: %w", err)
	}
	return listener, nil
}

// DialTLS 使用 TLS 连接到远程地址
func DialTLS(network, address string, config *tls.Config) (net.Conn, error) {
	conn, err := tls.Dial(network, address, config)
	if err != nil {
		return nil, fmt.Errorf("TLS 连接失败: %w", err)
	}
	return conn, nil
}
