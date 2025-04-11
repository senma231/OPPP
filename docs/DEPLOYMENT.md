# P3 部署指南

本文档提供了 P3 系统的详细部署说明，包括服务端和客户端的部署方法。

## 目录

- [系统要求](#系统要求)
- [服务端部署](#服务端部署)
  - [使用二进制文件部署](#使用二进制文件部署)
  - [使用 Docker 部署](#使用-docker-部署)
  - [使用 Docker Compose 部署](#使用-docker-compose-部署)
- [客户端部署](#客户端部署)
  - [Windows 客户端](#windows-客户端)
  - [Linux 客户端](#linux-客户端)
  - [macOS 客户端](#macos-客户端)
- [配置说明](#配置说明)
  - [服务端配置](#服务端配置)
  - [客户端配置](#客户端配置)
- [安全建议](#安全建议)
- [故障排除](#故障排除)

## 系统要求

### 服务端

- **操作系统**：Linux (推荐 Ubuntu 20.04+)、Windows Server 2016+、macOS 10.15+
- **CPU**：2 核或更多
- **内存**：2GB 或更多
- **存储**：10GB 或更多
- **网络**：公网 IP 地址或域名，开放相关端口

### 客户端

- **操作系统**：Windows 7+、Linux (各主流发行版)、macOS 10.13+
- **CPU**：1 核或更多
- **内存**：512MB 或更多
- **存储**：100MB 或更多
- **网络**：能够访问互联网

## 服务端部署

### 使用二进制文件部署

1. 下载最新的服务端二进制文件：

   ```bash
   wget https://github.com/senma231/OPPP/releases/latest/download/p3-server-linux-amd64.tar.gz
   tar -xzf p3-server-linux-amd64.tar.gz
   cd p3-server
   ```

2. 创建配置文件：

   ```bash
   cp config.example.yaml config.yaml
   ```

3. 编辑配置文件：

   ```bash
   nano config.yaml
   ```

   根据需要修改配置，特别是数据库连接信息和 JWT 密钥。

4. 初始化数据库：

   ```bash
   ./p3-server -init-db
   ```

5. 启动服务：

   ```bash
   ./p3-server -config config.yaml
   ```

6. (可选) 设置为系统服务：

   创建 systemd 服务文件：

   ```bash
   sudo nano /etc/systemd/system/p3-server.service
   ```

   添加以下内容：

   ```
   [Unit]
   Description=P3 Server
   After=network.target

   [Service]
   User=p3
   WorkingDirectory=/opt/p3-server
   ExecStart=/opt/p3-server/p3-server -config /opt/p3-server/config.yaml
   Restart=on-failure
   RestartSec=5

   [Install]
   WantedBy=multi-user.target
   ```

   启用并启动服务：

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable p3-server
   sudo systemctl start p3-server
   ```

### 使用 Docker 部署

1. 拉取 Docker 镜像：

   ```bash
   docker pull senma231/p3-server:latest
   ```

2. 创建配置文件目录：

   ```bash
   mkdir -p /opt/p3-server/config
   ```

3. 创建配置文件：

   ```bash
   cp config.example.yaml /opt/p3-server/config/config.yaml
   ```

4. 编辑配置文件：

   ```bash
   nano /opt/p3-server/config/config.yaml
   ```

5. 运行 Docker 容器：

   ```bash
   docker run -d \
     --name p3-server \
     -p 8080:8080 \
     -p 3478:3478/udp \
     -v /opt/p3-server/config:/app/config \
     -v /opt/p3-server/data:/app/data \
     senma231/p3-server:latest
   ```

### 使用 Docker Compose 部署

1. 创建 Docker Compose 配置文件：

   ```bash
   mkdir -p /opt/p3
   cd /opt/p3
   nano docker-compose.yml
   ```

2. 添加以下内容：

   ```yaml
   version: '3'

   services:
     postgres:
       image: postgres:14
       container_name: p3-postgres
       environment:
         POSTGRES_USER: postgres
         POSTGRES_PASSWORD: postgres
         POSTGRES_DB: p3
       volumes:
         - postgres_data:/var/lib/postgresql/data
       restart: unless-stopped

     redis:
       image: redis:6
       container_name: p3-redis
       command: redis-server --requirepass redis
       volumes:
         - redis_data:/data
       restart: unless-stopped

     server:
       image: senma231/p3-server:latest
       container_name: p3-server
       depends_on:
         - postgres
         - redis
       ports:
         - "8080:8080"
         - "3478:3478/udp"
       volumes:
         - ./config:/app/config
         - ./data:/app/data
       restart: unless-stopped

   volumes:
     postgres_data:
     redis_data:
   ```

3. 创建配置目录和配置文件：

   ```bash
   mkdir -p config
   cp config.example.yaml config/config.yaml
   ```

4. 编辑配置文件：

   ```bash
   nano config/config.yaml
   ```

   修改数据库和 Redis 连接信息：

   ```yaml
   database:
     driver: "postgres"
     host: "postgres"
     port: 5432
     user: "postgres"
     password: "postgres"
     dbname: "p3"
     sslmode: "disable"

   redis:
     host: "redis"
     port: 6379
     password: "redis"
     db: 0
   ```

5. 启动服务：

   ```bash
   docker-compose up -d
   ```

## 客户端部署

### Windows 客户端

1. 下载最新的 Windows 客户端安装程序：

   从 [Releases](https://github.com/senma231/OPPP/releases) 页面下载 `p3-client-windows-amd64.exe`。

2. 运行安装程序，按照向导完成安装。

3. 启动客户端，输入服务器地址、节点 ID 和令牌。

4. (可选) 创建配置文件：

   ```
   C:\Program Files\P3\config.yaml
   ```

   添加以下内容：

   ```yaml
   node:
     id: "your-node-id"
     token: "your-node-token"

   server:
     address: "http://your-server-address:8080"
     heartbeatInterval: 30

   network:
     enableUPnP: true
     enableNATPMP: true
     stunServers:
       - stun.l.google.com:19302
   ```

5. 以管理员身份运行客户端：

   ```
   "C:\Program Files\P3\p3-client.exe" -config "C:\Program Files\P3\config.yaml"
   ```

### Linux 客户端

1. 下载最新的 Linux 客户端：

   ```bash
   wget https://github.com/senma231/OPPP/releases/latest/download/p3-client-linux-amd64.tar.gz
   tar -xzf p3-client-linux-amd64.tar.gz
   cd p3-client
   ```

2. 创建配置文件：

   ```bash
   cp config.example.yaml config.yaml
   ```

3. 编辑配置文件：

   ```bash
   nano config.yaml
   ```

4. 运行客户端：

   ```bash
   ./p3-client -config config.yaml
   ```

5. (可选) 设置为系统服务：

   创建 systemd 服务文件：

   ```bash
   sudo nano /etc/systemd/system/p3-client.service
   ```

   添加以下内容：

   ```
   [Unit]
   Description=P3 Client
   After=network.target

   [Service]
   User=p3
   WorkingDirectory=/opt/p3-client
   ExecStart=/opt/p3-client/p3-client -config /opt/p3-client/config.yaml
   Restart=on-failure
   RestartSec=5

   [Install]
   WantedBy=multi-user.target
   ```

   启用并启动服务：

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable p3-client
   sudo systemctl start p3-client
   ```

### macOS 客户端

1. 下载最新的 macOS 客户端：

   从 [Releases](https://github.com/senma231/OPPP/releases) 页面下载 `p3-client-darwin-amd64.tar.gz` 或 `p3-client-darwin-arm64.tar.gz`（适用于 Apple Silicon）。

2. 解压文件：

   ```bash
   tar -xzf p3-client-darwin-amd64.tar.gz
   cd p3-client
   ```

3. 创建配置文件：

   ```bash
   cp config.example.yaml config.yaml
   ```

4. 编辑配置文件：

   ```bash
   nano config.yaml
   ```

5. 运行客户端：

   ```bash
   ./p3-client -config config.yaml
   ```

6. (可选) 创建 LaunchAgent：

   ```bash
   mkdir -p ~/Library/LaunchAgents
   nano ~/Library/LaunchAgents/com.p3.client.plist
   ```

   添加以下内容：

   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
   <plist version="1.0">
   <dict>
       <key>Label</key>
       <string>com.p3.client</string>
       <key>ProgramArguments</key>
       <array>
           <string>/opt/p3-client/p3-client</string>
           <string>-config</string>
           <string>/opt/p3-client/config.yaml</string>
       </array>
       <key>RunAtLoad</key>
       <true/>
       <key>KeepAlive</key>
       <true/>
       <key>StandardErrorPath</key>
       <string>/tmp/p3-client.err</string>
       <key>StandardOutPath</key>
       <string>/tmp/p3-client.out</string>
   </dict>
   </plist>
   ```

   加载 LaunchAgent：

   ```bash
   launchctl load ~/Library/LaunchAgents/com.p3.client.plist
   ```

## 配置说明

### 服务端配置

服务端配置文件 `config.yaml` 的主要参数说明：

| 参数 | 说明 | 默认值 |
|-----|------|-------|
| server.host | 服务器监听地址 | 0.0.0.0 |
| server.port | 服务器监听端口 | 8080 |
| database.driver | 数据库驱动 | postgres |
| database.host | 数据库主机 | localhost |
| database.port | 数据库端口 | 5432 |
| database.user | 数据库用户名 | postgres |
| database.password | 数据库密码 | postgres |
| database.dbname | 数据库名称 | p3 |
| redis.host | Redis 主机 | localhost |
| redis.port | Redis 端口 | 6379 |
| redis.password | Redis 密码 | - |
| jwt.secret | JWT 密钥 | - |
| jwt.expireTime | JWT 过期时间（小时） | 24 |
| p2p.udpPort1 | P2P UDP 端口 1 | 27182 |
| p2p.udpPort2 | P2P UDP 端口 2 | 27183 |
| p2p.tcpPort | P2P TCP 端口 | 27184 |
| relay.maxBandwidth | 中继最大带宽（Mbps） | 10 |
| relay.maxClients | 中继最大客户端数 | 100 |
| log.level | 日志级别 | info |
| log.output | 日志输出 | stdout |
| log.file | 日志文件路径 | p3-server.log |
| turn.address | TURN 服务器地址 | 0.0.0.0:3478 |
| turn.realm | TURN 服务器域 | p3.example.com |
| turn.authSecret | TURN 服务器认证密钥 | - |

### 客户端配置

客户端配置文件 `config.yaml` 的主要参数说明：

| 参数 | 说明 | 默认值 |
|-----|------|-------|
| node.id | 节点 ID | - |
| node.token | 节点令牌 | - |
| server.address | 服务器地址 | http://localhost:8080 |
| server.heartbeatInterval | 心跳间隔（秒） | 30 |
| network.enableUPnP | 启用 UPnP | true |
| network.enableNATPMP | 启用 NAT-PMP | true |
| network.stunServers | STUN 服务器列表 | stun.l.google.com:19302 |
| security.enableTLS | 启用 TLS | true |
| security.certFile | 证书文件路径 | cert.pem |
| security.keyFile | 密钥文件路径 | key.pem |
| logging.level | 日志级别 | info |
| logging.file | 日志文件路径 | p3-client.log |

## 安全建议

1. **更改默认密钥**：
   - 修改 JWT 密钥
   - 修改 TURN 认证密钥
   - 修改数据库密码
   - 修改 Redis 密码

2. **使用 HTTPS**：
   - 配置 SSL 证书
   - 使用反向代理（如 Nginx）

3. **防火墙配置**：
   - 只开放必要的端口
   - 限制 IP 访问

4. **定期更新**：
   - 保持系统和软件包更新
   - 定期更新 P3 到最新版本

5. **监控和日志**：
   - 启用详细日志
   - 设置日志轮转
   - 监控系统资源使用情况

## 故障排除

### 服务端问题

1. **服务无法启动**：
   - 检查配置文件是否正确
   - 检查端口是否被占用
   - 检查日志文件中的错误信息

2. **数据库连接失败**：
   - 检查数据库服务是否运行
   - 检查数据库连接信息是否正确
   - 检查数据库用户权限

3. **API 请求失败**：
   - 检查服务器是否正常运行
   - 检查网络连接
   - 检查 JWT 令牌是否有效

### 客户端问题

1. **无法连接到服务器**：
   - 检查服务器地址是否正确
   - 检查网络连接
   - 检查服务器是否在线

2. **认证失败**：
   - 检查节点 ID 和令牌是否正确
   - 检查节点是否已在服务器上注册

3. **NAT 穿透失败**：
   - 检查 STUN 服务器是否可访问
   - 检查防火墙设置
   - 尝试使用 TURN 中继

4. **端口转发失败**：
   - 检查端口是否被占用
   - 检查目标主机和端口是否正确
   - 检查防火墙设置
