# P3 API 文档

本文档描述了 P3 服务端提供的 API 接口。

## 基本信息

- 基础 URL: `http://localhost:8080/api/v1`
- 认证方式: JWT Token
- 请求格式: JSON
- 响应格式: JSON

## 认证

### 登录

```
POST /auth/login
```

请求参数:

```json
{
  "username": "admin",
  "password": "password"
}
```

响应:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "1",
    "username": "admin",
    "email": "admin@example.com"
  }
}
```

### 注册

```
POST /auth/register
```

请求参数:

```json
{
  "username": "newuser",
  "password": "password",
  "email": "newuser@example.com"
}
```

响应:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "2",
    "username": "newuser",
    "email": "newuser@example.com"
  }
}
```

### 登出

```
POST /auth/logout
```

响应:

```json
{
  "message": "登出成功"
}
```

## 设备管理

### 获取设备列表

```
GET /devices
```

响应:

```json
{
  "devices": [
    {
      "id": "1",
      "name": "Office-PC",
      "status": "online",
      "natType": "Port Restricted Cone NAT",
      "externalIP": "203.0.113.1",
      "lastSeen": "2023-01-01T12:00:00Z"
    },
    {
      "id": "2",
      "name": "Home-PC",
      "status": "offline",
      "natType": "Symmetric NAT",
      "externalIP": "203.0.113.2",
      "lastSeen": "2023-01-01T10:00:00Z"
    }
  ]
}
```

### 获取设备详情

```
GET /devices/:id
```

响应:

```json
{
  "id": "1",
  "name": "Office-PC",
  "status": "online",
  "natType": "Port Restricted Cone NAT",
  "externalIP": "203.0.113.1",
  "lastSeen": "2023-01-01T12:00:00Z",
  "localIP": "192.168.1.2",
  "version": "0.1.0",
  "os": "windows",
  "arch": "amd64",
  "uptime": 86400,
  "connections": [
    {
      "peerID": "2",
      "peerName": "Home-PC",
      "type": "relay",
      "established": "2023-01-01T11:00:00Z",
      "bytesSent": 1024000,
      "bytesRecv": 2048000
    }
  ]
}
```

### 创建设备

```
POST /devices
```

请求参数:

```json
{
  "name": "New-PC",
  "token": "device-token"
}
```

响应:

```json
{
  "id": "3",
  "name": "New-PC",
  "token": "device-token"
}
```

### 更新设备

```
PUT /devices/:id
```

请求参数:

```json
{
  "name": "Updated-PC"
}
```

响应:

```json
{
  "id": "1",
  "name": "Updated-PC",
  "status": "online",
  "natType": "Port Restricted Cone NAT",
  "externalIP": "203.0.113.1",
  "lastSeen": "2023-01-01T12:00:00Z"
}
```

### 删除设备

```
DELETE /devices/:id
```

响应:

```json
{
  "message": "设备已删除"
}
```

## 应用管理

### 获取应用列表

```
GET /apps
```

响应:

```json
{
  "apps": [
    {
      "id": "1",
      "name": "Remote Desktop",
      "protocol": "tcp",
      "srcPort": 23389,
      "peerNode": "Office-PC",
      "dstPort": 3389,
      "dstHost": "localhost",
      "status": "running"
    },
    {
      "id": "2",
      "name": "SSH Server",
      "protocol": "tcp",
      "srcPort": 2222,
      "peerNode": "Office-PC",
      "dstPort": 22,
      "dstHost": "192.168.1.5",
      "status": "stopped"
    }
  ]
}
```

### 获取应用详情

```
GET /apps/:id
```

响应:

```json
{
  "id": "1",
  "name": "Remote Desktop",
  "protocol": "tcp",
  "srcPort": 23389,
  "peerNode": "Office-PC",
  "dstPort": 3389,
  "dstHost": "localhost",
  "status": "running",
  "createdAt": "2023-01-01T10:00:00Z",
  "updatedAt": "2023-01-01T11:00:00Z",
  "stats": {
    "bytesSent": 1024000,
    "bytesReceived": 2048000,
    "connections": 5,
    "startTime": "2023-01-01T11:00:00Z"
  }
}
```

### 创建应用

```
POST /apps
```

请求参数:

```json
{
  "name": "New App",
  "protocol": "tcp",
  "srcPort": 8080,
  "peerNode": "Office-PC",
  "dstPort": 8080,
  "dstHost": "localhost"
}
```

响应:

```json
{
  "id": "3",
  "name": "New App",
  "protocol": "tcp",
  "srcPort": 8080,
  "peerNode": "Office-PC",
  "dstPort": 8080,
  "dstHost": "localhost",
  "status": "stopped"
}
```

### 更新应用

```
PUT /apps/:id
```

请求参数:

```json
{
  "name": "Updated App",
  "srcPort": 8081
}
```

响应:

```json
{
  "id": "1",
  "name": "Updated App",
  "protocol": "tcp",
  "srcPort": 8081,
  "peerNode": "Office-PC",
  "dstPort": 3389,
  "dstHost": "localhost",
  "status": "running"
}
```

### 删除应用

```
DELETE /apps/:id
```

响应:

```json
{
  "message": "应用已删除"
}
```

### 启动应用

```
POST /apps/:id/start
```

响应:

```json
{
  "id": "1",
  "name": "Remote Desktop",
  "status": "running"
}
```

### 停止应用

```
POST /apps/:id/stop
```

响应:

```json
{
  "id": "1",
  "name": "Remote Desktop",
  "status": "stopped"
}
```

## 端口转发管理

### 获取转发规则列表

```
GET /forwards
```

响应:

```json
{
  "forwards": [
    {
      "id": "1",
      "protocol": "tcp",
      "srcPort": 23389,
      "dstHost": "localhost",
      "dstPort": 3389,
      "description": "Remote Desktop",
      "enabled": true,
      "stats": {
        "bytesSent": 1024000,
        "bytesReceived": 2048000,
        "connections": 5,
        "startTime": "2023-01-01T11:00:00Z"
      }
    },
    {
      "id": "2",
      "protocol": "tcp",
      "srcPort": 2222,
      "dstHost": "192.168.1.5",
      "dstPort": 22,
      "description": "SSH Server",
      "enabled": false,
      "stats": {
        "bytesSent": 0,
        "bytesReceived": 0,
        "connections": 0,
        "startTime": "2023-01-01T11:00:00Z"
      }
    }
  ]
}
```

### 获取转发规则详情

```
GET /forwards/:id
```

响应:

```json
{
  "id": "1",
  "protocol": "tcp",
  "srcPort": 23389,
  "dstHost": "localhost",
  "dstPort": 3389,
  "description": "Remote Desktop",
  "enabled": true,
  "createdAt": "2023-01-01T10:00:00Z",
  "updatedAt": "2023-01-01T11:00:00Z",
  "stats": {
    "bytesSent": 1024000,
    "bytesReceived": 2048000,
    "connections": 5,
    "startTime": "2023-01-01T11:00:00Z"
  }
}
```

### 创建转发规则

```
POST /forwards
```

请求参数:

```json
{
  "protocol": "tcp",
  "srcPort": 8080,
  "dstHost": "localhost",
  "dstPort": 8080,
  "description": "Web Server"
}
```

响应:

```json
{
  "id": "3",
  "protocol": "tcp",
  "srcPort": 8080,
  "dstHost": "localhost",
  "dstPort": 8080,
  "description": "Web Server",
  "enabled": false
}
```

### 更新转发规则

```
PUT /forwards/:id
```

请求参数:

```json
{
  "description": "Updated Web Server",
  "srcPort": 8081
}
```

响应:

```json
{
  "id": "1",
  "protocol": "tcp",
  "srcPort": 8081,
  "dstHost": "localhost",
  "dstPort": 3389,
  "description": "Updated Web Server",
  "enabled": true
}
```

### 删除转发规则

```
DELETE /forwards/:id
```

响应:

```json
{
  "message": "转发规则已删除"
}
```

### 启用转发规则

```
POST /forwards/:id/enable
```

响应:

```json
{
  "id": "1",
  "description": "Remote Desktop",
  "enabled": true
}
```

### 禁用转发规则

```
POST /forwards/:id/disable
```

响应:

```json
{
  "id": "1",
  "description": "Remote Desktop",
  "enabled": false
}
```

## 系统状态

### 获取系统状态

```
GET /status
```

响应:

```json
{
  "version": "0.1.0",
  "uptime": 86400,
  "devices": {
    "total": 2,
    "online": 1
  },
  "apps": {
    "total": 2,
    "running": 1
  },
  "forwards": {
    "total": 2,
    "enabled": 1
  },
  "connections": {
    "direct": 0,
    "upnp": 0,
    "holePunch": 1,
    "relay": 1
  },
  "traffic": {
    "sent": 10240000,
    "received": 20480000
  }
}
```
