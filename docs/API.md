# P3 API 文档

本文档详细描述了 P3 服务端提供的 API 接口。所有 API 都使用 JSON 格式进行数据交换，并使用 JWT 进行认证。

## 基本信息

- **基础 URL**: `http://your-server-address:8080/api/v1`
- **认证方式**: Bearer Token (JWT)
- **内容类型**: `application/json`

## 认证

### 登录

获取访问令牌和刷新令牌。

**请求**:

```
POST /auth/login
```

**请求体**:

```json
{
  "username": "your-username",
  "password": "your-password"
}
```

**响应**:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### 刷新令牌

使用刷新令牌获取新的访问令牌。

**请求**:

```
POST /auth/refresh
```

**请求体**:

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应**:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### 注册

创建新用户。

**请求**:

```
POST /auth/register
```

**请求体**:

```json
{
  "username": "new-username",
  "password": "new-password",
  "email": "user@example.com"
}
```

**响应**:

```json
{
  "id": 123,
  "username": "new-username",
  "email": "user@example.com",
  "created_at": "2023-06-01T12:00:00Z"
}
```

### 注销

使当前令牌失效。

**请求**:

```
POST /auth/logout
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "message": "已成功注销"
}
```

## 设备管理

### 获取设备列表

获取当前用户的所有设备。

**请求**:

```
GET /devices
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "devices": [
    {
      "id": "device-1",
      "name": "My PC",
      "status": "online",
      "ip": "192.168.1.100",
      "nat_type": "NAT2",
      "last_seen": "2023-06-01T12:00:00Z"
    },
    {
      "id": "device-2",
      "name": "My Server",
      "status": "offline",
      "ip": "10.0.0.1",
      "nat_type": "NAT1",
      "last_seen": "2023-05-31T10:00:00Z"
    }
  ]
}
```

### 获取设备详情

获取特定设备的详细信息。

**请求**:

```
GET /devices/{device_id}
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "id": "device-1",
  "name": "My PC",
  "status": "online",
  "ip": "192.168.1.100",
  "external_ip": "203.0.113.1",
  "nat_type": "NAT2",
  "upnp_available": true,
  "version": "1.0.0",
  "os": "windows",
  "arch": "amd64",
  "cpu_usage": 25.5,
  "memory_usage": 40.2,
  "bandwidth_up": 1024,
  "bandwidth_down": 2048,
  "last_seen": "2023-06-01T12:00:00Z",
  "created_at": "2023-01-01T00:00:00Z"
}
```

### 添加设备

添加新设备。

**请求**:

```
POST /devices
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**请求体**:

```json
{
  "name": "New Device",
  "description": "My new device"
}
```

**响应**:

```json
{
  "id": "device-3",
  "name": "New Device",
  "description": "My new device",
  "token": "device-token-123",
  "created_at": "2023-06-01T12:00:00Z"
}
```

### 更新设备

更新设备信息。

**请求**:

```
PUT /devices/{device_id}
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**请求体**:

```json
{
  "name": "Updated Device Name",
  "description": "Updated description"
}
```

**响应**:

```json
{
  "id": "device-1",
  "name": "Updated Device Name",
  "description": "Updated description",
  "updated_at": "2023-06-01T12:30:00Z"
}
```

### 删除设备

删除设备。

**请求**:

```
DELETE /devices/{device_id}
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "message": "设备已成功删除"
}
```

## 应用管理

### 获取应用列表

获取当前用户的所有应用。

**请求**:

```
GET /apps
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "apps": [
    {
      "id": "app-1",
      "name": "SSH",
      "protocol": "tcp",
      "src_port": 12222,
      "dst_port": 22,
      "dst_host": "localhost",
      "status": "running"
    },
    {
      "id": "app-2",
      "name": "RDP",
      "protocol": "tcp",
      "src_port": 13389,
      "dst_port": 3389,
      "dst_host": "localhost",
      "status": "stopped"
    }
  ]
}
```

### 获取应用详情

获取特定应用的详细信息。

**请求**:

```
GET /apps/{app_id}
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "id": "app-1",
  "name": "SSH",
  "description": "SSH access to my server",
  "protocol": "tcp",
  "src_port": 12222,
  "dst_port": 22,
  "dst_host": "localhost",
  "peer_device": "device-2",
  "status": "running",
  "auto_start": true,
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-06-01T12:00:00Z",
  "stats": {
    "connections": 10,
    "bytes_sent": 1024000,
    "bytes_received": 2048000,
    "uptime": 3600
  }
}
```

### 添加应用

添加新应用。

**请求**:

```
POST /apps
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**请求体**:

```json
{
  "name": "Web Server",
  "description": "Access to my web server",
  "protocol": "tcp",
  "src_port": 18080,
  "dst_port": 80,
  "dst_host": "localhost",
  "peer_device": "device-2",
  "auto_start": true
}
```

**响应**:

```json
{
  "id": "app-3",
  "name": "Web Server",
  "description": "Access to my web server",
  "protocol": "tcp",
  "src_port": 18080,
  "dst_port": 80,
  "dst_host": "localhost",
  "peer_device": "device-2",
  "status": "stopped",
  "auto_start": true,
  "created_at": "2023-06-01T12:00:00Z"
}
```

### 更新应用

更新应用信息。

**请求**:

```
PUT /apps/{app_id}
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**请求体**:

```json
{
  "name": "Updated Web Server",
  "description": "Updated description",
  "dst_port": 8080,
  "auto_start": false
}
```

**响应**:

```json
{
  "id": "app-3",
  "name": "Updated Web Server",
  "description": "Updated description",
  "protocol": "tcp",
  "src_port": 18080,
  "dst_port": 8080,
  "dst_host": "localhost",
  "peer_device": "device-2",
  "status": "stopped",
  "auto_start": false,
  "updated_at": "2023-06-01T12:30:00Z"
}
```

### 删除应用

删除应用。

**请求**:

```
DELETE /apps/{app_id}
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "message": "应用已成功删除"
}
```

### 启动应用

启动应用。

**请求**:

```
POST /apps/{app_id}/start
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "id": "app-3",
  "name": "Web Server",
  "status": "running",
  "started_at": "2023-06-01T12:30:00Z"
}
```

### 停止应用

停止应用。

**请求**:

```
POST /apps/{app_id}/stop
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "id": "app-3",
  "name": "Web Server",
  "status": "stopped",
  "stopped_at": "2023-06-01T12:35:00Z"
}
```

## 用户管理

### 获取当前用户信息

获取当前登录用户的信息。

**请求**:

```
GET /users/me
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "id": 123,
  "username": "your-username",
  "email": "user@example.com",
  "role": "user",
  "created_at": "2023-01-01T00:00:00Z",
  "last_login": "2023-06-01T12:00:00Z"
}
```

### 更新用户信息

更新当前用户的信息。

**请求**:

```
PUT /users/me
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**请求体**:

```json
{
  "email": "new-email@example.com",
  "password": "new-password"
}
```

**响应**:

```json
{
  "id": 123,
  "username": "your-username",
  "email": "new-email@example.com",
  "updated_at": "2023-06-01T12:30:00Z"
}
```

### 启用双因素认证

为当前用户启用双因素认证。

**请求**:

```
POST /users/me/2fa/enable
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code_url": "otpauth://totp/P3:your-username?secret=JBSWY3DPEHPK3PXP&issuer=P3"
}
```

### 验证双因素认证

验证双因素认证代码。

**请求**:

```
POST /users/me/2fa/verify
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**请求体**:

```json
{
  "code": "123456"
}
```

**响应**:

```json
{
  "enabled": true,
  "message": "双因素认证已成功启用"
}
```

### 禁用双因素认证

为当前用户禁用双因素认证。

**请求**:

```
POST /users/me/2fa/disable
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**请求体**:

```json
{
  "code": "123456"
}
```

**响应**:

```json
{
  "enabled": false,
  "message": "双因素认证已成功禁用"
}
```

## 统计信息

### 获取系统统计信息

获取系统统计信息。

**请求**:

```
GET /stats/system
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "users_count": 100,
  "devices_count": 250,
  "apps_count": 500,
  "online_devices": 150,
  "total_connections": 1000,
  "total_traffic": 1073741824,
  "cpu_usage": 25.5,
  "memory_usage": 40.2,
  "uptime": 604800
}
```

### 获取用户统计信息

获取当前用户的统计信息。

**请求**:

```
GET /stats/user
```

**请求头**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应**:

```json
{
  "devices_count": 5,
  "apps_count": 10,
  "online_devices": 3,
  "total_connections": 100,
  "total_traffic": 104857600,
  "active_connections": 5
}
```

## 错误响应

所有 API 错误都使用标准格式返回：

```json
{
  "code": 1001,
  "message": "错误消息"
}
```

### 常见错误码

| 错误码 | 描述 |
|-------|------|
| 1000 | 未知错误 |
| 1001 | 无效参数 |
| 1002 | 未授权 |
| 1003 | 禁止访问 |
| 1004 | 未找到 |
| 1005 | 冲突 |
| 1006 | 内部错误 |
| 1007 | 数据库错误 |
| 1008 | 网络错误 |
| 1009 | 超时 |
| 1010 | 未实现 |
| 1011 | 服务不可用 |
| 1012 | 请求过多 |
| 1013 | 网关错误 |
| 1014 | 网关超时 |
| 1015 | 无效令牌 |
| 1016 | 令牌过期 |
