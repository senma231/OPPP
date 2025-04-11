# P3 用户手册

欢迎使用 P3（P2P Port Proxy Platform）！本手册将指导您如何使用 P3 系统进行端口转发和 P2P 连接。

## 目录

- [概述](#概述)
- [安装](#安装)
- [基本概念](#基本概念)
- [命令行界面](#命令行界面)
- [Web 管理界面](#web-管理界面)
- [使用场景](#使用场景)
- [故障排除](#故障排除)
- [常见问题](#常见问题)

## 概述

P3 是一个基于 P2P 技术的端口转发平台，支持内网穿透、P2P 打洞和端口转发管理。它允许您在不同网络环境下轻松建立点对点连接，实现远程访问和资源共享。

主要功能包括：

- P2P 网络连接
- 端口转发管理
- NAT 穿透
- 安全通信
- 跨平台支持

## 安装

### Windows

1. 从 [Releases](https://github.com/senma231/OPPP/releases) 页面下载最新的 Windows 安装程序。
2. 运行安装程序，按照向导完成安装。
3. 启动 P3 客户端。

### macOS

1. 从 [Releases](https://github.com/senma231/OPPP/releases) 页面下载最新的 macOS 安装包。
2. 打开安装包，将应用程序拖到应用程序文件夹。
3. 启动 P3 客户端。

### Linux

1. 从 [Releases](https://github.com/senma231/OPPP/releases) 页面下载最新的 Linux 安装包。
2. 解压安装包：
   ```bash
   tar -xzf p3-client-linux-amd64.tar.gz
   ```
3. 运行客户端：
   ```bash
   cd p3-client
   ./p3-client
   ```

## 基本概念

在使用 P3 之前，了解以下基本概念将有助于您更好地理解系统：

### 节点 (Node)

节点是指运行 P3 客户端的设备。每个节点都有一个唯一的 ID 和令牌，用于在 P3 网络中标识自己。

### 应用 (App)

应用是指端口转发规则，定义了如何将本地端口的流量转发到远程节点的端口。

### NAT 类型

NAT（网络地址转换）类型决定了设备与互联网的连接方式。P3 支持多种 NAT 类型，包括：

- **NAT1**：完全锥形 NAT
- **NAT2**：受限锥形 NAT
- **NAT3**：端口受限锥形 NAT
- **NAT4**：对称 NAT

### 连接方式

P3 支持多种连接方式：

- **直接连接**：两个节点可以直接通信
- **P2P 打洞**：通过 NAT 穿透技术实现点对点连接
- **中继连接**：当 P2P 打洞失败时，通过中继服务器转发流量

## 命令行界面

P3 客户端提供了丰富的命令行选项，方便您进行配置和管理。

### 基本命令

```bash
# 启动客户端
p3-client -config config.yaml

# 指定节点 ID 和令牌
p3-client -node my-node -token my-token

# 以守护进程模式运行
p3-client -d

# 安装为系统服务
p3-client -install

# 卸载系统服务
p3-client -uninstall
```

### 端口转发命令

```bash
# 添加端口转发规则
p3-client -add-forward -name ssh -protocol tcp -src-port 12222 -peer-node remote-node -dst-port 22 -dst-host localhost

# 删除端口转发规则
p3-client -remove-forward -name ssh

# 启用端口转发规则
p3-client -enable-forward -name ssh

# 禁用端口转发规则
p3-client -disable-forward -name ssh

# 列出所有端口转发规则
p3-client -list-forwards
```

### 网络命令

```bash
# 检测 NAT 类型
p3-client -detect-nat

# 测试与节点的连接
p3-client -test-connection -peer-node remote-node

# 显示网络状态
p3-client -network-status
```

### 其他命令

```bash
# 显示帮助信息
p3-client -help

# 显示版本信息
p3-client -version

# 显示详细日志
p3-client -verbose
```

## Web 管理界面

P3 提供了 Web 管理界面，方便您通过浏览器管理节点、应用和端口转发规则。

### 访问 Web 界面

1. 打开浏览器，访问 P3 服务器地址：`http://your-server-address:8080`
2. 使用您的用户名和密码登录

### 仪表盘

仪表盘显示系统概览，包括：

- 节点状态
- 应用状态
- 网络连接
- 流量统计

### 节点管理

在节点管理页面，您可以：

- 查看所有节点
- 添加新节点
- 编辑节点信息
- 删除节点
- 查看节点详情

### 应用管理

在应用管理页面，您可以：

- 查看所有应用
- 添加新应用
- 编辑应用配置
- 删除应用
- 启动/停止应用

### 端口转发管理

在端口转发管理页面，您可以：

- 查看所有端口转发规则
- 添加新规则
- 编辑规则配置
- 删除规则
- 启用/禁用规则

### 用户设置

在用户设置页面，您可以：

- 修改个人信息
- 更改密码
- 配置双因素认证
- 查看 API 令牌

## 使用场景

### 远程桌面连接

通过 P3，您可以轻松实现远程桌面连接：

1. 在远程计算机上运行 P3 客户端
2. 在本地计算机上添加端口转发规则：
   ```bash
   p3-client -add-forward -name rdp -protocol tcp -src-port 13389 -peer-node remote-pc -dst-port 3389 -dst-host localhost
   ```
3. 使用 RDP 客户端连接 `localhost:13389`

### SSH 连接

通过 P3，您可以轻松实现 SSH 连接：

1. 在远程服务器上运行 P3 客户端
2. 在本地计算机上添加端口转发规则：
   ```bash
   p3-client -add-forward -name ssh -protocol tcp -src-port 12222 -peer-node remote-server -dst-port 22 -dst-host localhost
   ```
3. 使用 SSH 客户端连接：
   ```bash
   ssh -p 12222 user@localhost
   ```

### Web 服务访问

通过 P3，您可以轻松访问远程 Web 服务：

1. 在远程服务器上运行 P3 客户端
2. 在本地计算机上添加端口转发规则：
   ```bash
   p3-client -add-forward -name web -protocol tcp -src-port 18080 -peer-node web-server -dst-port 80 -dst-host localhost
   ```
3. 在浏览器中访问 `http://localhost:18080`

### 游戏服务器

通过 P3，您可以轻松搭建游戏服务器：

1. 在游戏服务器上运行 P3 客户端
2. 在玩家计算机上添加端口转发规则：
   ```bash
   p3-client -add-forward -name minecraft -protocol tcp -src-port 25565 -peer-node game-server -dst-port 25565 -dst-host localhost
   ```
3. 在游戏中连接 `localhost:25565`

## 故障排除

### 连接问题

如果您遇到连接问题，请尝试以下步骤：

1. **检查网络连接**：确保您的设备已连接到互联网。
2. **检查服务器状态**：确保 P3 服务器正常运行。
3. **检查节点 ID 和令牌**：确保您使用了正确的节点 ID 和令牌。
4. **检查 NAT 类型**：使用 `-detect-nat` 命令检测您的 NAT 类型。
5. **尝试不同的连接方式**：如果 P2P 打洞失败，尝试使用中继连接。

### 端口转发问题

如果您遇到端口转发问题，请尝试以下步骤：

1. **检查端口是否被占用**：确保本地端口未被其他应用占用。
2. **检查防火墙设置**：确保防火墙未阻止相关端口。
3. **检查转发规则配置**：确保转发规则配置正确。
4. **检查目标主机和端口**：确保目标主机和端口可访问。

### 性能问题

如果您遇到性能问题，请尝试以下步骤：

1. **检查网络带宽**：确保您有足够的网络带宽。
2. **优化连接方式**：直接连接比中继连接更快。
3. **减少并发连接数**：过多的并发连接可能导致性能下降。
4. **检查系统资源**：确保您的设备有足够的 CPU 和内存资源。

## 常见问题

### P3 支持哪些操作系统？

P3 支持 Windows、macOS、Linux 和各种 Linux ARM 平台（如树莓派）。

### P3 如何保证安全性？

P3 使用 TLS1.3 + AES 双重加密，确保数据传输的安全性。此外，P3 还支持 TOTP 一次性密码授权，提供额外的安全保障。

### P3 是否支持 IPv6？

是的，P3 完全支持 IPv6 网络。

### P3 是否支持 UDP 协议？

是的，P3 支持 TCP 和 UDP 协议的端口转发。

### P3 是否支持移动平台？

目前 P3 主要支持桌面平台，但我们计划在未来支持 Android 和 iOS 平台。

### P3 是否支持多用户？

是的，P3 服务端支持多用户，每个用户可以管理多个节点和应用。

### P3 是否支持自定义中继服务器？

是的，您可以在配置文件中指定自定义的中继服务器。

### P3 是否支持端口范围转发？

目前 P3 支持单个端口的转发，我们计划在未来版本中支持端口范围转发。

### P3 是否支持流量限制？

是的，您可以在配置文件中设置带宽限制。

### P3 是否支持自动重连？

是的，P3 客户端会在连接断开时自动尝试重新连接。
