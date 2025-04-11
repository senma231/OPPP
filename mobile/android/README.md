# P3 Android 客户端

P3 Android 客户端是 P3 系统的移动端实现，提供端口转发和 P2P 连接功能。

## 功能特性

- 用户认证和设备管理
- 应用和端口转发管理
- P2P 连接和 NAT 穿透
- 网络状态监控
- 后台服务和推送通知
- 流量统计和设置管理

## 技术栈

- Kotlin 1.8+
- Jetpack Compose
- Coroutines & Flow
- Dagger Hilt
- Room Database
- Retrofit & OkHttp
- Go Mobile (gomobile)

## 开发环境设置

1. 安装 Android Studio Arctic Fox (2020.3.1) 或更高版本
2. 安装 Go 1.20+
3. 安装 gomobile:
   ```bash
   go install golang.org/x/mobile/cmd/gomobile@latest
   gomobile init
   ```
4. 克隆仓库:
   ```bash
   git clone https://github.com/senma231/OPPP.git
   cd OPPP/mobile/android
   ```
5. 在 Android Studio 中打开项目

## 构建说明

### 构建共享库

在构建 Android 应用之前，需要先构建 Go 共享库:

```bash
cd ../common
./build.sh android
```

这将在 `mobile/android/app/libs` 目录下生成 `p3.aar` 文件。

### 构建应用

在 Android Studio 中构建应用，或使用命令行:

```bash
./gradlew assembleDebug
```

生成的 APK 文件位于 `app/build/outputs/apk/debug/` 目录下。

## 项目结构

```
android/
├── app/                    # 应用模块
│   ├── src/
│   │   ├── main/
│   │   │   ├── java/       # Kotlin 代码
│   │   │   │   └── com/
│   │   │   │       └── p3/
│   │   │   │           ├── ui/           # UI 组件
│   │   │   │           ├── data/         # 数据层
│   │   │   │           ├── domain/       # 领域层
│   │   │   │           ├── service/      # 服务
│   │   │   │           └── util/         # 工具类
│   │   │   ├── res/        # 资源文件
│   │   │   └── AndroidManifest.xml
│   │   └── test/           # 测试代码
│   ├── build.gradle        # 应用构建脚本
│   └── libs/               # 本地库
├── build.gradle            # 项目构建脚本
└── gradle/                 # Gradle 配置
```

## 使用说明

1. 安装应用
2. 启动应用并登录
3. 添加设备和应用
4. 配置端口转发规则
5. 启动服务

详细使用说明请参阅应用内帮助或项目文档。

## 贡献指南

欢迎贡献代码、报告问题或提出改进建议。请参阅项目根目录的 [CONTRIBUTING.md](../../CONTRIBUTING.md) 了解更多信息。

## 许可证

MIT
