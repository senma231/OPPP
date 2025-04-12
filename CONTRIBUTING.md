# 贡献指南

感谢您对 P3 项目的关注！我们欢迎各种形式的贡献，包括但不限于代码贡献、文档改进、问题报告和功能建议。

## 目录

- [行为准则](#行为准则)
- [如何贡献](#如何贡献)
  - [报告问题](#报告问题)
  - [提交功能请求](#提交功能请求)
  - [提交代码](#提交代码)
- [开发流程](#开发流程)
  - [分支策略](#分支策略)
  - [提交信息规范](#提交信息规范)
  - [代码风格](#代码风格)
- [测试](#测试)
- [文档](#文档)
- [发布流程](#发布流程)

## 行为准则

本项目采用 [Contributor Covenant](https://www.contributor-covenant.org/) 行为准则。参与本项目即表示您同意遵守其条款。

## 如何贡献

### 报告问题

如果您发现了 bug 或有改进建议，请在 GitHub 上提交 issue。提交 issue 时，请尽可能详细地描述问题，包括：

1. 问题描述
2. 复现步骤
3. 预期行为
4. 实际行为
5. 环境信息（操作系统、Go 版本等）
6. 相关日志或截图

### 提交功能请求

如果您有新功能的想法，请在 GitHub 上提交 issue，并标记为 "enhancement"。在功能请求中，请包含：

1. 功能描述
2. 使用场景
3. 实现思路（如果有）

### 提交代码

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交您的更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建一个 Pull Request

## 开发流程

### 分支策略

- `main`: 主分支，包含最新的稳定代码
- `develop`: 开发分支，包含最新的开发代码
- `feature/*`: 特性分支，用于开发新功能
- `bugfix/*`: 修复分支，用于修复 bug
- `release/*`: 发布分支，用于准备发布
- `hotfix/*`: 热修复分支，用于紧急修复生产环境的问题

### 提交信息规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范来格式化提交信息。提交信息应该遵循以下格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

其中：

- `type`: 提交类型，如 `feat`、`fix`、`docs`、`style`、`refactor`、`perf`、`test`、`chore` 等
- `scope`: 可选，表示影响范围，如 `server`、`client`、`api` 等
- `subject`: 简短描述
- `body`: 可选，详细描述
- `footer`: 可选，包含关闭的 issue 等信息

示例：

```
feat(server): 添加用户认证功能

实现了基于 JWT 的用户认证功能，包括登录、注册和令牌刷新。

Closes #123
```

### 代码风格

我们使用 Go 官方的代码风格和 `gofmt` 工具来格式化代码。在提交代码前，请确保您的代码已经通过 `gofmt` 格式化。

此外，我们还使用以下工具来保证代码质量：

- `go vet`: 检查代码中的常见错误
- `golint`: 检查代码风格
- `goimports`: 管理导入语句

## 测试

所有新代码都应该包含适当的测试。我们使用 Go 的标准测试框架和 `go test` 命令来运行测试。

测试分为以下几类：

- 单元测试：测试单个函数或方法
- 集成测试：测试多个组件的交互
- 性能测试：测试代码的性能
- 安全测试：测试代码的安全性

在提交代码前，请确保所有测试都能通过：

```bash
make test
```

## 文档

文档是项目的重要组成部分。如果您添加了新功能或修改了现有功能，请确保更新相应的文档。

文档包括：

- 代码注释
- README 和其他 Markdown 文档
- API 文档
- 用户手册

## 发布流程

1. 从 `develop` 分支创建 `release/vX.Y.Z` 分支
2. 在 release 分支上进行最后的测试和修复
3. 更新版本号和 CHANGELOG.md
4. 将 release 分支合并到 `main` 分支
5. 在 `main` 分支上创建标签 `vX.Y.Z`
6. 将 `main` 分支合并回 `develop` 分支
7. 运行发布脚本创建发布包和 Docker 镜像
8. 在 GitHub 上创建 release，上传发布包

感谢您的贡献！
