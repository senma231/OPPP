# P3 开发文档

本文档提供了 P3 项目的开发指南，包括架构设计、模块说明、API 接口和二次开发指南。

## 目录

1. [架构设计](#架构设计)
2. [模块说明](#模块说明)
3. [开发环境](#开发环境)
4. [构建与部署](#构建与部署)
5. [二次开发](#二次开发)

## 架构设计

P3 系统由三个主要部分组成：

1. **服务端**：负责用户认证、设备管理、P2P 连接协调和中转服务
2. **客户端**：实现 NAT 穿透、端口转发、加密通信等核心功能
3. **管理平台**：提供 Web 界面，方便管理设备、应用和端口转发规则

### 服务端架构

服务端采用微服务架构，包含以下服务：

- **认证服务**：负责用户认证和授权
- **设备服务**：负责设备管理和状态监控
- **P2P 协调服务**：负责协调 P2P 连接的建立
- **中转服务**：当无法建立直接连接时提供数据中转
- **API 服务**：提供 RESTful API 接口

### 客户端架构

客户端采用模块化设计，包含以下模块：

- **核心引擎**：负责管理 P2P 连接和通信
- **NAT 穿透模块**：实现各种 NAT 穿透技术
- **端口转发模块**：管理本地端口映射
- **加密通信模块**：实现 TLS 和 AES 加密
- **本地服务模块**：管理客户端服务生命周期

### 管理平台架构

管理平台采用前后端分离架构：

- **前端**：使用 React + TypeScript 构建的单页应用
- **后端**：提供 RESTful API 接口的服务

## 模块说明

### 服务端模块

#### 认证服务

认证服务负责用户认证和授权，主要功能包括：

- 用户注册、登录和登出
- JWT Token 生成和验证
- 权限管理

#### 设备服务

设备服务负责设备管理和状态监控，主要功能包括：

- 设备注册和认证
- 设备状态监控
- 设备分组管理

#### P2P 协调服务

P2P 协调服务负责协调 P2P 连接的建立，主要功能包括：

- NAT 类型检测
- 打洞协调
- 中转节点选择

#### 中转服务

中转服务负责在无法建立直接连接时提供数据中转，主要功能包括：

- 数据中转
- 带宽管理
- 流量统计

#### API 服务

API 服务提供 RESTful API 接口，主要功能包括：

- 设备管理 API
- 应用管理 API
- 端口转发管理 API
- 系统状态 API

### 客户端模块

#### 核心引擎

核心引擎负责管理 P2P 连接和通信，主要功能包括：

- 连接管理
- 数据传输
- 错误处理

#### NAT 穿透模块

NAT 穿透模块实现各种 NAT 穿透技术，主要功能包括：

- STUN 协议实现
- UDP 打洞
- TCP 打洞
- UPnP 支持

#### 端口转发模块

端口转发模块管理本地端口映射，主要功能包括：

- 本地端口监听
- 数据转发
- 转发规则管理

#### 加密通信模块

加密通信模块实现 TLS 和 AES 加密，主要功能包括：

- TLS 1.3 实现
- AES 加密
- TOTP 一次性密码

#### 本地服务模块

本地服务模块管理客户端服务生命周期，主要功能包括：

- 服务安装和卸载
- 服务启动和停止
- 日志管理

### 管理平台模块

#### 前端模块

前端模块使用 React + TypeScript 构建，主要功能包括：

- 用户界面
- 状态管理
- API 调用

#### 后端模块

后端模块提供 RESTful API 接口，主要功能包括：

- 请求处理
- 数据验证
- 响应生成

## 开发环境

### 环境要求

- Go 1.18+
- Node.js 16+
- PostgreSQL 14+
- Redis 6+
- Docker & Docker Compose

### 环境搭建

1. 克隆仓库

```bash
git clone https://github.com/senma231/p3.git
cd p3
```

2. 安装依赖

```bash
# 服务端依赖
cd server
go mod download

# 客户端依赖
cd ../client
go mod download

# 前端依赖
cd ../web/frontend
npm install
```

3. 配置开发环境

```bash
# 复制配置文件
cp server/config.example.yaml server/config.yaml
cp client/config.example.yaml client/config.yaml

# 编辑配置文件
# 根据需要修改配置
```

## 构建与部署

### 构建

使用提供的构建脚本进行构建：

```bash
cd scripts/build
./build.sh
```

构建产物将输出到 `bin` 目录。

### 部署

使用 Docker Compose 进行部署：

```bash
cd scripts/deploy
./deploy.sh
```

或者手动部署：

```bash
# 启动服务端
cd server
./p3-server -config config.yaml

# 启动客户端
cd client
./p3-client -node YOUR_NODE_NAME -token YOUR_TOKEN
```

## 二次开发

### 扩展服务端

要扩展服务端功能，可以按照以下步骤进行：

1. 在 `server` 目录下创建新的模块
2. 实现新功能
3. 在 `cmd/main.go` 中注册新模块
4. 构建并测试

### 扩展客户端

要扩展客户端功能，可以按照以下步骤进行：

1. 在 `client` 目录下创建新的模块
2. 实现新功能
3. 在 `cmd/main.go` 中注册新模块
4. 构建并测试

### 扩展管理平台

要扩展管理平台功能，可以按照以下步骤进行：

1. 在 `web/frontend/src/pages` 目录下创建新的页面
2. 在 `web/frontend/src/App.tsx` 中注册新页面
3. 在 `web/backend` 中实现相应的 API
4. 构建并测试

### API 集成

要集成 P3 API，可以使用以下方式：

1. 直接调用 RESTful API
2. 使用提供的客户端库（待开发）
3. 使用 WebSocket 接口实时获取数据
