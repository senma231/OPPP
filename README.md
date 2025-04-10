# P3 - P2P 端口转发平台

P3（P2P Port Proxy Platform）是一个基于 P2P 技术的端口转发平台，支持内网穿透、P2P 打洞和端口转发管理。该项目参考了 [openp2p](https://github.com/openp2p-cn/openp2p) 开源项目的核心功能，并进行了重新设计和实现。

## 主要特性

- **P2P 网络连接**：支持 NAT 穿透（NAT1-NAT4，包括 Cone 和 Symmetric）、UDP 和 TCP 打洞、UPNP、IPv6
- **安全通信**：TLS1.3 + AES 双重加密，TOTP 一次性密码授权
- **跨平台支持**：支持 Windows、Linux、Linux Arm、Mac OS、Mac OS Arm
- **端口转发管理**：通过 Web 管理平台配置和管理端口转发规则
- **内网穿透**：内网只需一台服务器和公网服务器组网，其他机器无需安装客户端即可被访问

## 系统架构

系统由三个主要部分组成：

1. **服务端**：负责用户认证、设备管理、P2P 连接协调和中转服务
2. **客户端**：实现 NAT 穿透、端口转发、加密通信等核心功能
3. **管理平台**：提供 Web 界面，方便管理设备、应用和端口转发规则

## 项目状态

该项目目前处于开发阶段，按照 [开发计划](./开发计划.md) 进行实施。

## 技术栈

- **后端**：Go、Gin、PostgreSQL、Redis、NATS
- **前端**：React、TypeScript、Ant Design、Redux、ECharts
- **客户端**：Go（核心）、Electron（GUI，可选）
- **部署**：Docker、Docker Compose

## 目录结构

```
project-root/
├── server/                  # 服务端代码
├── client/                  # 客户端代码
├── web/                     # Web 管理平台
├── scripts/                 # 脚本工具
├── docs/                    # 文档
└── docker/                  # Docker 配置
```

## 开发进度

请参考 [开发计划](./开发计划.md) 了解详细的开发进度和里程碑。

## 快速开始

### 环境要求

- Go 1.20+
- PostgreSQL 14+
- Docker & Docker Compose (可选)

### 构建和运行

1. 克隆仓库

```bash
git clone https://github.com/senma231/p3.git
cd p3
```

2. 初始化环境

```bash
make setup
```

3. 构建服务端和客户端

```bash
make all
```

4. 运行服务端

```bash
./bin/p3-server -config server/config.yaml
```

5. 运行客户端

```bash
./bin/p3-client -node YOUR_NODE_NAME -token YOUR_TOKEN
```

### 使用 Docker 运行

```bash
docker-compose up -d
```

访问 Web 管理平台：http://localhost

## 贡献

欢迎贡献代码、报告问题或提出改进建议。

## 许可证

MIT
