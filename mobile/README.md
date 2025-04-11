# P3 移动平台支持

本目录包含 P3 移动平台（Android 和 iOS）的实现代码。

## 目录结构

```
mobile/
├── android/        # Android 平台代码
├── ios/            # iOS 平台代码
├── common/         # 跨平台共享代码
└── api/            # 移动平台 API 接口
```

## 技术栈

- **Android**: Kotlin, Jetpack Compose
- **iOS**: Swift, SwiftUI
- **共享代码**: Go Mobile (gomobile)

## 开发指南

### 环境设置

#### Android 开发环境

1. 安装 Android Studio
2. 安装 Go 1.20+
3. 安装 gomobile:
   ```bash
   go install golang.org/x/mobile/cmd/gomobile@latest
   gomobile init
   ```

#### iOS 开发环境

1. 安装 Xcode
2. 安装 Go 1.20+
3. 安装 gomobile:
   ```bash
   go install golang.org/x/mobile/cmd/gomobile@latest
   gomobile init
   ```

### 构建共享库

```bash
cd mobile/common
./build.sh
```

这将生成 Android 和 iOS 平台的共享库。

### 构建 Android 应用

```bash
cd mobile/android
./gradlew assembleDebug
```

### 构建 iOS 应用

```bash
cd mobile/ios
xcodebuild -scheme P3 -configuration Debug
```

## 功能列表

- [ ] 用户认证
- [ ] 设备管理
- [ ] 应用管理
- [ ] 端口转发
- [ ] P2P 连接
- [ ] 网络状态监控
- [ ] 推送通知
- [ ] 后台服务
- [ ] 流量统计
- [ ] 设置管理

## 路线图

1. **阶段一**：基础功能实现
   - 用户认证
   - 设备管理
   - 应用管理
   - 基本设置

2. **阶段二**：核心功能实现
   - 端口转发
   - P2P 连接
   - 网络状态监控
   - 后台服务

3. **阶段三**：高级功能实现
   - 推送通知
   - 流量统计
   - 高级设置
   - 性能优化

## 贡献指南

欢迎贡献代码、报告问题或提出改进建议。请参阅项目根目录的 [CONTRIBUTING.md](../CONTRIBUTING.md) 了解更多信息。
