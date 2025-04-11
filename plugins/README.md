# P3 插件系统

P3 插件系统允许开发者扩展 P3 的功能，添加新的特性和集成。

## 插件类型

P3 支持以下类型的插件：

1. **服务端插件**：扩展服务端功能
2. **客户端插件**：扩展客户端功能
3. **Web 界面插件**：扩展 Web 管理界面

## 目录结构

```
plugins/
├── server/              # 服务端插件
├── client/              # 客户端插件
├── web/                 # Web 界面插件
└── examples/            # 插件示例
```

## 插件开发指南

### 服务端插件

服务端插件使用 Go 语言开发，通过实现 `ServerPlugin` 接口来扩展服务端功能。

```go
// ServerPlugin 服务端插件接口
type ServerPlugin interface {
    // Name 返回插件名称
    Name() string
    
    // Version 返回插件版本
    Version() string
    
    // Init 初始化插件
    Init(ctx context.Context, config map[string]interface{}) error
    
    // Start 启动插件
    Start() error
    
    // Stop 停止插件
    Stop() error
    
    // Handlers 返回插件的 HTTP 处理器
    Handlers() map[string]http.Handler
}
```

### 客户端插件

客户端插件使用 Go 语言开发，通过实现 `ClientPlugin` 接口来扩展客户端功能。

```go
// ClientPlugin 客户端插件接口
type ClientPlugin interface {
    // Name 返回插件名称
    Name() string
    
    // Version 返回插件版本
    Version() string
    
    // Init 初始化插件
    Init(ctx context.Context, config map[string]interface{}) error
    
    // Start 启动插件
    Start() error
    
    // Stop 停止插件
    Stop() error
    
    // OnEvent 处理事件
    OnEvent(event Event) error
}
```

### Web 界面插件

Web 界面插件使用 React 和 TypeScript 开发，通过实现 `WebPlugin` 接口来扩展 Web 管理界面。

```typescript
// WebPlugin Web 界面插件接口
interface WebPlugin {
    // name 插件名称
    name: string;
    
    // version 插件版本
    version: string;
    
    // init 初始化插件
    init(config: Record<string, any>): Promise<void>;
    
    // getRoutes 获取插件路由
    getRoutes(): React.ReactNode;
    
    // getMenuItems 获取插件菜单项
    getMenuItems(): MenuItem[];
    
    // getComponents 获取插件组件
    getComponents(): Record<string, React.ComponentType<any>>;
}
```

## 插件配置

插件配置在 P3 配置文件中指定：

```yaml
plugins:
  server:
    - name: "metrics"
      enabled: true
      config:
        port: 9090
        path: "/metrics"
    - name: "webhook"
      enabled: true
      config:
        url: "https://example.com/webhook"
        events: ["device.connected", "device.disconnected"]
  
  client:
    - name: "auto-reconnect"
      enabled: true
      config:
        max_retries: 5
        retry_interval: 10
    - name: "bandwidth-limiter"
      enabled: true
      config:
        upload_limit: 1024
        download_limit: 2048
  
  web:
    - name: "dashboard-widgets"
      enabled: true
      config:
        widgets: ["cpu", "memory", "network"]
    - name: "theme-customizer"
      enabled: true
      config:
        themes: ["light", "dark", "blue", "green"]
```

## 插件安装

### 手动安装

1. 下载插件包
2. 解压到相应的插件目录
3. 在配置文件中启用插件
4. 重启 P3

### 通过插件管理器安装

```bash
# 安装服务端插件
p3-server plugin install metrics

# 安装客户端插件
p3-client plugin install auto-reconnect

# 安装 Web 界面插件
p3-web plugin install dashboard-widgets
```

## 插件示例

### 服务端指标插件

```go
package metrics

import (
    "context"
    "net/http"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsPlugin 指标插件
type MetricsPlugin struct {
    config map[string]interface{}
    registry *prometheus.Registry
    server *http.Server
}

// Name 返回插件名称
func (p *MetricsPlugin) Name() string {
    return "metrics"
}

// Version 返回插件版本
func (p *MetricsPlugin) Version() string {
    return "1.0.0"
}

// Init 初始化插件
func (p *MetricsPlugin) Init(ctx context.Context, config map[string]interface{}) error {
    p.config = config
    p.registry = prometheus.NewRegistry()
    
    // 注册指标
    // ...
    
    return nil
}

// Start 启动插件
func (p *MetricsPlugin) Start() error {
    port := p.config["port"].(int)
    path := p.config["path"].(string)
    
    mux := http.NewServeMux()
    mux.Handle(path, promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}))
    
    p.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    
    go p.server.ListenAndServe()
    
    return nil
}

// Stop 停止插件
func (p *MetricsPlugin) Stop() error {
    if p.server != nil {
        return p.server.Shutdown(context.Background())
    }
    return nil
}

// Handlers 返回插件的 HTTP 处理器
func (p *MetricsPlugin) Handlers() map[string]http.Handler {
    return map[string]http.Handler{
        "/metrics": promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}),
    }
}
```

## 贡献插件

欢迎贡献插件！请遵循以下步骤：

1. Fork 仓库
2. 创建插件目录
3. 实现插件接口
4. 添加文档和示例
5. 提交 Pull Request

## 插件仓库

P3 插件仓库收集了社区贡献的插件：

- [服务端插件仓库](https://github.com/senma231/p3-server-plugins)
- [客户端插件仓库](https://github.com/senma231/p3-client-plugins)
- [Web 界面插件仓库](https://github.com/senma231/p3-web-plugins)

## 许可证

MIT
