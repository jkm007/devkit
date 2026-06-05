# DevKit - 后台管理系统

基于 Go + Vue3 的全栈后台管理系统，包含后端服务和前端管理界面。

## 项目结构

```
manager/
├── backend-server/     # Go 后端服务
├── frontend-admin/     # Vue3 前端管理界面
├── vue-vben-admin/     # Vue Vben Admin 框架源码
├── docs/               # 项目文档
└── tmp/                # 临时文件
```

## 技术栈

### 后端 (backend-server)

| 组件 | 选型 |
|------|------|
| 语言 | Go 1.21+ |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 缓存/队列 | Redis 7.x |
| 日志 | Zap + Lumberjack |
| 配置 | Viper |
| 认证 | JWT (Access + Refresh Token) |
| WebSocket | Gorilla WebSocket |
| 文件存储 | 本地 / MinIO / OSS / COS |
| 验证码 | go-captcha (滑块/拼图/旋转/点选/数字) |

### 前端 (frontend-admin)

| 组件 | 选型 |
|------|------|
| 框架 | Vue 3 |
| 构建 | Vite 8 |
| 语言 | TypeScript |
| UI | Ant Design Vue 4 |
| 状态 | Pinia |
| 样式 | Tailwind CSS 4 |
| 国际化 | vue-i18n |

## 快速启动

### 后端

```bash
cd backend-server

# 1. 启动依赖服务
docker-compose up -d mysql redis

# 2. 修改配置
cp config/config.yaml config/config.local.yaml

# 3. 运行
make run
```

后端服务地址：`http://localhost:8080`
Swagger 文档：`http://localhost:8080/swagger/index.html`

### 前端

```bash
cd frontend-admin

# 1. 安装依赖
pnpm install

# 2. 开发模式
pnpm dev
```

前端服务地址：`http://localhost:5666`

## 核心功能

### 验证码系统

支持多种验证码类型：

| 类型 | 说明 |
|------|------|
| slider | 滑块验证 |
| puzzle | 拼图验证 |
| rotation | 旋转验证 |
| point | 点选验证 |
| numeric | 数字验证码 |

- 配置化管理（类型、触发条件、长度等）
- 登录失败次数触发验证码
- 配置修改实时生效（无需重启）

### RBAC 权限模型

```
User ──(直接关联)──> Role ──> Permissions
  │
  └── Group ──(递归向上)──> GroupRole ──> Role
```

### Token 机制

- Access Token：短期有效，JWT 签名
- Refresh Token：长期有效，Redis 存储
- Token 轮换：刷新时自动生成新 Token

## 子项目文档

- [后端详细文档](./backend-server/README.md)
- [前端详细文档](./frontend-admin/README.md)

## 开发指南

### Git 提交规范

```
feat: 新功能
fix: 修复问题
docs: 文档变更
style: 代码格式
refactor: 重构
test: 测试
chore: 构建/工具
```

### 分支管理

- `main`: 主分支，稳定版本
- `develop`: 开发分支
- `feature/*`: 功能分支
- `hotfix/*`: 紧急修复分支

## License

MIT