# P3 iOS 客户端

P3 iOS 客户端是 P3 系统的移动端实现，提供端口转发和 P2P 连接功能。

## 功能特性

- 用户认证和设备管理
- 应用和端口转发管理
- P2P 连接和 NAT 穿透
- 网络状态监控
- 后台服务和推送通知
- 流量统计和设置管理

## 技术栈

- Swift 5.7+
- SwiftUI
- Combine
- Core Data
- URLSession
- Go Mobile (gomobile)

## 开发环境设置

1. 安装 Xcode 14+ 和 iOS SDK
2. 安装 Go 1.20+
3. 安装 gomobile:
   ```bash
   go install golang.org/x/mobile/cmd/gomobile@latest
   gomobile init
   ```
4. 克隆仓库:
   ```bash
   git clone https://github.com/senma231/OPPP.git
   cd OPPP/mobile/ios
   ```
5. 在 Xcode 中打开项目

## 构建说明

### 构建共享库

在构建 iOS 应用之前，需要先构建 Go 共享库:

```bash
cd ../common
./build.sh ios
```

这将在 `mobile/ios/Frameworks` 目录下生成 `P3.xcframework` 文件。

### 构建应用

在 Xcode 中构建应用，或使用命令行:

```bash
xcodebuild -scheme P3 -configuration Debug -sdk iphonesimulator
```

## 项目结构

```
ios/
├── P3/                     # 应用模块
│   ├── Views/              # SwiftUI 视图
│   ├── Models/             # 数据模型
│   ├── ViewModels/         # 视图模型
│   ├── Services/           # 服务
│   ├── Utils/              # 工具类
│   ├── Resources/          # 资源文件
│   └── Info.plist          # 应用配置
├── Frameworks/             # 框架目录
│   └── P3.xcframework     # Go 共享库
├── P3Tests/               # 测试代码
├── P3.xcodeproj/          # Xcode 项目
└── P3.xcworkspace/        # Xcode 工作区
```

## 使用说明

1. 安装应用
2. 启动应用并登录
3. 添加设备和应用
4. 配置端口转发规则
5. 启动服务

详细使用说明请参阅应用内帮助或项目文档。

## 注意事项

- iOS 应用需要适当的权限才能在后台运行网络服务
- 需要在 Info.plist 中配置网络权限
- 推送通知需要 Apple Developer 账户和证书

## 贡献指南

欢迎贡献代码、报告问题或提出改进建议。请参阅项目根目录的 [CONTRIBUTING.md](../../CONTRIBUTING.md) 了解更多信息。

## 许可证

MIT
