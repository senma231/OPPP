# P3 Web 管理界面

P3 Web 管理界面是 P3 系统的 Web 前端，提供用户友好的界面来管理设备、应用和端口转发规则。

## 功能特性

- 用户认证和权限管理
- 设备管理和监控
- 应用和端口转发管理
- 网络状态和连接监控
- 流量统计和可视化
- 系统设置和配置

## 技术栈

- React 18+
- TypeScript 4.9+
- Ant Design 5+
- Redux Toolkit
- React Router 6+
- Axios
- ECharts
- Vite

## 开发环境设置

1. 安装 Node.js 16+ 和 npm 8+
2. 克隆仓库:
   ```bash
   git clone https://github.com/senma231/OPPP.git
   cd OPPP/web
   ```
3. 安装依赖:
   ```bash
   npm install
   ```
4. 启动开发服务器:
   ```bash
   npm run dev
   ```

## 构建说明

```bash
# 开发构建
npm run dev

# 生产构建
npm run build

# 预览生产构建
npm run preview

# 代码检查
npm run lint

# 运行测试
npm run test
```

## 项目结构

```
web/
├── public/              # 静态资源
├── src/
│   ├── api/             # API 接口
│   ├── assets/          # 资源文件
│   ├── components/      # 组件
│   ├── hooks/           # 自定义 Hooks
│   ├── layouts/         # 布局组件
│   ├── pages/           # 页面组件
│   ├── store/           # Redux 状态管理
│   ├── types/           # TypeScript 类型定义
│   ├── utils/           # 工具函数
│   ├── App.tsx          # 应用入口
│   ├── main.tsx         # 主入口
│   └── vite-env.d.ts    # Vite 环境定义
├── .eslintrc.js         # ESLint 配置
├── .prettierrc          # Prettier 配置
├── index.html           # HTML 模板
├── package.json         # 项目配置
├── tsconfig.json        # TypeScript 配置
└── vite.config.ts       # Vite 配置
```

## 页面和功能

### 仪表盘

- 系统概览
- 设备状态
- 应用状态
- 网络连接
- 流量统计

### 设备管理

- 设备列表
- 设备详情
- 添加设备
- 编辑设备
- 删除设备

### 应用管理

- 应用列表
- 应用详情
- 添加应用
- 编辑应用
- 删除应用
- 启动/停止应用

### 端口转发

- 转发规则列表
- 添加规则
- 编辑规则
- 删除规则
- 启用/禁用规则

### 网络

- 网络状态
- NAT 类型
- 连接测试
- 中继服务器

### 用户管理

- 用户列表
- 添加用户
- 编辑用户
- 删除用户
- 权限管理

### 设置

- 系统设置
- 安全设置
- 网络设置
- 日志设置
- 备份和恢复

## API 集成

Web 界面通过 RESTful API 与 P3 服务端通信。主要 API 端点包括：

- `/api/v1/auth/*` - 认证相关
- `/api/v1/devices/*` - 设备管理
- `/api/v1/apps/*` - 应用管理
- `/api/v1/users/*` - 用户管理
- `/api/v1/stats/*` - 统计信息

详细 API 文档请参阅 [API.md](../docs/API.md)。

## 主题和定制

Web 界面支持亮色和暗色主题，并可以通过 Ant Design 的主题定制功能进行个性化设置。

## 国际化

Web 界面支持多语言，包括：

- 简体中文
- 英文
- 日文
- 韩文

## 贡献指南

欢迎贡献代码、报告问题或提出改进建议。请参阅项目根目录的 [CONTRIBUTING.md](../CONTRIBUTING.md) 了解更多信息。

## 许可证

MIT
